package pocketbase

import (
	"context"
	"fmt"
	"net/http"
	"net/url"
)

// RecordServiceAPI defines the API operations related to records.
type RecordServiceAPI interface {
	GetList(ctx context.Context, collection string, opts *ListOptions) (*ListResult, error)
	GetOne(ctx context.Context, collection, recordID string, opts *GetOneOptions) (*Record, error)
	Create(ctx context.Context, collection string, body any) (*Record, error)
	CreateWithOptions(ctx context.Context, collection string, body any, opts *WriteOptions) (*Record, error)
	Update(ctx context.Context, collection, recordID string, body any) (*Record, error)
	UpdateWithOptions(ctx context.Context, collection, recordID string, body any, opts *WriteOptions) (*Record, error)
	Delete(ctx context.Context, collection, recordID string) error
	NewCreateRequest(collection string, body map[string]any) (*BatchRequest, error)
	NewUpdateRequest(collection, recordID string, body map[string]any) (*BatchRequest, error)
	NewDeleteRequest(collection, recordID string) (*BatchRequest, error)
	NewUpsertRequest(collection string, body map[string]any) (*BatchRequest, error)
}

// Mappable interface allows types to convert themselves to map[string]any.
type Mappable interface {
	ToMap() map[string]any
}

// RecordService handles record-related API operations.
type RecordService struct {
	Client *Client
}

var _ RecordServiceAPI = (*RecordService)(nil)

// GetList retrieves a list of records from a collection.
func (s *RecordService) GetList(ctx context.Context, collection string, opts *ListOptions) (*ListResult, error) {
	path := fmt.Sprintf("/api/collections/%s/records", url.PathEscape(collection))
	q := url.Values{}
	applyListOptions(q, opts)
	if qs := q.Encode(); qs != "" {
		path += "?" + qs
	}
	var result ListResult
	if err := s.Client.send(ctx, http.MethodGet, path, nil, &result); err != nil {
		return nil, fmt.Errorf("pocketbase: fetch records list: %w", err)
	}
	return &result, nil
}

func GetListAs[T any](ctx context.Context, client *Client, collection string, opts *ListOptions) (*ListResultAs[T], error) {
	path := fmt.Sprintf("/api/collections/%s/records", url.PathEscape(collection))
	q := url.Values{}
	applyListOptions(q, opts)
	if qs := q.Encode(); qs != "" {
		path += "?" + qs
	}

	var result ListResultAs[T]
	if err := client.send(ctx, http.MethodGet, path, nil, &result); err != nil {
		return nil, err
	}
	return &result, nil
}

// GetOne retrieves a single record.
func (s *RecordService) GetOne(ctx context.Context, collection, recordID string, opts *GetOneOptions) (*Record, error) {
	path := fmt.Sprintf("/api/collections/%s/records/%s", url.PathEscape(collection), url.PathEscape(recordID))
	q := url.Values{}
	if opts != nil {
		if opts.Expand != "" {
			q.Set("expand", opts.Expand)
		}
		if opts.Fields != "" {
			q.Set("fields", opts.Fields)
		}
	}
	if qs := q.Encode(); qs != "" {
		path += "?" + qs
	}
	var rec Record
	if err := s.Client.send(ctx, http.MethodGet, path, nil, &rec); err != nil {
		return nil, fmt.Errorf("pocketbase: fetch record: %w", err)
	}
	return &rec, nil
}

func GetOneAs[T any](ctx context.Context, client *Client, collection, recordID string, opts *GetOneOptions) (*T, error) {
	// API 경로 구성
	path := fmt.Sprintf("/api/collections/%s/records/%s", url.PathEscape(collection), url.PathEscape(recordID))
	q := url.Values{}
	if opts != nil {
		if opts.Expand != "" {
			q.Set("expand", opts.Expand)
		}
		if opts.Fields != "" {
			q.Set("fields", opts.Fields)
		}
	}
	if qs := q.Encode(); qs != "" {
		path += "?" + qs
	}
	var result T
	if err := client.send(ctx, "GET", path, nil, &result); err != nil {
		return nil, err
	}
	return &rec, nil
}

// Create creates a new record in the specified collection.
func (s *RecordService) Create(ctx context.Context, collection string, body any) (*Record, error) {
	return s.CreateWithOptions(ctx, collection, body, nil)
}

// CreateWithOptions creates a new record in the specified collection with additional options.
func (s *RecordService) CreateWithOptions(ctx context.Context, collection string, body any, opts *WriteOptions) (*Record, error) {
	path := fmt.Sprintf("/api/collections/%s/records", url.PathEscape(collection))
	q := url.Values{}
	if opts != nil {
		if opts.Expand != "" {
			q.Set("expand", opts.Expand)
		}
		if opts.Fields != "" {
			q.Set("fields", opts.Fields)
		}
	}
	if qs := q.Encode(); qs != "" {
		path += "?" + qs
	}
	requestBody := body
	if mappable, ok := body.(Mappable); ok {
		// If implemented, call ToMap() to convert to map
		requestBody = mappable.ToMap()
	}

	var rec Record
	if err := s.Client.send(ctx, http.MethodPost, path, requestBody, &rec); err != nil {
		return nil, fmt.Errorf("pocketbase: create record: %w", err)
	}
	return &rec, nil
}

