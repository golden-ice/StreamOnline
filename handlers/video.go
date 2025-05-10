package handlers

import (
	"net/http"
	"strconv"
	"video-streaming/models"

	"log"

	"github.com/gin-gonic/gin"
)

func GetVideoList(c *gin.Context) {
	log.Printf("Getting video list...")

	// 获取分页参数
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "10"))
	offset, _ := strconv.Atoi(c.DefaultQuery("offset", "0"))

	// 获取视频列表
	videos, err := models.GetVideoList(limit, offset)
	if err != nil {
		log.Printf("Error getting video list: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get video list"})
		return
	}

	log.Printf("Found %d videos", len(videos))
	for _, v := range videos {
		log.Printf("Video: ID=%s, Title=%s, Status=%s, Qualities=%v",
			v.ID, v.Title, v.Status, v.Qualities)
	}

	c.JSON(http.StatusOK, videos)
}
