package pocketbase

import (
	"context"
	"fmt"
	"net/http"
	"net/url"
)

// Model is the constraint for type-safe generic collection services.
//
// It is an alias of Mappable to keep the API consistent with the existing RecordService
// ToMap() behavior.
type Model = Mappable

// CollectionResult represents a collection list response with typed items.
//
// Note: this intentionally duplicates the pagination fields from ListResult to avoid
// JSON tag conflicts with ListResult.Items.
type CollectionResult[T Model] struct {
	Page       int `json:"page"`
	PerPage    int `json:"perPage"`
	TotalItems int `json:"totalItems"`
	TotalPages int `json:"totalPages"`
	Items      []T `json:"items"`
}

// Service provides type-safe CRUD operations for a single PocketBase collection.
type Service[T Model] struct {
	client         *Client
	collectionName string
	newModel       func() T
}

// NewService creates a new generic service for the specified collection.
func NewService[T Model](client *Client, collectionName string, newModel func() T) *Service[T] {
	return &Service[T]{
		client:         client,
		collectionName: collectionName,
		newModel:       newModel,
	}
}

// GetOne fetches a single record by its ID.
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
	if err := s.client.Send(ctx, http.MethodGet, path, nil, &record); err != nil {
		return zero, fmt.Errorf("pocketbase: fetch %s: %w", s.collectionName, err)
	}
	return record, nil
}

// GetList fetches a list of records.
func (s *Service[T]) GetList(ctx context.Context, opts *ListOptions) (*CollectionResult[T], error) {
	path := fmt.Sprintf("/api/collections/%s/records", url.PathEscape(s.collectionName))
	q := url.Values{}
	applyListOptions(q, opts)
	if qs := q.Encode(); qs != "" {
		path += "?" + qs
	}

	var result CollectionResult[T]
	if err := s.client.Send(ctx, http.MethodGet, path, nil, &result); err != nil {
		return nil, fmt.Errorf("pocketbase: fetch %s list: %w", s.collectionName, err)
	}
	return &result, nil
}

// Create creates a new record.
func (s *Service[T]) Create(ctx context.Context, record T, opts *WriteOptions) (T, error) {
	var zero T

	path := fmt.Sprintf("/api/collections/%s/records", url.PathEscape(s.collectionName))
	q := url.Values{}
	opts.apply(q)
	if qs := q.Encode(); qs != "" {
		path += "?" + qs
	}

	result := s.newModel()
	if err := s.client.Send(ctx, http.MethodPost, path, record.ToMap(), &result); err != nil {
		return zero, fmt.Errorf("pocketbase: create %s: %w", s.collectionName, err)
	}
	return result, nil
}

// Update updates an existing record.
func (s *Service[T]) Update(ctx context.Context, id string, record T, opts *WriteOptions) (T, error) {
	var zero T

	path := fmt.Sprintf("/api/collections/%s/records/%s", url.PathEscape(s.collectionName), url.PathEscape(id))
	q := url.Values{}
	opts.apply(q)
	if qs := q.Encode(); qs != "" {
		path += "?" + qs
	}

	result := s.newModel()
	if err := s.client.Send(ctx, http.MethodPatch, path, record.ToMap(), &result); err != nil {
		return zero, fmt.Errorf("pocketbase: update %s: %w", s.collectionName, err)
	}
	return result, nil
}

// Delete deletes a record by its ID.
func (s *Service[T]) Delete(ctx context.Context, id string) error {
	path := fmt.Sprintf("/api/collections/%s/records/%s", url.PathEscape(s.collectionName), url.PathEscape(id))
	if err := s.client.Send(ctx, http.MethodDelete, path, nil, nil); err != nil {
		return fmt.Errorf("pocketbase: delete %s: %w", s.collectionName, err)
	}
	return nil
}
