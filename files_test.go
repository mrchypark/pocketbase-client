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
			name:       "성공적인 파일 업로드",
			collection: "posts",
			recordID:   "record123",
			fieldName:  "image",
			filename:   "test.jpg",
			file:       strings.NewReader("fake image data"),
			wantErr:    false,
		},
		{
			name:       "빈 컬렉션 이름",
			collection: "",
			recordID:   "record123",
			fieldName:  "image",
			filename:   "test.jpg",
			file:       strings.NewReader("fake image data"),
			wantErr:    true,
			errMsg:     "collection name is required",
		},
		{
			name:       "빈 레코드 ID",
			collection: "posts",
			recordID:   "",
			fieldName:  "image",
			filename:   "test.jpg",
			file:       strings.NewReader("fake image data"),
			wantErr:    true,
			errMsg:     "record ID is required",
		},
		{
			name:       "빈 필드 이름",
			collection: "posts",
			recordID:   "record123",
			fieldName:  "",
			filename:   "test.jpg",
			file:       strings.NewReader("fake image data"),
			wantErr:    true,
			errMsg:     "field name is required",
		},
		{
			name:       "빈 파일명",
			collection: "posts",
			recordID:   "record123",
			fieldName:  "image",
			filename:   "",
			file:       strings.NewReader("fake image data"),
			wantErr:    true,
			errMsg:     "filename is required",
		},
		{
			name:       "nil 파일",
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
					// 파라미터 검증 에러는 서버에 도달하기 전에 발생
					return
				}

				// 성공적인 응답 시뮬레이션
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
			name:       "성공적인 파일 다운로드",
			collection: "posts",
			recordID:   "record123",
			filename:   "test.jpg",
			wantErr:    false,
		},
		{
			name:       "빈 컬렉션 이름",
			collection: "",
			recordID:   "record123",
			filename:   "test.jpg",
			wantErr:    true,
			errMsg:     "collection name is required",
		},
		{
			name:       "빈 레코드 ID",
			collection: "posts",
			recordID:   "",
			filename:   "test.jpg",
			wantErr:    true,
			errMsg:     "record ID is required",
		},
		{
			name:       "빈 파일명",
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
					// 파라미터 검증 에러는 서버에 도달하기 전에 발생
					return
				}

				// 성공적인 파일 응답 시뮬레이션
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

			// 파일 내용 읽기 테스트
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
			name:       "기본 파일 URL",
			collection: "posts",
			recordID:   "record123",
			filename:   "test.jpg",
			opts:       nil,
			expected:   "https://example.com/api/files/posts/record123/test.jpg",
		},
		{
			name:       "썸네일 옵션 포함",
			collection: "posts",
			recordID:   "record123",
			filename:   "test.jpg",
			opts:       &FileDownloadOptions{Thumb: "100x100"},
			expected:   "https://example.com/api/files/posts/record123/test.jpg?thumb=100x100",
		},
		{
			name:       "다운로드 옵션 포함",
			collection: "posts",
			recordID:   "record123",
			filename:   "test.jpg",
			opts:       &FileDownloadOptions{Download: true},
			expected:   "https://example.com/api/files/posts/record123/test.jpg?download=1",
		},
		{
			name:       "모든 옵션 포함",
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
			name:       "빈 컬렉션 이름",
			collection: "",
			recordID:   "record123",
			fieldName:  "image",
			filename:   "test.jpg",
			wantErr:    true,
			errMsg:     "collection name is required",
		},
		{
			name:       "빈 레코드 ID",
			collection: "posts",
			recordID:   "",
			fieldName:  "image",
			filename:   "test.jpg",
			wantErr:    true,
			errMsg:     "record ID is required",
		},
		{
			name:       "빈 필드 이름",
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

// 벤치마크 테스트
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
	fileData := bytes.NewReader(make([]byte, 1024)) // 1KB 테스트 파일

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		fileData.Seek(0, 0) // 파일 포인터 리셋
		_, err := client.Files.Upload(ctx, "posts", "record123", "image", "test.jpg", fileData)
		if err != nil {
			b.Fatalf("Upload failed: %v", err)
		}
	}
}
