package handlers

import (
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"sort"
	"strconv"
	"strings"
	"time"
	"video-streaming/models"
	"video-streaming/services"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

const (
	ChunkSize = 1024 * 1024 // 1MB
	UploadDir = "./videos/temp"
)

func InitUpload(c *gin.Context) {
	var uploadInfo struct {
		FileName    string `json:"fileName"`
		FileSize    int64  `json:"fileSize"`
		ContentType string `json:"contentType"`
	}

	if err := c.ShouldBindJSON(&uploadInfo); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// 创建上传ID
	uploadID := uuid.New().String()

	// 创建临时目录
	uploadPath := filepath.Join(UploadDir, uploadID)
	if err := os.MkdirAll(uploadPath, 0755); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create upload directory"})
		return
	}

	// 创建视频记录
	video := &models.Video{
		ID:          uploadID,
		Title:       uploadInfo.FileName,
		FileName:    uploadInfo.FileName,
		FileSize:    uploadInfo.FileSize,
		ContentType: uploadInfo.ContentType,
		Status:      "pending",
		CreatedAt:   time.Now(),
		UpdatedAt:   time.Now(),
	}

	// 保存到数据库
	if err := video.Save(); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to save video info"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"uploadId":  uploadID,
		"chunkSize": ChunkSize,
	})
}

func UploadChunk(c *gin.Context) {
	uploadID := c.PostForm("uploadId")
	chunkIndex := c.PostForm("chunkIndex")

	if uploadID == "" || chunkIndex == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Missing required parameters"})
		return
	}

	// 获取上传的文件
	file, _, err := c.Request.FormFile("chunk")
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Failed to get chunk file"})
		return
	}
	defer file.Close()

	// 创建分片文件
	chunkPath := filepath.Join(UploadDir, uploadID, fmt.Sprintf("chunk_%s", chunkIndex))
	chunkFile, err := os.Create(chunkPath)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create chunk file"})
		return
	}
	defer chunkFile.Close()

	// 保存分片
	if _, err := io.Copy(chunkFile, file); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to save chunk"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message":    "Chunk uploaded successfully",
		"chunkIndex": chunkIndex,
	})
}

func CompleteUpload(c *gin.Context) {
	var completeInfo struct {
		UploadID string `json:"uploadId"`
	}

	if err := c.ShouldBindJSON(&completeInfo); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	tempDir := filepath.Join(UploadDir, completeInfo.UploadID)
	outputDir := filepath.Join("./videos", completeInfo.UploadID)

	// 确保输出目录存在
	if err := os.MkdirAll(outputDir, 0755); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create output directory"})
		return
	}

	outputFile := filepath.Join(outputDir, "original.mp4")

	// 创建输出文件
	out, err := os.Create(outputFile)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create output file"})
		return
	}
	defer out.Close()

	// 读取所有分片并排序
	files, err := os.ReadDir(tempDir)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to read temp directory"})
		return
	}

	var chunkFiles []string
	for _, f := range files {
		if !f.IsDir() && strings.HasPrefix(f.Name(), "chunk_") {
			chunkFiles = append(chunkFiles, f.Name())
		}
	}

	// 按 chunk_数字 排序
	sort.Slice(chunkFiles, func(i, j int) bool {
		a, _ := strconv.Atoi(strings.TrimPrefix(chunkFiles[i], "chunk_"))
		b, _ := strconv.Atoi(strings.TrimPrefix(chunkFiles[j], "chunk_"))
		return a < b
	})

	// 合并分片
	for _, fname := range chunkFiles {
		chunkPath := filepath.Join(tempDir, fname)
		chunkData, err := os.ReadFile(chunkPath)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to read chunk"})
			return
		}
		if _, err := out.Write(chunkData); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to write chunk"})
			return
		}
		// 删除已处理的分片文件
		os.Remove(chunkPath)
	}

	// 关闭文件句柄
	out.Close()

	// 验证文件是否完整
	if err := VerifyFile(outputFile); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "File verification failed: " + err.Error()})
		return
	}

	// 更新视频状态为 processing
	if err := models.UpdateVideoStatus(completeInfo.UploadID, "processing"); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update video status"})
		return
	}

	// 开始转码
	go func() {
		transcodeService := services.NewTranscodeService("./videos")
		if err := transcodeService.TranscodeVideo(completeInfo.UploadID); err != nil {
			log.Printf("Transcoding failed for upload %s: %v", completeInfo.UploadID, err)
			models.UpdateVideoStatus(completeInfo.UploadID, "error")
			return
		}

		// 验证转码后的文件
		if err := VerifyTranscodedFiles(completeInfo.UploadID); err != nil {
			log.Printf("Transcoded files verification failed for %s: %v", completeInfo.UploadID, err)
			models.UpdateVideoStatus(completeInfo.UploadID, "error")
			return
		}

		models.UpdateVideoStatus(completeInfo.UploadID, "ready")

		// 清理临时目录
		if err := os.RemoveAll(tempDir); err != nil {
			log.Printf("Failed to remove temp directory %s: %v", tempDir, err)
		}
	}()

	c.JSON(http.StatusOK, gin.H{
		"message": "Upload completed, transcoding started",
	})
}

func UploadVideo(c *gin.Context) {
	// 确保目录存在
	if err := os.MkdirAll(UploadDir, 0755); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create temp directory"})
		return
	}

	// 获取上传的文件
	file, header, err := c.Request.FormFile("video")
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "No video file provided"})
		return
	}
	defer file.Close()

	// 创建临时文件
	tempFile := filepath.Join(UploadDir, header.Filename)
	out, err := os.Create(tempFile)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create temp file"})
		return
	}
	defer out.Close()

	// 保存文件
	_, err = io.Copy(out, file)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to save file"})
		return
	}

	// 创建视频记录
	video := &models.Video{
		Title:       header.Filename,
		FileName:    header.Filename,
		FileSize:    header.Size,
		ContentType: header.Header.Get("Content-Type"),
		Status:      "processing",
	}

	if err := models.CreateVideo(video); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create video record"})
		return
	}

	// 启动转码
	go func() {
		if err := services.TranscodeVideo(video.ID, tempFile); err != nil {
			fmt.Printf("Transcoding failed: %v\n", err)
			// 更新视频状态为失败
			video.Status = "failed"
			models.UpdateVideo(video)
			return
		}
		// 更新视频状态为完成
		video.Status = "ready"
		models.UpdateVideo(video)
	}()

	c.JSON(http.StatusOK, gin.H{
		"message": "Video upload started",
		"videoId": video.ID,
	})
}

func fileExists(path string) bool {
	_, err := os.Stat(path)
	return err == nil
}

func CleanInvalidVideos() error {
	rows, err := models.DB.Query(`SELECT id FROM videos`)
	if err != nil {
		return err
	}
	defer rows.Close()

	for rows.Next() {
		var id string
		if err := rows.Scan(&id); err != nil {
			continue
		}
		videoDir := "./videos/" + id
		found := false
		for _, res := range []string{"1080p.mp4", "720p.mp4", "480p.mp4"} {
			if fileExists(videoDir + "/" + res) {
				found = true
				break
			}
		}
		if !found {
			models.DB.Exec(`DELETE FROM videos WHERE id = ?`, id)
		}
	}
	return nil
}
