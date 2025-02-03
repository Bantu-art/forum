package controllers

import (
	"crypto/md5"
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"io"
	"log"
	"mime/multipart"
	"os"
	"path/filepath"
	"strings"
	"time"
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
		return "", err
	}

	// Return web-accessible path
	return "/static/uploads/" + newFileName, nil
}

// isValidFileType checks if the file type is allowed
func (ih *ImageHandler) isValidFileType(filename string) bool {
	ext := strings.ToLower(filepath.Ext(filename))
	validTypes := map[string]bool{
		".jpg":  true,
		".jpeg": true,
		".png":  true,
		".gif":  true,
	}
	return validTypes[ext]
}

// generateUniqueFilename creates a unique filename
func (ih *ImageHandler) generateUniqueFilename(originalFilename string) string {
	bytes := make([]byte, 16)
	if _, err := rand.Read(bytes); err != nil {
		log.Printf("Error generating random bytes: %v", err)
	}
	return hex.EncodeToString(bytes) + filepath.Ext(originalFilename)
}
