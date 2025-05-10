package handlers

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
)

// VerifyFile 验证文件是否完整
func VerifyFile(filePath string) error {
	// 检查文件是否存在
	if _, err := os.Stat(filePath); os.IsNotExist(err) {
		return fmt.Errorf("file does not exist")
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

	return nil
}

// VerifyTranscodedFiles 验证转码后的文件
func VerifyTranscodedFiles(videoID string) error {
	qualities := []string{"1080p", "720p", "480p"}
	for _, quality := range qualities {
		filePath := filepath.Join("./videos", videoID, quality+".mp4")
		if err := VerifyFile(filePath); err != nil {
			return fmt.Errorf("quality %s verification failed: %v", quality, err)
		}
	}
	return nil
}
