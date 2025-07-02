package pocketbase

import (
	"context"
	"fmt"
	"net/http"
	"net/url"
)

// AdminServiceAPI defines the API operations for admin accounts.
type AdminServiceAPI interface {
	GetList(ctx context.Context, opts *ListOptions) (*ListResult, error)
	GetOne(ctx context.Context, adminID string) (*Admin, error)
	Create(ctx context.Context, body interface{}) (*Admin, error)
	Update(ctx context.Context, adminID string, body interface{}) (*Admin, error)
	Delete(ctx context.Context, adminID string) error
}

// AdminService provides API for managing admin accounts.
type AdminService struct {
	Client *Client
}

var _ AdminServiceAPI = (*AdminService)(nil)

// GetList retrieves a list of administrators.
// ListOptions can be used to specify page numbers, etc.
func (s *AdminService) GetList(ctx context.Context, opts *ListOptions) (*ListResult, error) {
	path := "/api/admins"
	q := url.Values{}
	applyListOptions(q, opts)
	if qs := q.Encode(); qs != "" {
		path += "?" + qs
	}
	var res ListResult
	if err := s.Client.send(ctx, http.MethodGet, path, nil, &res); err != nil {
		return nil, fmt.Errorf("pocketbase: fetch admins list: %w", err)
	}
	return &res, nil
}

// GetOne retrieves a single administrator.
func (s *AdminService) GetOne(ctx context.Context, adminID string) (*Admin, error) {
	path := fmt.Sprintf("/api/admins/%s", url.PathEscape(adminID))
	var adm Admin
	if err := s.Client.send(ctx, http.MethodGet, path, nil, &adm); err != nil {
		return nil, fmt.Errorf("pocketbase: fetch admin: %w", err)
	}
	return &adm, nil
}

// Create creates a new administrator.
func (s *AdminService) Create(ctx context.Context, body interface{}) (*Admin, error) {
	path := "/api/admins"
	var adm Admin
	if err := s.Client.send(ctx, http.MethodPost, path, body, &adm); err != nil {
		return nil, fmt.Errorf("pocketbase: create admin: %w", err)
	}
	return &adm, nil
}

// Update modifies administrator information.
func (s *AdminService) Update(ctx context.Context, adminID string, body interface{}) (*Admin, error) {
	path := fmt.Sprintf("/api/admins/%s", url.PathEscape(adminID))
	var adm Admin
	if err := s.Client.send(ctx, http.MethodPatch, path, body, &adm); err != nil {
		return nil, fmt.Errorf("pocketbase: update admin: %w", err)
	}
	return &adm, nil
}

// Delete deletes an administrator.
func (s *AdminService) Delete(ctx context.Context, adminID string) error {
	path := fmt.Sprintf("/api/admins/%s", url.PathEscape(adminID))
	if err := s.Client.send(ctx, http.MethodDelete, path, nil, nil); err != nil {
		return fmt.Errorf("pocketbase: delete admin: %w", err)
	}
	return nil
}
