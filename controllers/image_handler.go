package controllers

import (
	"crypto/md5"
	"fmt"
	"io"
	"log"
	"mime/multipart"
	"os"
	"path/filepath"
	"time"

	"forum/utils"
)

const (
	maxUploadSize = 10 << 20 // 10MB
	uploadDir     = "static/uploads"
)

// ImageHandler handles image upload and processing
type ImageHandler struct {
	uploadPath string
}

// NewImageHandler creates a new image handler
func NewImageHandler() *ImageHandler {
	// Create upload directory if it doesn't exist
	if err := os.MkdirAll(uploadDir, 0o755); err != nil {
		log.Printf("Error creating upload directory: %v", err)
	}
	return &ImageHandler{uploadPath: uploadDir}
}

// ProcessImage handles the image upload process
func (ih *ImageHandler) ProcessImage(file multipart.File, header *multipart.FileHeader) (string, error) {
	if err := utils.ValidateImage(file, header); err != nil {
		return "", err
	}
	// Generate unique filename
	ext := filepath.Ext(header.Filename)
	newFileName := fmt.Sprintf("%x%s", md5.Sum([]byte(time.Now().String())), ext)

	// Create uploads directory if it doesn't exist
	uploadsDir := "static/uploads"
	if err := os.MkdirAll(uploadsDir, 0o755); err != nil {
		return "", err
	}

	// Save file
	filePath := filepath.Join(uploadsDir, newFileName)
	dst, err := os.Create(filePath)
	if err != nil {
		return "", err
	}
	defer dst.Close()

	if _, err = io.Copy(dst, file); err != nil {
		os.Remove(filePath) // Clean up on error
		return "", err
	}

	// Return web-accessible path
	return "/static/uploads/" + newFileName, nil
}
