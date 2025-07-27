package pocketbase

import (
	"context"
	"fmt"
	"net/http"
	"net/url"
)

// RecordServiceLegacyAPI defines the API operations related to records (legacy non-generic interface).
type RecordServiceLegacyAPI interface {
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

// RecordService provides CRUD operations for a specific type T.
type RecordService[T any] struct {
	client         *Client
	collectionName string
}

// RecordServiceAPI defines the interface for generic record service operations.
type RecordServiceAPI[T any] interface {
	GetList(ctx context.Context, opts *ListOptions) (*ListResultAs[T], error)
	GetOne(ctx context.Context, recordID string, opts *GetOneOptions) (*T, error)
	Create(ctx context.Context, body *T, opts *WriteOptions) (*T, error)
	Update(ctx context.Context, recordID string, body *T, opts *WriteOptions) (*T, error)
	Delete(ctx context.Context, recordID string) error
}

// Mappable interface allows types to convert themselves to map[string]any.
type Mappable interface {
	ToMap() map[string]any
}

// NewRecordService creates a new generic service for a specific type T and collection name.
func NewRecordService[T any](client *Client, collectionName string) *RecordService[T] {
	return &RecordService[T]{
		client:         client,
		collectionName: collectionName,
	}
}

// RecordServiceLegacy handles record-related API operations (legacy non-generic service).
type RecordServiceLegacy struct {
	Client *Client
}

var _ RecordServiceLegacyAPI = (*RecordServiceLegacy)(nil)

// RecordService[T] implements RecordServiceAPI[T] interface - compile time check
var _ RecordServiceAPI[any] = (*RecordService[any])(nil)

// GetList retrieves a list of records from a collection.
func (s *RecordServiceLegacy) GetList(ctx context.Context, collection string, opts *ListOptions) (*ListResult, error) {
	basePath := fmt.Sprintf("/api/collections/%s/records", url.PathEscape(collection))
	path := buildPathWithQuery(basePath, buildQueryString(opts))
	var result ListResult
	if err := s.Client.send(ctx, http.MethodGet, path, nil, &result); err != nil {
		return nil, wrapError("fetch", "records list", err)
	}
	return &result, nil
}

// GetOne retrieves a single record.
func (s *RecordServiceLegacy) GetOne(ctx context.Context, collection, recordID string, opts *GetOneOptions) (*Record, error) {
	basePath := fmt.Sprintf("/api/collections/%s/records/%s", url.PathEscape(collection), url.PathEscape(recordID))
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
	var rec Record
	if err := s.Client.send(ctx, http.MethodGet, path, nil, &rec); err != nil {
		return nil, wrapError("fetch", "record", err)
	}
	return &rec, nil
}

// Create creates a new record in the specified collection.
func (s *RecordServiceLegacy) Create(ctx context.Context, collection string, body any) (*Record, error) {
	return s.CreateWithOptions(ctx, collection, body, nil)
}

// CreateWithOptions creates a new record in the specified collection with additional options.
func (s *RecordServiceLegacy) CreateWithOptions(ctx context.Context, collection string, body any, opts *WriteOptions) (*Record, error) {
	basePath := fmt.Sprintf("/api/collections/%s/records", url.PathEscape(collection))
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

	requestBody := body
	if mappable, ok := body.(Mappable); ok {
		requestBody = mappable.ToMap()
	}

	var rec Record
	if err := s.Client.send(ctx, http.MethodPost, path, requestBody, &rec); err != nil {
		return nil, wrapError("create", "record", err)
	}
	return &rec, nil
}

// Update updates an existing record in the specified collection.
func (s *RecordServiceLegacy) Update(ctx context.Context, collection, recordID string, body any) (*Record, error) {
	return s.UpdateWithOptions(ctx, collection, recordID, body, nil)
}

// UpdateWithOptions updates an existing record in the specified collection with additional options.
func (s *RecordServiceLegacy) UpdateWithOptions(ctx context.Context, collection, recordID string, body any, opts *WriteOptions) (*Record, error) {
	basePath := fmt.Sprintf("/api/collections/%s/records/%s", url.PathEscape(collection), url.PathEscape(recordID))
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

	requestBody := body
	if mappable, ok := body.(Mappable); ok {
		requestBody = mappable.ToMap()
	}

	var rec Record
	if err := s.Client.send(ctx, http.MethodPatch, path, requestBody, &rec); err != nil {
		return nil, wrapError("update", "record", err)
	}
	return &rec, nil
}

// Delete deletes a record from the specified collection.
func (s *RecordServiceLegacy) Delete(ctx context.Context, collection, recordID string) error {
	path := fmt.Sprintf("/api/collections/%s/records/%s", url.PathEscape(collection), url.PathEscape(recordID))
	if err := s.Client.send(ctx, http.MethodDelete, path, nil, nil); err != nil {
		return wrapError("delete", "record", err)
	}
	return nil
}

// NewCreateRequest creates a new batch request for creating a record.
func (s *RecordServiceLegacy) NewCreateRequest(collection string, body map[string]any) (*BatchRequest, error) {
	return &BatchRequest{
		Method: http.MethodPost,
		URL:    fmt.Sprintf("/api/collections/%s/records", url.PathEscape(collection)),
		Body:   body,
	}, nil
}

// NewUpdateRequest creates a new batch request for updating a record.
func (s *RecordServiceLegacy) NewUpdateRequest(collection, recordID string, body map[string]any) (*BatchRequest, error) {
	return &BatchRequest{
		Method: http.MethodPatch,
		URL:    fmt.Sprintf("/api/collections/%s/records/%s", url.PathEscape(collection), url.PathEscape(recordID)),
		Body:   body,
	}, nil
}

// NewDeleteRequest creates a new batch request for deleting a record.
func (s *RecordServiceLegacy) NewDeleteRequest(collection, recordID string) (*BatchRequest, error) {
	return &BatchRequest{
		Method: http.MethodDelete,
		URL:    fmt.Sprintf("/api/collections/%s/records/%s", url.PathEscape(collection), url.PathEscape(recordID)),
	}, nil
}

// NewUpsertRequest creates a new batch request for upserting a record.
func (s *RecordServiceLegacy) NewUpsertRequest(collection string, body map[string]any) (*BatchRequest, error) {
	if _, ok := body["id"]; !ok {
		return nil, fmt.Errorf("upsert error: 'id' field is required in the body")
	}
	return &BatchRequest{
		Method: http.MethodPut,
		URL:    fmt.Sprintf("/api/collections/%s/records", url.PathEscape(collection)),
		Body:   body,
	}, nil
}

// RecordService[T] method implementations

// GetList retrieves a list of records from the collection as generic type T.
func (s *RecordService[T]) GetList(ctx context.Context, opts *ListOptions) (*ListResultAs[T], error) {
	basePath := fmt.Sprintf("/api/collections/%s/records", url.PathEscape(s.collectionName))
	path := buildPathWithQuery(basePath, buildQueryString(opts))

	var result ListResultAs[T]
	if err := s.client.send(ctx, http.MethodGet, path, nil, &result); err != nil {
		return nil, wrapError("fetch", "typed records list", err)
	}

	return &result, nil
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
		return nil, wrapError("fetch", "typed record", err)
	}

	return &result, nil
}

