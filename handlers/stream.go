package handlers

import (
	"fmt"
	"log"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"video-streaming/models"

	"github.com/gin-gonic/gin"
)

const (
	VideoDir = "./videos"
)

func StreamVideo(c *gin.Context) {
	videoID := c.Param("id")
	quality := c.Query("quality")

	if quality == "" {
		quality = "720p" // 默认质量
	}

	// 构建视频文件路径
	videoPath := filepath.Join(VideoDir, videoID, quality+".mp4")
	log.Printf("Attempting to stream video from: %s", videoPath)

	// 检查文件是否存在且可读
	if _, err := os.Stat(videoPath); os.IsNotExist(err) {
		log.Printf("Video file not found: %s", videoPath)
		c.JSON(http.StatusNotFound, gin.H{"error": "Video not found"})
		return
	}

	// 验证文件是否完整
	if err := VerifyFile(videoPath); err != nil {
		log.Printf("Video file verification failed: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Video file is not valid"})
		return
	}

	// 设置响应头
	c.Header("Content-Type", "video/mp4")
	c.Header("Accept-Ranges", "bytes")
	c.Header("Cache-Control", "no-cache")
	c.Header("X-Content-Type-Options", "nosniff")

	// 使用 http.ServeFile 来正确处理文件服务
	http.ServeFile(c.Writer, c.Request, videoPath)
}

func GetVideoInfo(c *gin.Context) {
	videoID := c.Param("id")
	videoDir := filepath.Join("./videos", videoID)

	files, err := os.ReadDir(videoDir)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Video not found"})
		return
	}

	var qualities []map[string]string
	for _, f := range files {
		if !f.IsDir() && strings.HasSuffix(f.Name(), ".mp4") {
			resolution := strings.TrimSuffix(f.Name(), ".mp4")
			qualities = append(qualities, map[string]string{
				"resolution": resolution,
				"path":       fmt.Sprintf("/api/videos/%s/stream?quality=%s", videoID, resolution),
			})
		}
	}

	c.JSON(http.StatusOK, gin.H{
		"id":        videoID,
		"title":     videoID + ".mp4",
		"qualities": qualities,
	})
}

func getVideoInfo(videoID string) (*models.Video, error) {
	// TODO: 从数据库获取视频信息
	// 这里暂时返回模拟数据
	return &models.Video{
		ID:    videoID,
		Title: "Sample Video",
		Qualities: []models.Quality{
			{Resolution: "1080p", Path: "/api/videos/" + videoID + "/stream?quality=1080p"},
			{Resolution: "720p", Path: "/api/videos/" + videoID + "/stream?quality=720p"},
			{Resolution: "480p", Path: "/api/videos/" + videoID + "/stream?quality=480p"},
		},
	}, nil
}

func verifyFile(filePath string) error {
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
