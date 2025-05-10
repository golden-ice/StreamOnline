package models

import (
	"database/sql"

	_ "modernc.org/sqlite"
)

var DB *sql.DB

func InitDB(dbPath string) error {
	var err error
	DB, err = sql.Open("sqlite", dbPath)
	if err != nil {
		return err
	}

	// 创建视频表
	_, err = DB.Exec(`
        CREATE TABLE IF NOT EXISTS videos (
            id TEXT PRIMARY KEY,
            title TEXT NOT NULL,
            file_name TEXT NOT NULL,
            file_size INTEGER NOT NULL,
            content_type TEXT NOT NULL,
            status TEXT NOT NULL,
            created_at DATETIME NOT NULL,
            updated_at DATETIME NOT NULL
        )
    `)
	if err != nil {
		return err
	}

	// 创建视频质量表
	_, err = DB.Exec(`
        CREATE TABLE IF NOT EXISTS video_qualities (
            id INTEGER PRIMARY KEY AUTOINCREMENT,
            video_id TEXT NOT NULL,
            resolution TEXT NOT NULL,
            path TEXT NOT NULL,
            size INTEGER NOT NULL,
            FOREIGN KEY (video_id) REFERENCES videos(id)
        )
    `)
	if err != nil {
		return err
	}

	return nil
}
