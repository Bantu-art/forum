package controllers

import (
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"io"
	"log"
	"mime/multipart"
	"os"
	"path/filepath"
	"strings"
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
	// Validate file size
	if header.Size > maxUploadSize {
		return "", fmt.Errorf("file size exceeds maximum limit")
	}

	// Validate file type
	if !ih.isValidFileType(header.Filename) {
		return "", fmt.Errorf("invalid file type")
	}

	// Generate unique filename
	filename := ih.generateUniqueFilename(header.Filename)
	filepath := filepath.Join(ih.uploadPath, filename)

	// Create destination file
	dst, err := os.Create(filepath)
	if err != nil {
		return "", fmt.Errorf("error creating destination file: %v", err)
	}
	defer dst.Close()

	// Copy uploaded file to destination
	if _, err := io.Copy(dst, file); err != nil {
		return "", fmt.Errorf("error copying file: %v", err)
	}

	// Return the relative path for database storage
	return filepath, nil
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
