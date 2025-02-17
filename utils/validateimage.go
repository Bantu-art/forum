package utils

import (
	"errors"
	"mime/multipart"
	"path/filepath"
)

const MaxFileSize = 20 << 20 // 20MB

var ValidImageTypes = map[string]bool{
	"image/jpeg": true,
	"image/png":  true,
	"image/gif":  true,
}

func ValidateImage(file multipart.File, header *multipart.FileHeader) error {
	// Check file size
	if header.Size > MaxFileSize {
		return errors.New("file too large")
	}

	// Check file type
	ext := filepath.Ext(header.Filename)
	switch ext {
	case ".jpg", ".jpeg", ".png", ".gif":
		contentType := header.Header.Get("Content-Type")
		if !ValidImageTypes[contentType] {
			return errors.New("invalid file type")
		}
	default:
		return errors.New("invalid file type")
	}

	return nil
}
