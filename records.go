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
		requestBody = mappable.ToMap()
	}

	var rec Record
	if err := s.Client.send(ctx, http.MethodPost, path, requestBody, &rec); err != nil {
		return nil, fmt.Errorf("pocketbase: create record: %w", err)
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
		requestBody = mappable.ToMap()
	}

	var rec Record
	if err := s.Client.send(ctx, http.MethodPatch, path, requestBody, &rec); err != nil {
		return nil, fmt.Errorf("pocketbase: update record: %w", err)
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
		Body:   nil,
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

// TypedRecordService provides type-safe record operations.
// T must be a pointer type that embeds Record or implements compatible interface.
type TypedRecordService[T any] struct {
	*RecordService
	Collection string
}

// NewTypedRecordService creates a new TypedRecordService for the given collection.
func NewTypedRecordService[T any](client *Client, collection string) *TypedRecordService[T] {
	return &TypedRecordService[T]{
		RecordService: &RecordService{Client: client},
		Collection:    collection,
	}
}

// GetOne retrieves a single record and converts it to type T.
func (s *TypedRecordService[T]) GetOne(ctx context.Context, recordID string, opts *GetOneOptions) (*T, error) {
	rec, err := s.RecordService.GetOne(ctx, s.Collection, recordID, opts)
	if err != nil {
		return nil, err
	}
	return convertRecord[T](rec)
}

// Create creates a new record from type T.
func (s *TypedRecordService[T]) Create(ctx context.Context, body *T) (*T, error) {
	rec, err := s.RecordService.Create(ctx, s.Collection, body)
	if err != nil {
		return nil, err
	}
	return convertRecord[T](rec)
}

// Update updates an existing record.
func (s *TypedRecordService[T]) Update(ctx context.Context, recordID string, body *T) (*T, error) {
	rec, err := s.RecordService.Update(ctx, s.Collection, recordID, body)
	if err != nil {
		return nil, err
	}
	return convertRecord[T](rec)
}

// GetList retrieves a list of records.
func (s *TypedRecordService[T]) GetList(ctx context.Context, opts *ListOptions) (*TypedListResult[T], error) {
	res, err := s.RecordService.GetList(ctx, s.Collection, opts)
	if err != nil {
		return nil, err
	}
	items := make([]*T, len(res.Items))
	for i, rec := range res.Items {
		item, err := convertRecord[T](rec)
		if err != nil {
			return nil, err
		}
		items[i] = item
	}
	return &TypedListResult[T]{
		Page:       res.Page,
		PerPage:    res.PerPage,
		TotalItems: res.TotalItems,
		TotalPages: res.TotalPages,
		Items:      items,
	}, nil
}

// GetAll retrieves all records (auto-pagination).
func (s *TypedRecordService[T]) GetAll(ctx context.Context, opts *ListOptions) ([]*T, error) {
	var all []*T
	page := 1
	perPage := 100
	if opts != nil && opts.PerPage > 0 {
		perPage = opts.PerPage
	}

	for {
		res, err := s.GetList(ctx, &ListOptions{
			Page:      page,
			PerPage:   perPage,
			Filter:    opts.Filter,
			Sort:      opts.Sort,
			Expand:    opts.Expand,
			Fields:    opts.Fields,
			SkipTotal: opts.SkipTotal,
		})
		if err != nil {
			return nil, err
		}
		all = append(all, res.Items...)
		if page >= res.TotalPages {
			break
		}
		page++
	}
	return all, nil
}

// TypedListResult is a typed version of ListResult.
type TypedListResult[T any] struct {
	Page       int  `json:"page"`
	PerPage    int  `json:"perPage"`
	TotalItems int  `json:"totalItems"`
	TotalPages int  `json:"totalPages"`
	Items      []*T `json:"items"`
}

// convertRecord converts a Record to type T.
func convertRecord[T any](rec *Record) (*T, error) {
	var t T
	if ptr, ok := any(&t).(**Record); ok {
		*ptr = rec
		return &t, nil
	}
	return nil, fmt.Errorf("pocketbase: cannot convert Record to %T", t)
}
