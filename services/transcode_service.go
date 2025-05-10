package services

import (
	"fmt"
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"sync"
	"video-streaming/models"
)

const (
	VideoDir = "./videos"
)

type TranscodeService struct {
	BaseDir   string
	Qualities []Quality
}

type Quality struct {
	Name       string
	Resolution string
	Bitrate    string
}

func NewTranscodeService(baseDir string) *TranscodeService {
	return &TranscodeService{
		BaseDir: baseDir,
		Qualities: []Quality{
			{Name: "1080p", Resolution: "1920x1080", Bitrate: "4000k"},
			{Name: "720p", Resolution: "1280x720", Bitrate: "2500k"},
			{Name: "480p", Resolution: "854x480", Bitrate: "1000k"},
		},
	}
}

func (s *TranscodeService) TranscodeVideo(uploadID string) error {
	inputPath := filepath.Join(s.BaseDir, uploadID, "original.mp4")

	// 检查输入文件
	if _, err := os.Stat(inputPath); os.IsNotExist(err) {
		return fmt.Errorf("input video not found at %s: %v", inputPath, err)
	}

	// 检查文件大小
	fileInfo, err := os.Stat(inputPath)
	if err != nil {
		return fmt.Errorf("failed to get file info: %v", err)
	}
	if fileInfo.Size() == 0 {
		return fmt.Errorf("input file is empty")
	}

	log.Printf("Starting transcoding for upload %s, input file size: %d bytes", uploadID, fileInfo.Size())

	// 验证输入文件
	if err := s.verifyInputFile(inputPath); err != nil {
		return fmt.Errorf("input file verification failed: %v", err)
	}

	var wg sync.WaitGroup
	errors := make(chan error, len(s.Qualities))

	for _, quality := range s.Qualities {
		wg.Add(1)
		go func(q Quality) {
			defer wg.Done()
			if err := s.transcodeToQuality(inputPath, uploadID, q); err != nil {
				errors <- fmt.Errorf("failed to transcode to %s: %v", q.Name, err)
			}
		}(quality)
	}

	// 等待所有转码完成
	wg.Wait()
	close(errors)

	// 检查是否有错误发生
	for err := range errors {
		if err != nil {
			return err
		}
	}

	return nil
}

func (s *TranscodeService) transcodeToQuality(inputPath, uploadID string, quality Quality) error {
	outputPath := filepath.Join(s.BaseDir, uploadID, fmt.Sprintf("%s.mp4", quality.Name))

	// 修改 FFmpeg 命令参数，添加更多参数确保生成正确的 MP4 文件
	args := []string{
		"-i", inputPath,
		"-c:v", "libx264",
		"-preset", "medium",
		"-profile:v", "main",
		"-level", "4.0",
		"-crf", "23",
		"-s", quality.Resolution,
		"-b:v", quality.Bitrate,
		"-c:a", "aac",
		"-b:a", "128k",
		"-movflags", "+faststart", // 确保 moov atom 在文件开头
		"-y",        // 覆盖已存在的文件
		"-f", "mp4", // 强制输出格式为 MP4
		"-strict", "experimental", // 允许实验性编码器
		outputPath,
	}

	cmd := exec.Command("ffmpeg", args...)

	// 捕获输出
	output, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("ffmpeg error: %v\nOutput: %s", err, string(output))
	}

	// 验证输出文件
	if err := s.verifyOutputFile(outputPath); err != nil {
		return fmt.Errorf("output file verification failed: %v", err)
	}

	return nil
}

// 添加新方法：验证输出文件
func (s *TranscodeService) verifyOutputFile(filePath string) error {
	// 检查文件是否存在
	if _, err := os.Stat(filePath); os.IsNotExist(err) {
		return fmt.Errorf("output file does not exist: %s", filePath)
	}

	// 使用 ffprobe 检查文件
	cmd := exec.Command("ffprobe",
		"-v", "error",
		"-select_streams", "v:0",
		"-show_entries", "stream=codec_name,width,height",
		"-of", "json",
		filePath,
	)

	output, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("ffprobe error: %v\nOutput: %s", err, string(output))
	}

	// 检查文件大小
	fileInfo, err := os.Stat(filePath)
	if err != nil {
		return fmt.Errorf("failed to get file info: %v", err)
	}

	if fileInfo.Size() == 0 {
		return fmt.Errorf("output file is empty")
	}

	return nil
}

// 添加新方法：验证输入文件
func (s *TranscodeService) verifyInputFile(filePath string) error {
	cmd := exec.Command("ffprobe",
		"-v", "error",
		"-select_streams", "v:0",
		"-show_entries", "stream=codec_name,width,height",
		"-of", "json",
		filePath,
	)

	output, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("ffprobe error: %v\nOutput: %s", err, string(output))
	}

	return nil
}

func TranscodeVideo(videoID string, inputPath string) error {
	// 创建视频目录
	videoDir := filepath.Join(VideoDir, videoID)
	if err := os.MkdirAll(videoDir, 0755); err != nil {
		return fmt.Errorf("failed to create video directory: %v", err)
	}

	// 定义转码配置
	qualities := []struct {
		resolution string
		bitrate    string
	}{
		{"720p", "2000k"},
		{"1080p", "4000k"},
	}

	// 执行转码
	for _, quality := range qualities {
		outputPath := filepath.Join(videoDir, quality.resolution+".mp4")

		// 构建 ffmpeg 命令
		cmd := exec.Command("ffmpeg",
			"-i", inputPath,
			"-c:v", "libx264",
			"-b:v", quality.bitrate,
			"-c:a", "aac",
			"-b:a", "128k",
			"-vf", fmt.Sprintf("scale=-2:%s", quality.resolution),
			"-y",
			outputPath,
		)

		// 执行命令
		if err := cmd.Run(); err != nil {
			return fmt.Errorf("failed to transcode to %s: %v", quality.resolution, err)
		}

		// 获取文件大小
		fileInfo, err := os.Stat(outputPath)
		if err != nil {
			return fmt.Errorf("failed to get file info: %v", err)
		}

		// 创建质量记录
		quality := &models.Quality{
			VideoID:    videoID,
			Resolution: quality.resolution,
			Path:       fmt.Sprintf("/api/videos/%s/stream?quality=%s", videoID, quality.resolution),
			Size:       fileInfo.Size(),
		}

		if err := models.CreateQuality(quality); err != nil {
			return fmt.Errorf("failed to create quality record: %v", err)
		}
	}

	// 删除临时文件
	if err := os.Remove(inputPath); err != nil {
		fmt.Printf("Warning: failed to remove temp file: %v\n", err)
	}

	return nil
}
