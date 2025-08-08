package pocketbase

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"mime/multipart"
	"net/http"
	"net/url"
	"strings"
)

// FileServiceAPI defines the interface for file operations.
type FileServiceAPI interface {
	Upload(ctx context.Context, collection, recordID, fieldName, filename string, file io.Reader) (*Record, error)
	Download(ctx context.Context, collection, recordID, filename string, opts *FileDownloadOptions) (io.ReadCloser, error)
	GetFileURL(collection, recordID, filename string, opts *FileDownloadOptions) string
	Delete(ctx context.Context, collection, recordID, fieldName, filename string) (*Record, error)
}

// FileService handles file operations for PocketBase records.
type FileService struct {
	Client *Client
}

// Upload uploads a file to a specific field of a record.
// Returns the updated record with the file information.
func (s *FileService) Upload(ctx context.Context, collection, recordID, fieldName, filename string, file io.Reader) (*Record, error) {
	if collection == "" {
		return nil, fmt.Errorf("collection name is required")
	}
	if recordID == "" {
		return nil, fmt.Errorf("record ID is required")
	}
	if fieldName == "" {
		return nil, fmt.Errorf("field name is required")
	}
	if filename == "" {
		return nil, fmt.Errorf("filename is required")
	}
	if file == nil {
		return nil, fmt.Errorf("file reader is required")
	}

	// Create multipart form data
	var buf bytes.Buffer
	writer := multipart.NewWriter(&buf)

	// Create form file field
	part, err := writer.CreateFormFile(fieldName, filename)
	if err != nil {
		return nil, fmt.Errorf("failed to create form file: %w", err)
	}

	// Copy file content to form
	if _, err := io.Copy(part, file); err != nil {
		return nil, fmt.Errorf("failed to copy file content: %w", err)
	}

	// Close the writer to finalize the form
	if err := writer.Close(); err != nil {
		return nil, fmt.Errorf("failed to close multipart writer: %w", err)
	}

	// Build the API path
	path := fmt.Sprintf("/api/collections/%s/records/%s", url.PathEscape(collection), url.PathEscape(recordID))

	// Send the request
	var result Record
	err = s.Client.do(ctx, http.MethodPatch, path, &buf, writer.FormDataContentType(), &result)
	if err != nil {
		return nil, fmt.Errorf("failed to upload file: %w", err)
	}

	return &result, nil
}

// Download downloads a file from a record.
// Returns an io.ReadCloser that must be closed by the caller.
func (s *FileService) Download(ctx context.Context, collection, recordID, filename string, opts *FileDownloadOptions) (io.ReadCloser, error) {
	if collection == "" {
		return nil, fmt.Errorf("collection name is required")
	}
	if recordID == "" {
		return nil, fmt.Errorf("record ID is required")
	}
	if filename == "" {
		return nil, fmt.Errorf("filename is required")
	}

	// Build the file URL
	fileURL := s.GetFileURL(collection, recordID, filename, opts)

	// Parse the URL to get the path
	parsedURL, err := url.Parse(fileURL)
	if err != nil {
		return nil, fmt.Errorf("failed to parse file URL: %w", err)
	}

	// Use sendStream to get the file content
	return s.Client.sendStream(ctx, http.MethodGet, parsedURL.Path+"?"+parsedURL.RawQuery, nil, "")
}

// GetFileURL generates the URL for accessing a file.
func (s *FileService) GetFileURL(collection, recordID, filename string, opts *FileDownloadOptions) string {
	baseURL := strings.TrimSuffix(s.Client.BaseURL, "/")
	path := fmt.Sprintf("/api/files/%s/%s/%s",
		url.PathEscape(collection),
		url.PathEscape(recordID),
		url.PathEscape(filename))

	fileURL := baseURL + path

	if opts != nil {
		params := url.Values{}
		if opts.Thumb != "" {
			params.Set("thumb", opts.Thumb)
		}
		if opts.Download {
			params.Set("download", "1")
		}
		if len(params) > 0 {
			fileURL += "?" + params.Encode()
		}
	}

	return fileURL
}

// Delete removes a file from a record field.
// This is done by updating the record and removing the file from the specified field.
func (s *FileService) Delete(ctx context.Context, collection, recordID, fieldName, filename string) (*Record, error) {
	if collection == "" {
		return nil, fmt.Errorf("collection name is required")
	}
	if recordID == "" {
		return nil, fmt.Errorf("record ID is required")
	}
	if fieldName == "" {
		return nil, fmt.Errorf("field name is required")
	}

	// First, get the current record to see the current file list
	record, err := s.Client.Records.GetOne(ctx, collection, recordID, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to get current record: %w", err)
	}

	// Get the current value of the file field
	currentValue := record.Get(fieldName)
	if currentValue == nil {
		return record, nil // Field doesn't exist, nothing to delete
	}

	var updatedValue any

	// Handle different field types (single file vs multiple files)
	switch v := currentValue.(type) {
	case string:
		// Single file field
		if v == filename {
			updatedValue = "" // Remove the file
		} else {
			return record, nil // File not found in this field
		}
	case []any:
		// Multiple files field
		var newFiles []string
		found := false
		for _, item := range v {
			if str, ok := item.(string); ok && str != filename {
				newFiles = append(newFiles, str)
			} else if str == filename {
				found = true
			}
		}
		if !found {
			return record, nil // File not found in this field
		}
		updatedValue = newFiles
	default:
		return nil, fmt.Errorf("unsupported field type for file deletion")
	}

	// Update the record with the new file list
	updateData := map[string]interface{}{
		fieldName: updatedValue,
	}

	return s.Client.Records.Update(ctx, collection, recordID, updateData)
}