func CreateAs[T any](ctx context.Context, s *RecordService, collection string, body *T, opts *WriteOptions) (*T, error) {
	path := fmt.Sprintf("/api/collections/%s/records", url.PathEscape(collection))
	q := url.Values{}
	if opts != nil {
		opts.apply(q)
	}
	if qs := q.Encode(); qs != "" {
		path += "?" + qs
	}

	var result T
	if err := s.Client.send(ctx, http.MethodPost, path, body, &result); err != nil {
		return nil, fmt.Errorf("pocketbase: create typed record: %w", err)
	}
	return &rec, nil
}

// Update updates an existing record in the specified collection.
func (s *RecordService) Update(ctx context.Context, collection, recordID string, body any) (*Record, error) {
	return s.UpdateWithOptions(ctx, collection, recordID, body, nil)
}

// UpdateWithOptions updates an existing record in the specified collection with additional options.
func (s *RecordService) UpdateWithOptions(ctx context.Context, collection, recordID string, body any, opts *WriteOptions) (*Record, error) {
	path := fmt.Sprintf("/api/collections/%s/records/%s", url.PathEscape(collection), url.PathEscape(recordID))
	q := url.Values{}
	if opts != nil {
		if opts.Expand != "" {
			q.Set("expand", opts.Expand)
		}
		if opts.Fields != "" {
			q.Set("fields", opts.Fields)
		}
	}
	if qs := q.Encode(); qs != "" {
		path += "?" + qs
	}
	requestBody := body
	if mappable, ok := body.(Mappable); ok {
		// If implemented, call ToMap() to convert to map
		requestBody = mappable.ToMap()
	}

	var rec Record
	if err := s.Client.send(ctx, http.MethodPatch, path, requestBody, &rec); err != nil {
		return nil, fmt.Errorf("pocketbase: update record: %w", err)
	}
	return &rec, nil
}

// UpdateAs updates an existing record with data from a typed struct `T`.
// It returns the updated record as a new instance of `*T`.
// This is the recommended way to update records for maximum type safety.
func UpdateAs[T any](ctx context.Context, s *RecordService, collection, recordID string, body *T, opts *WriteOptions) (*T, error) {
	path := fmt.Sprintf("/api/collections/%s/records/%s", url.PathEscape(collection), url.PathEscape(recordID))
	q := url.Values{}
	if opts != nil {
		opts.apply(q)
	}
	if qs := q.Encode(); qs != "" {
		path += "?" + qs
	}

	var result T
	if err := s.Client.send(ctx, http.MethodPatch, path, body, &result); err != nil {
		return nil, fmt.Errorf("pocketbase: update typed record: %w", err)
	}
	return &rec, nil
}

// Delete deletes a record from the specified collection.
func (s *RecordService) Delete(ctx context.Context, collection, recordID string) error {
	path := fmt.Sprintf("/api/collections/%s/records/%s", url.PathEscape(collection), url.PathEscape(recordID))
	if err := s.Client.send(ctx, http.MethodDelete, path, nil, nil); err != nil {
		return fmt.Errorf("pocketbase: delete record: %w", err)
	}
	return nil
}

// NewCreateRequest creates a new batch request for creating a record.
func (s *RecordService) NewCreateRequest(collection string, body map[string]any) (*BatchRequest, error) {
	return &BatchRequest{
		Method: http.MethodPost,
		URL:    fmt.Sprintf("/api/collections/%s/records", url.PathEscape(collection)),
		Body:   body,
	}, nil
}

// NewUpdateRequest creates a new batch request for updating a record.
func (s *RecordService) NewUpdateRequest(collection, recordID string, body map[string]any) (*BatchRequest, error) {
	return &BatchRequest{
		Method: http.MethodPatch,
		URL:    fmt.Sprintf("/api/collections/%s/records/%s", url.PathEscape(collection), url.PathEscape(recordID)),
		Body:   body,
	}, nil
}

// NewDeleteRequest creates a new batch request for deleting a record.
func (s *RecordService) NewDeleteRequest(collection, recordID string) (*BatchRequest, error) {
	return &BatchRequest{
		Method: http.MethodDelete,
		URL:    fmt.Sprintf("/api/collections/%s/records/%s", url.PathEscape(collection), url.PathEscape(recordID)),
	}, nil
}

// NewUpsertRequest creates a new batch request for upserting a record.
func (s *RecordService) NewUpsertRequest(collection string, body map[string]any) (*BatchRequest, error) {
	if _, ok := body["id"]; !ok {
		return nil, fmt.Errorf("upsert error: 'id' field is required in the body")
	}
	return &BatchRequest{
		Method: http.MethodPut,
		URL:    fmt.Sprintf("/api/collections/%s/records", url.PathEscape(collection)),
		Body:   body,
	}, nil
}
