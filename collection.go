package pocketbase

import (
	"context"
	"encoding/json"
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
	ID         string        `json:"id,omitempty"`
	Created    DateTime      `json:"created,omitempty"`
	Updated    DateTime      `json:"updated,omitempty"`
	Name       string        `json:"name"`
	Type       string        `json:"type"`
	System     bool          `json:"system"`
	Fields     []SchemaField `json:"fields"`
	ListRule   *string       `json:"listRule"`
	ViewRule   *string       `json:"viewRule"`
	CreateRule *string       `json:"createRule"`
	UpdateRule *string       `json:"updateRule"`
	DeleteRule *string       `json:"deleteRule"`
	Indexes    []string      `json:"indexes"`
	// Options holds flattened auth/view configuration, such as passwordAuth and viewQuery.
	Options map[string]any `json:"-"`
}

func (c *Collection) GetID() string             { return c.ID }
func (c *Collection) GetCollectionName() string { return c.Name }

func (c Collection) MarshalJSON() ([]byte, error) {
	type plain Collection
	return marshalCollectionOptions(plain(c), c.Options)
}

func (c *Collection) UnmarshalJSON(data []byte) error {
	type plain Collection
	var decoded plain
	options, err := unmarshalCollectionOptions(data, &decoded)
	if err != nil {
		return err
	}
	*c = Collection(decoded)
	c.Options = options
	return nil
}

// SchemaField defines a collection field.
type SchemaField struct {
	ID          string `json:"id,omitempty"`
	Name        string `json:"name"`
	Type        string `json:"type"`
	Required    bool   `json:"required"`
	Presentable bool   `json:"presentable"`
	System      bool   `json:"system"`
	Hidden      bool   `json:"hidden"`
	// Options holds flattened type-specific settings such as min, max and maxSelect.
	Options map[string]any `json:"-"`
}

func (f SchemaField) MarshalJSON() ([]byte, error) {
	type plain SchemaField
	return marshalCollectionOptions(plain(f), f.Options)
}

func (f *SchemaField) UnmarshalJSON(data []byte) error {
	type plain SchemaField
	var decoded plain
	options, err := unmarshalCollectionOptions(data, &decoded)
	if err != nil {
		return err
	}
	*f = SchemaField(decoded)
	f.Options = options
	return nil
}

// Typed properties take precedence over options, including nullable access rules.
func marshalCollectionOptions(value any, options map[string]any) ([]byte, error) {
	data, err := json.Marshal(value)
	if err != nil {
		return nil, err
	}
	var fields map[string]any
	if err := json.Unmarshal(data, &fields); err != nil {
		return nil, err
	}
	for key, value := range options {
		if _, exists := fields[key]; !exists {
			fields[key] = value
		}
	}
	return json.Marshal(fields)
}

func unmarshalCollectionOptions(data []byte, value any) (map[string]any, error) {
	if err := json.Unmarshal(data, value); err != nil {
		return nil, err
	}
	var options map[string]any
	if err := json.Unmarshal(data, &options); err != nil {
		return nil, err
	}
	known, err := json.Marshal(value)
	if err != nil {
		return nil, err
	}
	var fields map[string]any
	if err := json.Unmarshal(known, &fields); err != nil {
		return nil, err
	}
	for key := range fields {
		delete(options, key)
	}
	return options, nil
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
	applyListOptions(q, opts)
	if qs := q.Encode(); qs != "" {
		path += "?" + qs
	}
	var res CollectionListResult
	if err := s.Client.send(ctx, http.MethodGet, path, nil, &res); err != nil {
		return nil, err
	}
	return &res, nil
}

// GetOne retrieves a specific collection.
func (s *CollectionService) GetOne(ctx context.Context, idOrName string) (*Collection, error) {
	path := fmt.Sprintf("/api/collections/%s", url.PathEscape(idOrName))
	var col Collection
	if err := s.Client.send(ctx, http.MethodGet, path, nil, &col); err != nil {
		return nil, err
	}
	return &col, nil
}

// Create creates a new collection.
func (s *CollectionService) Create(ctx context.Context, col *Collection) (*Collection, error) {
	path := "/api/collections"
	var res Collection
	if err := s.Client.send(ctx, http.MethodPost, path, col, &res); err != nil {
		return nil, err
	}
	return &res, nil
}

// Update modifies an existing collection.
func (s *CollectionService) Update(ctx context.Context, idOrName string, col *Collection) (*Collection, error) {
	path := fmt.Sprintf("/api/collections/%s", url.PathEscape(idOrName))
	var res Collection
	if err := s.Client.send(ctx, http.MethodPatch, path, col, &res); err != nil {
		return nil, err
	}
	return &res, nil
}

// Delete deletes a collection.
func (s *CollectionService) Delete(ctx context.Context, idOrName string) error {
	path := fmt.Sprintf("/api/collections/%s", url.PathEscape(idOrName))
	if err := s.Client.send(ctx, http.MethodDelete, path, nil, nil); err != nil {
		return err
	}
	return nil
}

// Import creates or updates multiple collections at once. Success returns nil, nil
// because PocketBase responds with 204 No Content.
func (s *CollectionService) Import(ctx context.Context, cols []*Collection, deleteMissing bool) ([]*Collection, error) {
	path := "/api/collections/import"
	body := map[string]any{"collections": cols, "deleteMissing": deleteMissing}
	if err := s.Client.send(ctx, http.MethodPut, path, body, nil); err != nil {
		return nil, err
	}
	return nil, nil
}
