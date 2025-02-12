package handlers

import (
	"net/http"
	"net/http/httptest"
	"os"
	"testing"
)

func TestServeStatic(t *testing.T) {
	// Create a temporary directory for test files
	tempDir, err := os.MkdirTemp("", "test")
	if err != nil {
		t.Fatalf("Failed to create temp directory: %v", err)
	}
	defer os.RemoveAll(tempDir)

	// Create a test file
	testFile := tempDir + "/testfile.txt"
	err = os.WriteFile(testFile, []byte("test content"), 0o644)
	if err != nil {
		t.Fatalf("Failed to create test file: %v", err)
	}

	type args struct {
		w http.ResponseWriter
		r *http.Request
	}
	tests := []struct {
		name           string
		args           args
		expectedStatus int
		expectedBody   string
	}{
		{
			name: "Existing file",
			args: args{
				w: httptest.NewRecorder(),
				r: httptest.NewRequest("GET", "/testfile.txt", nil),
			},
			expectedStatus: http.StatusOK,
			expectedBody:   "test content",
		},
		{
			name: "Non-existent file",
			args: args{
				w: httptest.NewRecorder(),
				r: httptest.NewRequest("GET", "/nonexistent.txt", nil),
			},
			expectedStatus: http.StatusNotFound,
			expectedBody:   "Internal Server Error\n",
		},
		{
			name: "Directory",
			args: args{
				w: httptest.NewRecorder(),
				r: httptest.NewRequest("GET", "/", nil),
			},
			expectedStatus: http.StatusNotFound,
			expectedBody:   "Internal Server Error\n",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Set the working directory to the temp directory
			oldWd, _ := os.Getwd()
			os.Chdir(tempDir)
			defer os.Chdir(oldWd)

			ServeStatic(tt.args.w, tt.args.r)

			resp := tt.args.w.(*httptest.ResponseRecorder)
			if resp.Code != tt.expectedStatus {
				t.Errorf("ServeStatic() status = %v, want %v", resp.Code, tt.expectedStatus)
			}
		})
	}
}
