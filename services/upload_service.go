package services

import (
	"fmt"
	"io"
	"os"
	"path/filepath"
)

type UploadService struct {
	BaseDir string
}

func NewUploadService(baseDir string) *UploadService {
	return &UploadService{
		BaseDir: baseDir,
	}
}

func (s *UploadService) MergeChunks(uploadID string, totalChunks int) error {
	uploadPath := filepath.Join(s.BaseDir, uploadID)
	outputPath := filepath.Join(s.BaseDir, uploadID, "original.mp4")

	outputFile, err := os.Create(outputPath)
	if err != nil {
		return fmt.Errorf("failed to create output file: %v", err)
	}
	defer outputFile.Close()

	for i := 0; i < totalChunks; i++ {
		chunkPath := filepath.Join(uploadPath, fmt.Sprintf("chunk_%d", i))
		chunkFile, err := os.Open(chunkPath)
		if err != nil {
			return fmt.Errorf("failed to open chunk %d: %v", i, err)
		}

		_, err = io.Copy(outputFile, chunkFile)
		chunkFile.Close()
		if err != nil {
			return fmt.Errorf("failed to copy chunk %d: %v", i, err)
		}
	}

	return nil
}
