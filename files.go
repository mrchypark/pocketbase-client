package pocketbase

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"mime/multipart"
	"net/http"
	"net/url"
)

// FileServiceAPI defines file upload and download operations.
type FileServiceAPI interface {
	Upload(ctx context.Context, collection, recordID, field, filename string, r io.Reader) (*Record, error)
	Download(ctx context.Context, collection, recordID, filename string, opts *FileDownloadOptions) (io.ReadCloser, error)
	GetProtectedFileToken(ctx context.Context) (string, error)
}

// FileService handles file uploads and URL generation.
type FileService struct {
	Client *Client
}

var _ FileServiceAPI = (*FileService)(nil)

// Upload uploads a file to the specified record field in a collection.
func (s *FileService) Upload(ctx context.Context, collection, recordID, field, filename string, r io.Reader) (*Record, error) {
	path := fmt.Sprintf("/api/files/%s/%s", url.PathEscape(collection), url.PathEscape(recordID))

	var buf bytes.Buffer
	mw := multipart.NewWriter(&buf)
	part, err := mw.CreateFormFile(field, filename)
	if err != nil {
		return nil, fmt.Errorf("pocketbase: create form file: %w", err)
	}
	if _, err := io.Copy(part, r); err != nil {
		return nil, fmt.Errorf("pocketbase: copy file: %w", err)
	}
	if err := mw.Close(); err != nil {
		return nil, fmt.Errorf("pocketbase: close writer: %w", err)
	}

	var rec Record
	if err := s.Client.do(ctx, http.MethodPost, path, &buf, mw.FormDataContentType(), &rec); err != nil {
		return nil, fmt.Errorf("pocketbase: upload file: %w", err)
	}
	return &rec, nil
}

// Download downloads a file and returns it as an io.ReadCloser.
// The returned ReadCloser must be closed by the caller.
func (s *FileService) Download(ctx context.Context, collection, recordID, filename string, opts *FileDownloadOptions) (io.ReadCloser, error) {
	path := fmt.Sprintf("/api/files/%s/%s/%s", url.PathEscape(collection), url.PathEscape(recordID), url.PathEscape(filename))
	q := url.Values{}
	if opts != nil {
		if opts.Thumb != "" {
			q.Set("thumb", opts.Thumb)
		}
		if opts.Download {
			q.Set("download", "1")
		}
	}
	if qs := q.Encode(); qs != "" {
		path += "?" + qs
	}
	rc, err := s.Client.sendStream(ctx, http.MethodGet, path, nil, "")
	if err != nil {
		return nil, fmt.Errorf("pocketbase: download file: %w", err)
	}
	return rc, nil
}

// GetProtectedFileToken returns a temporary token for accessing protected files.
func (s *FileService) GetProtectedFileToken(ctx context.Context) (string, error) {
	var resp struct {
		Token string `json:"token"`
	}
	if err := s.Client.send(ctx, http.MethodPost, "/api/files/token", nil, &resp); err != nil {
		return "", fmt.Errorf("pocketbase: get protected file token: %w", err)
	}
	return resp.Token, nil
}
