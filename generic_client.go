package pocketbase

import (
	"context"
	"fmt"
	"net/http"
	"net/url"

	"github.com/goccy/go-json"
)

// TypedRecordService provides type-safe CRUD operations for a single PocketBase
// collection.
//
// Unlike the dynamic RecordService, responses are decoded directly from the raw
// HTTP response body into the concrete type T. This avoids the intermediate
// dynamic Record representation and its extra JSON round-trip, making the typed
// read path effectively a single decode pass.
//
// T must be a struct whose JSON tags match the collection's fields (generated
// models satisfy this). Create and Update bodies are serialized through ToMap()
// when T implements Mappable (generated models do), preserving PATCH semantics.
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

// GetOne retrieves a single record and decodes it directly into *T.
func (s *TypedRecordService[T]) GetOne(ctx context.Context, recordID string, opts *GetOneOptions) (*T, error) {
	path := fmt.Sprintf("/api/collections/%s/records/%s", url.PathEscape(s.Collection), url.PathEscape(recordID))
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

	data, err := s.Client.SendRaw(ctx, http.MethodGet, path, nil)
	if err != nil {
		return nil, fmt.Errorf("pocketbase: fetch %s: %w", s.Collection, err)
	}

	result := new(T)
	if err := json.Unmarshal(data, result); err != nil {
		return nil, fmt.Errorf("pocketbase: decode %T: %w", result, err)
	}
	return result, nil
}

// Create creates a new record from type T and decodes the response directly.
func (s *TypedRecordService[T]) Create(ctx context.Context, body *T) (*T, error) {
	path := fmt.Sprintf("/api/collections/%s/records", url.PathEscape(s.Collection))

	requestBody := prepareRequestBody(body)
	data, err := s.Client.SendRaw(ctx, http.MethodPost, path, requestBody)
	if err != nil {
		return nil, fmt.Errorf("pocketbase: create %s: %w", s.Collection, err)
	}

	result := new(T)
	if err := json.Unmarshal(data, result); err != nil {
		return nil, fmt.Errorf("pocketbase: decode %T: %w", result, err)
	}
	return result, nil
}

// Update updates an existing record and decodes the response directly.
func (s *TypedRecordService[T]) Update(ctx context.Context, recordID string, body *T) (*T, error) {
	path := fmt.Sprintf("/api/collections/%s/records/%s", url.PathEscape(s.Collection), url.PathEscape(recordID))

	requestBody := prepareRequestBody(body)
	data, err := s.Client.SendRaw(ctx, http.MethodPatch, path, requestBody)
	if err != nil {
		return nil, fmt.Errorf("pocketbase: update %s: %w", s.Collection, err)
	}

	result := new(T)
	if err := json.Unmarshal(data, result); err != nil {
		return nil, fmt.Errorf("pocketbase: decode %T: %w", result, err)
	}
	return result, nil
}

// GetList retrieves a list of records and decodes them directly into
// a TypedListResult[T].
func (s *TypedRecordService[T]) GetList(ctx context.Context, opts *ListOptions) (*TypedListResult[T], error) {
	path := fmt.Sprintf("/api/collections/%s/records", url.PathEscape(s.Collection))
	if qs := buildQueryString(opts); qs != "" {
		path += "?" + qs
	}

	data, err := s.Client.SendRaw(ctx, http.MethodGet, path, nil)
	if err != nil {
		return nil, fmt.Errorf("pocketbase: fetch %s list: %w", s.Collection, err)
	}

	var result TypedListResult[T]
	if err := json.Unmarshal(data, &result); err != nil {
		return nil, fmt.Errorf("pocketbase: decode %T: %w", result, err)
	}
	return &result, nil
}

// GetAll retrieves all records via automatic pagination.
func (s *TypedRecordService[T]) GetAll(ctx context.Context, opts *ListOptions) ([]*T, error) {
	base := ListOptions{}
	if opts != nil {
		base = *opts
	}
	if base.PerPage <= 0 {
		base.PerPage = 100
	}
	base.Page = 1

	var all []*T
	for {
		res, err := s.GetList(ctx, &base)
		if err != nil {
			return nil, err
		}
		all = append(all, res.Items...)
		if len(res.Items) == 0 || res.Page >= res.TotalPages {
			break
		}
		base.Page++
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

// prepareRequestBody serializes a typed body for Create/Update.
// Types implementing Mappable (generated models) are converted via ToMap() to
// preserve PATCH semantics (omitting empty optional fields). Other types are
// marshaled as-is.
func prepareRequestBody(body any) any {
	if m, ok := body.(Mappable); ok {
		return m.ToMap()
	}
	return body
}
