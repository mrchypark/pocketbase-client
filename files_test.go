package pocketbase

import (
	"bytes"
	"context"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestFileService_Upload(t *testing.T) {
	tests := []struct {
		name       string
		collection string
		recordID   string
		fieldName  string
		filename   string
		file       io.Reader
		wantErr    bool
		errMsg     string
	}{
		{
			name:       "successful file upload",
			collection: "posts",
			recordID:   "record123",
			fieldName:  "image",
			filename:   "test.jpg",
			file:       strings.NewReader("fake image data"),
			wantErr:    false,
		},
		{
			name:       "empty collection name",
			collection: "",
			recordID:   "record123",
			fieldName:  "image",
			filename:   "test.jpg",
			file:       strings.NewReader("fake image data"),
			wantErr:    true,
			errMsg:     "collection name is required",
		},
		{
			name:       "empty record ID",
			collection: "posts",
			recordID:   "",
			fieldName:  "image",
			filename:   "test.jpg",
			file:       strings.NewReader("fake image data"),
			wantErr:    true,
			errMsg:     "record ID is required",
		},
		{
			name:       "empty field name",
			collection: "posts",
			recordID:   "record123",
			fieldName:  "",
			filename:   "test.jpg",
			file:       strings.NewReader("fake image data"),
			wantErr:    true,
			errMsg:     "field name is required",
		},
		{
			name:       "empty filename",
			collection: "posts",
			recordID:   "record123",
			fieldName:  "image",
			filename:   "",
			file:       strings.NewReader("fake image data"),
			wantErr:    true,
			errMsg:     "filename is required",
		},
		{
			name:       "nil file",
			collection: "posts",
			recordID:   "record123",
			fieldName:  "image",
			filename:   "test.jpg",
			file:       nil,
			wantErr:    true,
			errMsg:     "file reader is required",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Mock server setup
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				if tt.wantErr && tt.errMsg != "" {
					// Parameter validation errors occur before reaching the server
					return
				}

				// Simulate successful response
				w.Header().Set("Content-Type", "application/json")
				w.WriteHeader(http.StatusOK)
				w.Write([]byte(`{
					"id": "record123",
					"collectionId": "collection456",
					"collectionName": "posts",
					"created": "2023-01-01T00:00:00.000Z",
					"updated": "2023-01-01T00:00:00.000Z",
					"image": "test.jpg"
				}`))
			}))
			defer server.Close()

			client := NewClient(server.URL)
			ctx := context.Background()

			result, err := client.Files.Upload(ctx, tt.collection, tt.recordID, tt.fieldName, tt.filename, tt.file)

			if tt.wantErr {
				if err == nil {
					t.Errorf("Upload() error = nil, wantErr %v", tt.wantErr)
					return
				}
				if tt.errMsg != "" && !strings.Contains(err.Error(), tt.errMsg) {
					t.Errorf("Upload() error = %v, want error containing %v", err, tt.errMsg)
				}
				return
			}

			if err != nil {
				t.Errorf("Upload() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if result == nil {
				t.Error("Upload() result is nil")
				return
			}

			if result.ID != "record123" {
				t.Errorf("Upload() result.ID = %v, want %v", result.ID, "record123")
			}
		})
	}
}

func TestFileService_Download(t *testing.T) {
	tests := []struct {
		name       string
		collection string
		recordID   string
		filename   string
		wantErr    bool
		errMsg     string
	}{
		{
			name:       "successful file download",
			collection: "posts",
			recordID:   "record123",
			filename:   "test.jpg",
			wantErr:    false,
		},
		{
			name:       "empty collection name",
			collection: "",
			recordID:   "record123",
			filename:   "test.jpg",
			wantErr:    true,
			errMsg:     "collection name is required",
		},
		{
			name:       "empty record ID",
			collection: "posts",
			recordID:   "",
			filename:   "test.jpg",
			wantErr:    true,
			errMsg:     "record ID is required",
		},
		{
			name:       "empty filename",
			collection: "posts",
			recordID:   "record123",
			filename:   "",
			wantErr:    true,
			errMsg:     "filename is required",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Mock server setup
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				if tt.wantErr && tt.errMsg != "" {
					// Parameter validation errors occur before reaching the server
					return
				}

				// Simulate successful file response
				w.Header().Set("Content-Type", "image/jpeg")
				w.WriteHeader(http.StatusOK)
				w.Write([]byte("fake image data"))
			}))
			defer server.Close()

			client := NewClient(server.URL)
			ctx := context.Background()

			result, err := client.Files.Download(ctx, tt.collection, tt.recordID, tt.filename, nil)

			if tt.wantErr {
				if err == nil {
					t.Errorf("Download() error = nil, wantErr %v", tt.wantErr)
					return
				}
				if tt.errMsg != "" && !strings.Contains(err.Error(), tt.errMsg) {
					t.Errorf("Download() error = %v, want error containing %v", err, tt.errMsg)
				}
				return
			}

			if err != nil {
				t.Errorf("Download() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if result == nil {
				t.Error("Download() result is nil")
				return
			}

			defer result.Close()

			// Test reading file content
			data, err := io.ReadAll(result)
			if err != nil {
				t.Errorf("Failed to read downloaded file: %v", err)
				return
			}

			expected := "fake image data"
			if string(data) != expected {
				t.Errorf("Downloaded data = %v, want %v", string(data), expected)
			}
		})
	}
}

