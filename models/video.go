package models

import (
	"time"

	"github.com/google/uuid"
)

type Video struct {
	ID          string    `json:"id"`
	Title       string    `json:"title"`
	FileName    string    `json:"fileName"`
	FileSize    int64     `json:"fileSize"`
	ContentType string    `json:"contentType"`
	Status      string    `json:"status"` // pending, processing, ready, error
	Qualities   []Quality `json:"qualities"`
	CreatedAt   time.Time `json:"createdAt"`
	UpdatedAt   time.Time `json:"updatedAt"`
}

type Quality struct {
	ID         int64  `json:"id"`
	VideoID    string `json:"videoId"`
	Resolution string `json:"resolution"` // 例如: "1080p", "720p", "480p"
	Path       string `json:"path"`       // 视频文件路径
	Size       int64  `json:"size"`       // 文件大小
}

// 保存视频信息到数据库
func (v *Video) Save() error {
	tx, err := DB.Begin()
	if err != nil {
		return err
	}
	defer tx.Rollback()

	// 插入视频信息
	_, err = tx.Exec(`
		INSERT INTO videos (id, title, file_name, file_size, content_type, status, created_at, updated_at)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?)
	`, v.ID, v.Title, v.FileName, v.FileSize, v.ContentType, v.Status, v.CreatedAt, v.UpdatedAt)
	if err != nil {
		return err
	}

	// 插入视频质量信息
	for _, quality := range v.Qualities {
		_, err = tx.Exec(`
			INSERT INTO video_qualities (video_id, resolution, path, size)
			VALUES (?, ?, ?, ?)
		`, v.ID, quality.Resolution, quality.Path, quality.Size)
		if err != nil {
			return err
		}
	}

	return tx.Commit()
}

// 从数据库获取视频信息
func GetVideoByID(id string) (*Video, error) {
	var v Video
	err := DB.QueryRow(`
		SELECT id, title, file_name, file_size, content_type, status, created_at, updated_at
		FROM videos WHERE id = ?
	`, id).Scan(&v.ID, &v.Title, &v.FileName, &v.FileSize, &v.ContentType, &v.Status, &v.CreatedAt, &v.UpdatedAt)
	if err != nil {
		return nil, err
	}

	// 获取视频质量信息
	rows, err := DB.Query(`
		SELECT id, video_id, resolution, path, size
		FROM video_qualities WHERE video_id = ?
	`, id)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	for rows.Next() {
		var q Quality
		err := rows.Scan(&q.ID, &q.VideoID, &q.Resolution, &q.Path, &q.Size)
		if err != nil {
			return nil, err
		}
		v.Qualities = append(v.Qualities, q)
	}

	return &v, nil
}

// 更新视频状态
func UpdateVideoStatus(id, status string) error {
	_, err := DB.Exec(`
		UPDATE videos SET status = ?, updated_at = ?
		WHERE id = ?
	`, status, time.Now(), id)
	return err
}

// 获取视频列表
func GetVideoList(limit, offset int) ([]*Video, error) {
	rows, err := DB.Query(`
		SELECT id, title, file_name, file_size, content_type, status, created_at, updated_at
		FROM videos
		ORDER BY created_at DESC
		LIMIT ? OFFSET ?
	`, limit, offset)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var videos []*Video
	for rows.Next() {
		var v Video
		err := rows.Scan(&v.ID, &v.Title, &v.FileName, &v.FileSize, &v.ContentType, &v.Status, &v.CreatedAt, &v.UpdatedAt)
		if err != nil {
			return nil, err
		}

		// 获取视频质量信息
		qualityRows, err := DB.Query(`
			SELECT id, video_id, resolution, path, size
			FROM video_qualities WHERE video_id = ?
		`, v.ID)
		if err != nil {
			return nil, err
		}
		for qualityRows.Next() {
			var q Quality
			err := qualityRows.Scan(&q.ID, &q.VideoID, &q.Resolution, &q.Path, &q.Size)
			if err != nil {
				qualityRows.Close()
				return nil, err
			}
			v.Qualities = append(v.Qualities, q)
		}
		qualityRows.Close()
		videos = append(videos, &v)
	}
	return videos, nil
}

func CreateVideo(video *Video) error {
	video.ID = uuid.New().String()
	_, err := DB.Exec(`
		INSERT INTO videos (id, title, file_name, file_size, content_type, status, created_at, updated_at)
		VALUES (?, ?, ?, ?, ?, ?, datetime('now'), datetime('now'))
	`, video.ID, video.Title, video.FileName, video.FileSize, video.ContentType, video.Status)
	return err
}

func UpdateVideo(video *Video) error {
	_, err := DB.Exec(`
		UPDATE videos
		SET status = ?, updated_at = datetime('now')
		WHERE id = ?
	`, video.Status, video.ID)
	return err
}

func CreateQuality(quality *Quality) error {
	_, err := DB.Exec(`
		INSERT INTO video_qualities (video_id, resolution, path, size)
		VALUES (?, ?, ?, ?)
	`, quality.VideoID, quality.Resolution, quality.Path, quality.Size)
	return err
}
