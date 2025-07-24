package pocketbase

import (
	"context"
	"fmt"
	"net/http"
	"net/url"
)

// CollectionServiceAPI defines the API operations for managing collections.
type CollectionServiceAPI interface {
	GetList(ctx context.Context, opts *ListOptions) (*CollectionListResult, error)
	GetOne(ctx context.Context, idOrName string) (*Collection, error)
	Create(ctx context.Context, col *Collection) (*Collection, error)
	Update(ctx context.Context, idOrName string, col *Collection) (*Collection, error)
	Delete(ctx context.Context, idOrName string) error
	Import(ctx context.Context, cols []*Collection, deleteMissing bool) ([]*Collection, error)
}

// Collection represents the schema of a PocketBase collection.
type Collection struct {
	BaseModel
	Name       string         `json:"name"`
	Type       string         `json:"type"`
	System     bool           `json:"system"`
	Schema     []SchemaField  `json:"schema"`
	ListRule   string         `json:"listRule"`
	ViewRule   string         `json:"viewRule"`
	CreateRule string         `json:"createRule"`
	UpdateRule string         `json:"updateRule"`
	DeleteRule string         `json:"deleteRule"`
	Indexes    string         `json:"indexes"`
	Options    map[string]any `json:"options"`
}

// SchemaField defines a collection field.
type SchemaField struct {
	Name        string         `json:"name"`
	Type        string         `json:"type"`
	Required    bool           `json:"required"`
	Presentable bool           `json:"presentable"`
	Unique      bool           `json:"unique"`
	Options     map[string]any `json:"options"`
}

// CollectionListResult contains the result of a collection list query.
type CollectionListResult struct {
	Page       int           `json:"page"`
	PerPage    int           `json:"perPage"`
	TotalItems int           `json:"totalItems"`
	TotalPages int           `json:"totalPages"`
	Items      []*Collection `json:"items"`
}

// CollectionService provides collection management API.
type CollectionService struct {
	Client *Client
}

var _ CollectionServiceAPI = (*CollectionService)(nil)

// GetList retrieves a list of collections.
func (s *CollectionService) GetList(ctx context.Context, opts *ListOptions) (*CollectionListResult, error) {
	path := "/api/collections"
	q := url.Values{}
	ApplyListOptions(q, opts)
	if qs := q.Encode(); qs != "" {
		path += "?" + qs
	}
	var res CollectionListResult
	if err := s.Client.send(ctx, http.MethodGet, path, nil, &res); err != nil {
		return nil, fmt.Errorf("pocketbase: fetch collections list: %w", err)
	}
	return &res, nil
}

// GetOne retrieves a specific collection.
func (s *CollectionService) GetOne(ctx context.Context, idOrName string) (*Collection, error) {
	path := fmt.Sprintf("/api/collections/%s", url.PathEscape(idOrName))
	var col Collection
	if err := s.Client.send(ctx, http.MethodGet, path, nil, &col); err != nil {
		return nil, fmt.Errorf("pocketbase: fetch collection: %w", err)
	}
	return &col, nil
}

// Create creates a new collection.
func (s *CollectionService) Create(ctx context.Context, col *Collection) (*Collection, error) {
	path := "/api/collections"
	var res Collection
	if err := s.Client.send(ctx, http.MethodPost, path, col, &res); err != nil {
		return nil, fmt.Errorf("pocketbase: create collection: %w", err)
	}
	return &res, nil
}

// Update modifies an existing collection.
func (s *CollectionService) Update(ctx context.Context, idOrName string, col *Collection) (*Collection, error) {
	path := fmt.Sprintf("/api/collections/%s", url.PathEscape(idOrName))
	var res Collection
	if err := s.Client.send(ctx, http.MethodPatch, path, col, &res); err != nil {
		return nil, fmt.Errorf("pocketbase: update collection: %w", err)
	}
	return &res, nil
}

// Delete deletes a collection.
func (s *CollectionService) Delete(ctx context.Context, idOrName string) error {
	path := fmt.Sprintf("/api/collections/%s", url.PathEscape(idOrName))
	if err := s.Client.send(ctx, http.MethodDelete, path, nil, nil); err != nil {
		return fmt.Errorf("pocketbase: delete collection: %w", err)
	}
	return nil
}

// Import creates or updates multiple collections at once.
func (s *CollectionService) Import(ctx context.Context, cols []*Collection, deleteMissing bool) ([]*Collection, error) {
	path := "/api/collections/import"
	q := url.Values{}
	if deleteMissing {
		q.Set("deleteMissing", "1")
	}
	if qs := q.Encode(); qs != "" {
		path += "?" + qs
	}
	var res []*Collection
	if err := s.Client.send(ctx, http.MethodPut, path, cols, &res); err != nil {
		return nil, fmt.Errorf("pocketbase: import collections: %w", err)
	}
	return res, nil
}
