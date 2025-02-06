package utils

import (
	"errors"
	"mime/multipart"
	"path/filepath"
)

const MaxFileSize = 10 << 20 // 10MB

var ValidImageTypes = map[string]bool{
	"image/jpeg": true,
	"image/png":  true,
	"image/gif":  true,
}

func ValidateImage(file multipart.File, header *multipart.FileHeader) error {
	// Check file size
	if header.Size > MaxFileSize {
		return errors.New(ErrFileTooLarge)
	}

	// Check file type
	ext := filepath.Ext(header.Filename)
	switch ext {
	case ".jpg", ".jpeg", ".png", ".gif":
		contentType := header.Header.Get("Content-Type")
		if !ValidImageTypes[contentType] {
			return errors.New(ErrInvalidFileType)
		}
	default:
		return errors.New(ErrInvalidFileType)
	}

	return nil
}
