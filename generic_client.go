package pocketbase

import (
	"context"
	"fmt"
	"net/url"
)

// Model represents any PocketBase model that embeds BaseModel
type Model interface {
	ToMap() map[string]any
}

// Collection represents a collection response with items
type CollectionResult[T Model] struct {
	*ListResult
	Items []T `json:"items"`
}

// Service provides type-safe operations for any PocketBase collection
type Service[T Model] struct {
	client         *Client
	collectionName string
	newModel       func() T
}

// NewService creates a new generic service for the specified collection
func NewService[T Model](client *Client, collectionName string, newModel func() T) *Service[T] {
	return &Service[T]{
		client:         client,
		collectionName: collectionName,
		newModel:       newModel,
	}
}

// GetOne fetches a single record by its ID
func (s *Service[T]) GetOne(ctx context.Context, id string, opts *GetOneOptions) (T, error) {
	var zero T

	path := fmt.Sprintf("/api/collections/%s/records/%s", url.PathEscape(s.collectionName), url.PathEscape(id))
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

	record := s.newModel()
	if err := s.client.Send(ctx, "GET", path, nil, &record); err != nil {
		return zero, fmt.Errorf("pocketbase: fetch %s: %w", s.collectionName, err)
	}
	return record, nil
}

// GetList fetches a list of records
func (s *Service[T]) GetList(ctx context.Context, opts *ListOptions) (*CollectionResult[T], error) {
	path := fmt.Sprintf("/api/collections/%s/records", url.PathEscape(s.collectionName))
	q := url.Values{}
	ApplyListOptions(q, opts)
	if qs := q.Encode(); qs != "" {
		path += "?" + qs
	}

	var result CollectionResult[T]
	if err := s.client.Send(ctx, "GET", path, nil, &result); err != nil {
		return nil, fmt.Errorf("pocketbase: fetch %s list: %w", s.collectionName, err)
	}
	return &result, nil
}

// Create creates a new record
func (s *Service[T]) Create(ctx context.Context, record T, opts *WriteOptions) (T, error) {
	var zero T

	path := fmt.Sprintf("/api/collections/%s/records", url.PathEscape(s.collectionName))
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

	result := s.newModel()
	if err := s.client.Send(ctx, "POST", path, record.ToMap(), &result); err != nil {
		return zero, fmt.Errorf("pocketbase: create %s: %w", s.collectionName, err)
	}
	return result, nil
}

// Update updates an existing record
func (s *Service[T]) Update(ctx context.Context, id string, record T, opts *WriteOptions) (T, error) {
	var zero T

	path := fmt.Sprintf("/api/collections/%s/records/%s", url.PathEscape(s.collectionName), url.PathEscape(id))
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

	result := s.newModel()
	if err := s.client.Send(ctx, "PATCH", path, record.ToMap(), &result); err != nil {
		return zero, fmt.Errorf("pocketbase: update %s: %w", s.collectionName, err)
	}
	return result, nil
}

// Delete deletes a record by its ID
func (s *Service[T]) Delete(ctx context.Context, id string) error {
	path := fmt.Sprintf("/api/collections/%s/records/%s", url.PathEscape(s.collectionName), url.PathEscape(id))
	if err := s.client.Send(ctx, "DELETE", path, nil, nil); err != nil {
		return fmt.Errorf("pocketbase: delete %s: %w", s.collectionName, err)
	}
	return nil
}