func TestFileService_GetFileURL(t *testing.T) {
	client := NewClient("https://example.com")

	tests := []struct {
		name       string
		collection string
		recordID   string
		filename   string
		opts       *FileDownloadOptions
		expected   string
	}{
		{
			name:       "basic file URL",
			collection: "posts",
			recordID:   "record123",
			filename:   "test.jpg",
			opts:       nil,
			expected:   "https://example.com/api/files/posts/record123/test.jpg",
		},
		{
			name:       "with thumbnail option",
			collection: "posts",
			recordID:   "record123",
			filename:   "test.jpg",
			opts:       &FileDownloadOptions{Thumb: "100x100"},
			expected:   "https://example.com/api/files/posts/record123/test.jpg?thumb=100x100",
		},
		{
			name:       "with download option",
			collection: "posts",
			recordID:   "record123",
			filename:   "test.jpg",
			opts:       &FileDownloadOptions{Download: true},
			expected:   "https://example.com/api/files/posts/record123/test.jpg?download=1",
		},
		{
			name:       "with all options",
			collection: "posts",
			recordID:   "record123",
			filename:   "test.jpg",
			opts:       &FileDownloadOptions{Thumb: "100x100", Download: true},
			expected:   "https://example.com/api/files/posts/record123/test.jpg?download=1&thumb=100x100",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := client.Files.GetFileURL(tt.collection, tt.recordID, tt.filename, tt.opts)
			if result != tt.expected {
				t.Errorf("GetFileURL() = %v, want %v", result, tt.expected)
			}
		})
	}
}

func TestFileService_Delete(t *testing.T) {
	tests := []struct {
		name       string
		collection string
		recordID   string
		fieldName  string
		filename   string
		wantErr    bool
		errMsg     string
	}{
		{
			name:       "empty collection name",
			collection: "",
			recordID:   "record123",
			fieldName:  "image",
			filename:   "test.jpg",
			wantErr:    true,
			errMsg:     "collection name is required",
		},
		{
			name:       "empty record ID",
			collection: "posts",
			recordID:   "",
			fieldName:  "image",
			filename:   "test.jpg",
			wantErr:    true,
			errMsg:     "record ID is required",
		},
		{
			name:       "empty field name",
			collection: "posts",
			recordID:   "record123",
			fieldName:  "",
			filename:   "test.jpg",
			wantErr:    true,
			errMsg:     "field name is required",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			client := NewClient("https://example.com")
			ctx := context.Background()

			_, err := client.Files.Delete(ctx, tt.collection, tt.recordID, tt.fieldName, tt.filename)

			if tt.wantErr {
				if err == nil {
					t.Errorf("Delete() error = nil, wantErr %v", tt.wantErr)
					return
				}
				if tt.errMsg != "" && !strings.Contains(err.Error(), tt.errMsg) {
					t.Errorf("Delete() error = %v, want error containing %v", err, tt.errMsg)
				}
				return
			}

			if err != nil {
				t.Errorf("Delete() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

// Benchmark test
func BenchmarkFileService_Upload(b *testing.B) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{
			"id": "record123",
			"collectionId": "collection456",
			"collectionName": "posts",
			"created": "2023-01-01T00:00:00.000Z",
			"updated": "2023-01-01T00:00:00.000Z",
			"image": "test.jpg"
		}`))
	}))
	defer server.Close()

	client := NewClient(server.URL)
	ctx := context.Background()
	fileData := bytes.NewReader(make([]byte, 1024)) // 1KB test file

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		fileData.Seek(0, 0) // Reset file pointer
		_, err := client.Files.Upload(ctx, "posts", "record123", "image", "test.jpg", fileData)
		if err != nil {
			b.Fatalf("Upload failed: %v", err)
		}
	}
}
