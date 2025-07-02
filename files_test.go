package pocketbase

import (
	"context"
	"errors"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/goccy/go-json"
)

func TestFileServiceUpload(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			t.Fatalf("unexpected method: %s", r.Method)
		}
		if r.URL.Path != "/api/files/posts/1/image" {
			t.Fatalf("unexpected path: %s", r.URL.Path)
		}
		mr, err := r.MultipartReader()
		if err != nil {
			t.Fatalf("multipart reader: %v", err)
		}
		part, err := mr.NextPart()
		if err != nil {
			t.Fatalf("next part: %v", err)
		}
		if part.FormName() != "image" {
			t.Fatalf("unexpected part name: %s", part.FormName())
		}
		data, err := io.ReadAll(part)
		if err != nil {
			t.Fatalf("failed to read part data: %v", err)
		}
		if string(data) != "data" {
			t.Fatalf("unexpected part data: %s", data)
		}
		if err := json.NewEncoder(w).Encode(Record{BaseModel: BaseModel{ID: "1"}}); err != nil {
			t.Fatalf("failed to write response: %v", err)
		}
	}))
	defer srv.Close()

	c := NewClient(srv.URL)
	rec, err := c.Files.Upload(context.Background(), "posts", "1", "image", "file.txt", strings.NewReader("data"))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if rec.ID != "1" {
		t.Fatalf("unexpected id: %s", rec.ID)
	}
}

func TestFileServiceDownload(t *testing.T) {
	const body = "filedata"
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			t.Fatalf("unexpected method: %s", r.Method)
		}
		if r.URL.Path != "/api/files/posts/1/image.png" {
			t.Fatalf("unexpected path: %s", r.URL.Path)
		}
		if r.URL.Query().Get("thumb") != "100x100" {
			t.Fatalf("unexpected thumb: %s", r.URL.Query().Get("thumb"))
		}
		if r.URL.Query().Get("download") != "1" {
			t.Fatalf("unexpected download flag: %s", r.URL.Query().Get("download"))
		}
		_, _ = w.Write([]byte(body))
	}))
	defer srv.Close()

	c := NewClient(srv.URL)
	rc, err := c.Files.Download(context.Background(), "posts", "1", "image.png", &FileDownloadOptions{Thumb: "100x100", Download: true})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	data, err := io.ReadAll(rc)
	if err != nil {
		t.Fatalf("read body: %v", err)
	}
	rc.Close()
	if string(data) != body {
		t.Fatalf("unexpected body: %s", data)
	}
}

func TestFileServiceGetProtectedFileToken(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			t.Fatalf("unexpected method: %s", r.Method)
		}
		if r.URL.Path != "/api/files/token" {
			t.Fatalf("unexpected path: %s", r.URL.Path)
		}
		_ = json.NewEncoder(w).Encode(map[string]string{"token": "abc"})
	}))
	defer srv.Close()

	c := NewClient(srv.URL)
	token, err := c.Files.GetProtectedFileToken(context.Background())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if token != "abc" {
		t.Fatalf("unexpected token: %s", token)
	}
}

func TestFileServiceDownloadError(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusBadRequest)
		_ = json.NewEncoder(w).Encode(APIError{Code: 400, Message: "bad"})
	}))
	defer srv.Close()

	c := NewClient(srv.URL)
	rc, err := c.Files.Download(context.Background(), "posts", "1", "f.txt", nil)
	if err == nil {
		t.Fatal("expected error")
	}
	if rc != nil {
		t.Fatal("expected nil reader")
	}
	var apiErr *APIError
	if !errors.As(err, &apiErr) {
		t.Fatalf("unexpected error type: %T", err)
	}
	if apiErr.Code != 400 || apiErr.Message != "bad" {
		t.Fatalf("unexpected api error: %+v", apiErr)
	}
}
