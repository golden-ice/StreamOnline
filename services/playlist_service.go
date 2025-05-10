package services

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
)

type PlaylistService struct {
	BaseDir string
}

func NewPlaylistService(baseDir string) *PlaylistService {
	return &PlaylistService{
		BaseDir: baseDir,
	}
}

func (s *PlaylistService) GenerateHLSPlaylist(videoID string) error {
	videoPath := filepath.Join(s.BaseDir, videoID, "original.mp4")
	outputDir := filepath.Join(s.BaseDir, videoID, "hls")

	// 创建输出目录
	if err := os.MkdirAll(outputDir, 0755); err != nil {
		return fmt.Errorf("failed to create output directory: %v", err)
	}

	// 生成不同质量的HLS流
	qualities := []struct {
		name       string
		resolution string
		bitrate    string
	}{
		{"1080p", "1920x1080", "4000k"},
		{"720p", "1280x720", "2500k"},
		{"480p", "854x480", "1000k"},
	}

	// 生成主播放列表
	masterPlaylist := "#EXTM3U\n"
	masterPlaylist += "#EXT-X-VERSION:3\n"

	for _, quality := range qualities {
		// 为每个质量生成单独的播放列表
		playlistPath := filepath.Join(outputDir, quality.name+".m3u8")
		segmentPath := filepath.Join(outputDir, quality.name)

		// 使用FFmpeg生成HLS流
		cmd := exec.Command("ffmpeg",
			"-i", videoPath,
			"-c:v", "libx264",
			"-c:a", "aac",
			"-b:v", quality.bitrate,
			"-s", quality.resolution,
			"-f", "hls",
			"-hls_time", "10",
			"-hls_list_size", "0",
			"-hls_segment_filename", filepath.Join(segmentPath, "segment_%03d.ts"),
			playlistPath,
		)

		if err := cmd.Run(); err != nil {
			return fmt.Errorf("failed to generate HLS stream for %s: %v", quality.name, err)
		}

		// 添加到主播放列表
		masterPlaylist += fmt.Sprintf("#EXT-X-STREAM-INF:BANDWIDTH=%s,RESOLUTION=%s\n",
			quality.bitrate, quality.resolution)
		masterPlaylist += fmt.Sprintf("%s.m3u8\n", quality.name)
	}

	// 保存主播放列表
	masterPlaylistPath := filepath.Join(outputDir, "master.m3u8")
	if err := os.WriteFile(masterPlaylistPath, []byte(masterPlaylist), 0644); err != nil {
		return fmt.Errorf("failed to write master playlist: %v", err)
	}

	return nil
}
