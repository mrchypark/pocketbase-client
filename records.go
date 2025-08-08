package pocketbase

import (
	"context"
	"fmt"
	"net/http"
	"net/url"

	"github.com/goccy/go-json"
)

// RecordServiceAPI defines the interface for generic record service operations.
type RecordServiceAPI[T any] interface {
	GetList(ctx context.Context, opts *ListOptions) (*ListResultAs[T], error)
	GetAll(ctx context.Context, opts *ListOptions) (*ListResultAs[T], error)
	GetOne(ctx context.Context, recordID string, opts *GetOneOptions) (*T, error)
	Create(ctx context.Context, model *T) (*T, error)
	CreateWithOptions(ctx context.Context, model *T, opts *WriteOptions) (*T, error)
	Update(ctx context.Context, recordID string, model *T) (*T, error)
	UpdateWithOptions(ctx context.Context, recordID string, model *T, opts *WriteOptions) (*T, error)
	Delete(ctx context.Context, recordID string) error
	NewCreateRequest(model *T) (*BatchRequest, error)
	NewUpdateRequest(recordID string, model *T) (*BatchRequest, error)
	NewDeleteRequest(recordID string) (*BatchRequest, error)
	NewUpsertRequest(model *T) (*BatchRequest, error)
}

// RecordService provides CRUD operations for a specific type T.
type RecordService[T any] struct {
	client         *Client
	collectionName string
}

// NewRecordService creates a new generic service for a specific type T and collection name.
func NewRecordService[T any](client *Client, collectionName string) RecordServiceAPI[T] {
	return &RecordService[T]{
		client:         client,
		collectionName: collectionName,
	}
}

// RecordService[T] implements RecordServiceAPI[T] interface - compile time check
var _ RecordServiceAPI[any] = (*RecordService[any])(nil)

// GetList retrieves a list of records from the collection as generic type T.
func (s *RecordService[T]) GetList(ctx context.Context, opts *ListOptions) (*ListResultAs[T], error) {
	basePath := fmt.Sprintf("/api/collections/%s/records", url.PathEscape(s.collectionName))
	path := buildPathWithQuery(basePath, buildQueryString(opts))

	var result ListResultAs[T]
	if err := s.client.send(ctx, http.MethodGet, path, nil, &result); err != nil {
		return nil, err
	}

	return &result, nil
}

// GetAll retrieves all records from the collection by automatically paginating through all pages.
// It returns a single consolidated ListResultAs.
//
// Note: This method overrides the 'Page' and 'PerPage' options in the provided ListOptions.
// It fetches records in batches of 500 (the maximum allowed) for efficiency.
// The returned ListResultAs will have Page=1, TotalPages=1, and PerPage set to the total number of items.
func (s *RecordService[T]) GetAll(ctx context.Context, opts *ListOptions) (*ListResultAs[T], error) {
	// Initialize options with default values if nil
	if opts == nil {
		opts = &ListOptions{}
	}

	// Configure options for GetAll operation
	allOpts := *opts // Create a copy
	allOpts.Page = 1
	allOpts.PerPage = 500
	allOpts.SkipTotal = true

	// Request the first page
	result, err := s.GetList(ctx, &allOpts)
	if err != nil {
		return nil, err
	}

	// If first page has less than 500 items, we have all data
	if len(result.Items) < 500 {
		result.TotalItems = len(result.Items)
		result.TotalPages = 1
		return result, nil
	}

	// Slice to store all items
	allItems := make([]*T, 0, len(result.Items))
	allItems = append(allItems, result.Items...)

	// Request subsequent pages sequentially
	for page := 2; ; page++ {
		allOpts.Page = page
		pageResult, err := s.GetList(ctx, &allOpts)
		if err != nil {
			return nil, err
		}

		// Add items to the complete result
		allItems = append(allItems, pageResult.Items...)

		// If less than 500 items, this is the last page
		if len(pageResult.Items) < 500 {
			break
		}
	}

	// Construct final result
	finalResult := &ListResultAs[T]{
		Page:       1,
		PerPage:    len(allItems),
		TotalItems: len(allItems), // Set to actual count since skipTotal=true
		TotalPages: 1,
		Items:      allItems,
	}

	return finalResult, nil
}

// GetOne retrieves a single record as generic type T.
func (s *RecordService[T]) GetOne(ctx context.Context, recordID string, opts *GetOneOptions) (*T, error) {
	basePath := fmt.Sprintf("/api/collections/%s/records/%s", url.PathEscape(s.collectionName), url.PathEscape(recordID))
	q := url.Values{}
	if opts != nil {
		if opts.Expand != "" {
			q.Set("expand", opts.Expand)
		}
		if opts.Fields != "" {
			q.Set("fields", opts.Fields)
		}
	}
	path := buildPathWithQuery(basePath, q.Encode())

	var result T
	if err := s.client.send(ctx, http.MethodGet, path, nil, &result); err != nil {
		return nil, err
	}

	return &result, nil
}

