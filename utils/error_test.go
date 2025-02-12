package utils

import (
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestRenderErrorPage(t *testing.T) {
	// Create a temporary directory for templates
	tmpDir, err := os.MkdirTemp("", "forum-test")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	// Create templates directory
	templatesDir := filepath.Join(tmpDir, "templates")
	if err := os.MkdirAll(templatesDir, 0o755); err != nil {
		t.Fatalf("Failed to create templates dir: %v", err)
	}

	// Create error.html template
	errorTemplate := `<!DOCTYPE html>
<html>
<head>
    <title>Error</title>
</head>
<body>
    <h1>Error {{.Code}}</h1>
    <p>{{.Message}}</p>
</body>
</html>`

	if err := os.WriteFile(filepath.Join(templatesDir, "error.html"), []byte(errorTemplate), 0o644); err != nil {
		t.Fatalf("Failed to create error template: %v", err)
	}

	// Change working directory to temp dir for test
	origWd, err := os.Getwd()
	if err != nil {
		t.Fatalf("Failed to get working directory: %v", err)
	}
	if err := os.Chdir(tmpDir); err != nil {
		t.Fatalf("Failed to change working directory: %v", err)
	}
	defer os.Chdir(origWd)

	tests := []struct {
		name       string
		code       int
		message    string
		wantInBody []string
	}{
		{
			name:       "Not Found Error",
			code:       http.StatusNotFound,
			message:    ErrNotFound,
			wantInBody: []string{"404", ErrNotFound},
		},
		{
			name:       "Internal Server Error",
			code:       http.StatusInternalServerError,
			message:    ErrInternalServer,
			wantInBody: []string{"500", ErrInternalServer},
		},
		{
			name:       "Unauthorized Error",
			code:       http.StatusUnauthorized,
			message:    ErrUnauthorized,
			wantInBody: []string{"401", ErrUnauthorized},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			recorder := httptest.NewRecorder()
			RenderErrorPage(recorder, tt.code, tt.message)

			// Check status code
			if recorder.Code != tt.code {
				t.Errorf("RenderErrorPage() status code = %v, want %v", recorder.Code, tt.code)
			}

			// Check response body contains expected content
			body := recorder.Body.String()
			for _, want := range tt.wantInBody {
				if !strings.Contains(body, want) {
					t.Errorf("RenderErrorPage() body does not contain %q", want)
				}
			}

			// Verify Content-Type header
			contentType := recorder.Header().Get("Content-Type")
			if !strings.Contains(contentType, "text/html") {
				t.Errorf("RenderErrorPage() Content-Type = %v, want text/html", contentType)
			}
		})
	}
}
