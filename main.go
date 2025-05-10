package main

import (
	"log"
	"os"
	"path/filepath"
	"time"

	"video-streaming/handlers"
	"video-streaming/models"

	"github.com/gin-gonic/gin"
)

const UploadDir = "./videos/temp"

func main() {
	// 设置日志
	log.SetFlags(log.LstdFlags | log.Lshortfile)
	log.Println("Starting video streaming server...")

	// 确保必要的目录存在
	dirs := []string{
		"./videos",
		"./videos/temp",
		"./frontend",
	}
	for _, dir := range dirs {
		if err := os.MkdirAll(dir, 0755); err != nil {
			log.Fatalf("Failed to create directory %s: %v", dir, err)
		}
	}

	// 初始化数据库
	dbPath := "./videos.db"
	if err := models.InitDB(dbPath); err != nil {
		log.Fatalf("Failed to initialize database: %v", err)
	}
	log.Println("Database initialized successfully")

	// 设置 Gin 模式
	gin.SetMode(gin.DebugMode)
	r := gin.Default()

	// 静态文件服务
	r.Static("/videos", "./videos")
	r.Static("/static", "./frontend")

	// 路由设置
	r.GET("/", func(c *gin.Context) {
		c.File("./frontend/index.html")
	})

	// API 路由
	api := r.Group("/api")
	{
		// 视频上传相关
		api.POST("/upload/init", handlers.InitUpload)
		api.POST("/upload/chunk", handlers.UploadChunk)
		api.POST("/upload/complete", handlers.CompleteUpload)

		// 视频列表和播放相关
		api.GET("/videos", handlers.GetVideoList)
		api.GET("/videos/:id", handlers.GetVideoInfo)
		api.GET("/videos/:id/info", handlers.GetVideoInfo)
		api.GET("/videos/:id/stream", handlers.StreamVideo)
	}

	// 启动清理任务
	cleanupTempFiles()

	// 启动服务器
	port := ":8080"
	log.Printf("Server starting on port %s...", port)
	if err := r.Run(port); err != nil {
		log.Fatalf("Failed to start server: %v", err)
	}
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
		exists := false
		for _, res := range []string{"1080p.mp4", "720p.mp4", "480p.mp4"} {
			if fileExists(videoDir + "/" + res) {
				exists = true
				break
			}
		}
		if !exists {
			models.DB.Exec(`DELETE FROM videos WHERE id = ?`, id)
		}
	}
	return nil
}

func cleanupTempFiles() {
	ticker := time.NewTicker(1 * time.Hour)
	go func() {
		for range ticker.C {
			files, err := os.ReadDir(UploadDir)
			if err != nil {
				log.Printf("Failed to read temp directory: %v", err)
				continue
			}

			for _, file := range files {
				if file.IsDir() {
					// 检查目录是否超过24小时
					info, err := file.Info()
					if err != nil {
						continue
					}
					if time.Since(info.ModTime()) > 24*time.Hour {
						path := filepath.Join(UploadDir, file.Name())
						if err := os.RemoveAll(path); err != nil {
							log.Printf("Failed to remove old temp directory %s: %v", path, err)
						}
					}
				}
			}
		}
	}()
}