// Create creates a new record with generic type T.
func (s *RecordService[T]) Create(ctx context.Context, model *T) (*T, error) {
	return s.CreateWithOptions(ctx, model, nil)
}

// CreateWithOptions creates a new record with generic type T and options.
func (s *RecordService[T]) CreateWithOptions(ctx context.Context, model *T, opts *WriteOptions) (*T, error) {
	basePath := fmt.Sprintf("/api/collections/%s/records", url.PathEscape(s.collectionName))
	q := url.Values{}
	if opts != nil {
		opts.apply(q)
	}
	path := buildPathWithQuery(basePath, q.Encode())

	var result T
	if err := s.client.send(ctx, http.MethodPost, path, model, &result); err != nil {
		return nil, err
	}

	return &result, nil
}

// Update updates an existing record with generic type T.
func (s *RecordService[T]) Update(ctx context.Context, recordID string, model *T) (*T, error) {
	return s.UpdateWithOptions(ctx, recordID, model, nil)
}

// UpdateWithOptions updates an existing record with generic type T and options.
func (s *RecordService[T]) UpdateWithOptions(ctx context.Context, recordID string, model *T, opts *WriteOptions) (*T, error) {
	basePath := fmt.Sprintf("/api/collections/%s/records/%s", url.PathEscape(s.collectionName), url.PathEscape(recordID))
	q := url.Values{}
	if opts != nil {
		opts.apply(q)
	}
	path := buildPathWithQuery(basePath, q.Encode())

	var result T
	if err := s.client.send(ctx, http.MethodPatch, path, model, &result); err != nil {
		return nil, err
	}

	return &result, nil
}

// Delete deletes a record from the collection.
func (s *RecordService[T]) Delete(ctx context.Context, recordID string) error {
	path := fmt.Sprintf("/api/collections/%s/records/%s", url.PathEscape(s.collectionName), url.PathEscape(recordID))
	if err := s.client.send(ctx, http.MethodDelete, path, nil, nil); err != nil {
		return err
	}
	return nil
}

// NewCreateRequest creates a new batch request for creating a record.
func (s *RecordService[T]) NewCreateRequest(model *T) (*BatchRequest, error) {
	body, err := modelToMap(model)
	if err != nil {
		return nil, err
	}
	return &BatchRequest{
		Method: http.MethodPost,
		URL:    fmt.Sprintf("/api/collections/%s/records", url.PathEscape(s.collectionName)),
		Body:   body,
	}, nil
}

// NewUpdateRequest creates a new batch request for updating a record.
func (s *RecordService[T]) NewUpdateRequest(recordID string, model *T) (*BatchRequest, error) {
	body, err := modelToMap(model)
	if err != nil {
		return nil, err
	}

	return &BatchRequest{
		Method: http.MethodPatch,
		URL:    fmt.Sprintf("/api/collections/%s/records/%s", url.PathEscape(s.collectionName), url.PathEscape(recordID)),
		Body:   body,
	}, nil
}

// NewDeleteRequest creates a new batch request for deleting a record.
func (s *RecordService[T]) NewDeleteRequest(recordID string) (*BatchRequest, error) {
	return &BatchRequest{
		Method: http.MethodDelete,
		URL:    fmt.Sprintf("/api/collections/%s/records/%s", url.PathEscape(s.collectionName), url.PathEscape(recordID)),
	}, nil
}

// NewUpsertRequest creates a new batch request for upserting a record.
func (s *RecordService[T]) NewUpsertRequest(model *T) (*BatchRequest, error) {
	body, err := modelToMap(model)
	if err != nil {
		return nil, err
	}

	// Check if id exists and is not empty
	id, exists := body["id"]
	if !exists || id == nil || id == "" {
		return nil, fmt.Errorf("upsert error: 'id' field is required in the body")
	}
	return &BatchRequest{
		Method: http.MethodPut,
		URL:    fmt.Sprintf("/api/collections/%s/records", url.PathEscape(s.collectionName)),
		Body:   body,
	}, nil
}

func modelToMap[T any](model *T) (map[string]any, error) {
	// The standard json.Marshal function handles everything.
	bytes, err := json.Marshal(model)
	if err != nil {
		return nil, err
	}
	var body map[string]any
	if err := json.Unmarshal(bytes, &body); err != nil {
		return nil, err
	}
	return body, nil
}
