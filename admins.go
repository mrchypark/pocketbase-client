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
func (s *AdminService) GetList(ctx context.Context, opts *ListOptions) (*ListResult, error) {
	path := buildPathWithQuery("/api/admins", buildQueryString(opts))
	var res ListResult
	if err := s.Client.send(ctx, http.MethodGet, path, nil, &res); err != nil {
		return nil, wrapError("fetch", "admins list", err)
	}
	return &res, nil
}

// GetOne retrieves a single administrator.
func (s *AdminService) GetOne(ctx context.Context, adminID string) (*Admin, error) {
	path := fmt.Sprintf("/api/admins/%s", url.PathEscape(adminID))
	var adm Admin
	if err := s.Client.send(ctx, http.MethodGet, path, nil, &adm); err != nil {
		return nil, wrapError("fetch", "admin", err)
	}
	return &adm, nil
}

// Create creates a new administrator.
func (s *AdminService) Create(ctx context.Context, body interface{}) (*Admin, error) {
	var adm Admin
	if err := s.Client.send(ctx, http.MethodPost, "/api/admins", body, &adm); err != nil {
		return nil, wrapError("create", "admin", err)
	}
	return &adm, nil
}

// Update modifies administrator information.
func (s *AdminService) Update(ctx context.Context, adminID string, body interface{}) (*Admin, error) {
	path := fmt.Sprintf("/api/admins/%s", url.PathEscape(adminID))
	var adm Admin
	if err := s.Client.send(ctx, http.MethodPatch, path, body, &adm); err != nil {
		return nil, wrapError("update", "admin", err)
	}
	return &adm, nil
}

// Delete deletes an administrator.
func (s *AdminService) Delete(ctx context.Context, adminID string) error {
	path := fmt.Sprintf("/api/admins/%s", url.PathEscape(adminID))
	if err := s.Client.send(ctx, http.MethodDelete, path, nil, nil); err != nil {
		return wrapError("delete", "admin", err)
	}
	return nil
}