// Create creates a new record with generic type T.
func (s *RecordService[T]) Create(ctx context.Context, body *T, opts *WriteOptions) (*T, error) {
	basePath := fmt.Sprintf("/api/collections/%s/records", url.PathEscape(s.collectionName))
	q := url.Values{}
	if opts != nil {
		opts.apply(q)
	}
	path := buildPathWithQuery(basePath, q.Encode())

	var result T
	if err := s.client.send(ctx, http.MethodPost, path, body, &result); err != nil {
		return nil, wrapError("create", "typed record", err)
	}

	return &result, nil
}

// Update updates an existing record with generic type T.
func (s *RecordService[T]) Update(ctx context.Context, recordID string, body *T, opts *WriteOptions) (*T, error) {
	basePath := fmt.Sprintf("/api/collections/%s/records/%s", url.PathEscape(s.collectionName), url.PathEscape(recordID))
	q := url.Values{}
	if opts != nil {
		opts.apply(q)
	}
	path := buildPathWithQuery(basePath, q.Encode())

	var result T
	if err := s.client.send(ctx, http.MethodPatch, path, body, &result); err != nil {
		return nil, wrapError("update", "typed record", err)
	}

	return &result, nil
}

// Delete deletes a record from the collection.
func (s *RecordService[T]) Delete(ctx context.Context, recordID string) error {
	path := fmt.Sprintf("/api/collections/%s/records/%s", url.PathEscape(s.collectionName), url.PathEscape(recordID))
	if err := s.client.send(ctx, http.MethodDelete, path, nil, nil); err != nil {
		return wrapError("delete", "typed record", err)
	}
	return nil
}

// Deprecated functions for backward compatibility

// GetListAs retrieves a list of records from a collection as generic type T.
//
// Deprecated: GetListAs is deprecated. Use NewRecordService[T](client, collection).GetList() instead.
// This function will be removed in a future version.
func GetListAs[T any](ctx context.Context, client *Client, collection string, opts *ListOptions) (*ListResultAs[T], error) {
	service := NewRecordService[T](client, collection)
	return service.GetList(ctx, opts)
}

// GetOneAs retrieves a single record as generic type T.
//
// Deprecated: GetOneAs is deprecated. Use NewRecordService[T](client, collection).GetOne() instead.
// This function will be removed in a future version.
func GetOneAs[T any](ctx context.Context, client *Client, collection, recordID string, opts *GetOneOptions) (*T, error) {
	service := NewRecordService[T](client, collection)
	return service.GetOne(ctx, recordID, opts)
}

// CreateAs creates a new record with generic type T.
//
// Deprecated: CreateAs is deprecated. Use NewRecordService[T](client, collection).Create() instead.
// This function will be removed in a future version.
func CreateAs[T any](ctx context.Context, s *RecordServiceLegacy, collection string, body *T, opts *WriteOptions) (*T, error) {
	service := NewRecordService[T](s.Client, collection)
	return service.Create(ctx, body, opts)
}

// UpdateAs updates an existing record with data from a typed struct `T`.
//
// Deprecated: UpdateAs is deprecated. Use NewRecordService[T](client, collection).Update() instead.
// This function will be removed in a future version.
func UpdateAs[T any](ctx context.Context, s *RecordServiceLegacy, collection, recordID string, body *T, opts *WriteOptions) (*T, error) {
	service := NewRecordService[T](s.Client, collection)
	return service.Update(ctx, recordID, body, opts)
}
