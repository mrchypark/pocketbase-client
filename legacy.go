package pocketbase

import (
	"context"
	"fmt"
	"net/http"
	"net/url"
)

// LegacyServiceAPI defines the legacy API operations.
type LegacyServiceAPI interface {
	AdminAuthRefresh(ctx context.Context) (*AuthResponse, error)
	RecordAuthRefresh(ctx context.Context, collection string) (*AuthResponse, error)
	RequestEmailChange(ctx context.Context, collection, newEmail string) error
	ConfirmEmailChange(ctx context.Context, collection, token, password string) error
	ListExternalAuths(ctx context.Context, collection, recordID string) ([]map[string]any, error)
	UnlinkExternalAuth(ctx context.Context, collection, recordID, provider string) error
	AuthWithOAuth2(ctx context.Context, collection string, req *OAuth2Request) (*AuthResponse, error)
}

// LegacyService provides access to PocketBase legacy API endpoints.
type LegacyService struct {
	Client *Client
}

var _ LegacyServiceAPI = (*LegacyService)(nil)

// AdminAuthRefresh refreshes the admin auth token.
func (s *LegacyService) AdminAuthRefresh(ctx context.Context) (*AuthResponse, error) {
	path := "/api/admins/auth-refresh"
	var res AuthResponse
	if err := s.Client.send(ctx, http.MethodPost, path, nil, &res); err != nil {
		return nil, fmt.Errorf("pocketbase: refresh admin auth: %w", err)
	}
	s.Client.AuthStore.Set(res.Token, res.Admin)
	return &res, nil
}

// AuthenticateAsAdmin authenticates as an admin and stores the authentication information.
// This method is for PocketBase v0.22 and older.
func (s *LegacyService) AuthenticateAsAdmin(ctx context.Context, identity, password string) (*AuthResponse, error) {
	s.Client.ClearAuthStore()

	reqBody := map[string]string{
		"identity": identity,
		"password": password,
	}

	var authResponse AuthResponse
	path := "/api/admins/auth-with-password"
	if err := s.Client.send(ctx, http.MethodPost, path, reqBody, &authResponse); err != nil {
		return nil, err
	}

	s.Client.AuthStore.Set(authResponse.Token, authResponse.Admin)

	return &authResponse, nil
}

// RecordAuthRefresh refreshes the auth token for the specified collection.
func (s *LegacyService) RecordAuthRefresh(ctx context.Context, collection string) (*AuthResponse, error) {
	path := fmt.Sprintf("/api/collections/%s/auth-refresh", url.PathEscape(collection))
	var res AuthResponse
	if err := s.Client.send(ctx, http.MethodPost, path, nil, &res); err != nil {
		return nil, fmt.Errorf("pocketbase: refresh record auth: %w", err)
	}
	s.Client.AuthStore.Set(res.Token, res.Record)
	return &res, nil
}

// RequestEmailChange sends an email change request for the logged in record.
func (s *LegacyService) RequestEmailChange(ctx context.Context, collection, newEmail string) error {
	path := fmt.Sprintf("/api/collections/%s/request-email-change", url.PathEscape(collection))
	body := map[string]string{"newEmail": newEmail}
	if err := s.Client.send(ctx, http.MethodPost, path, body, nil); err != nil {
		return fmt.Errorf("pocketbase: request email change: %w", err)
	}
	return nil
}

// ConfirmEmailChange confirms an email change using the provided token.
func (s *LegacyService) ConfirmEmailChange(ctx context.Context, collection, token, password string) error {
	path := fmt.Sprintf("/api/collections/%s/confirm-email-change", url.PathEscape(collection))
	body := map[string]string{"token": token, "password": password}
	if err := s.Client.send(ctx, http.MethodPost, path, body, nil); err != nil {
		return fmt.Errorf("pocketbase: confirm email change: %w", err)
	}
	return nil
}

// ListExternalAuths lists the external auth providers linked to the record.
func (s *LegacyService) ListExternalAuths(ctx context.Context, collection, recordID string) ([]map[string]any, error) {
	path := fmt.Sprintf("/api/collections/%s/records/%s/external-auths", url.PathEscape(collection), url.PathEscape(recordID))
	var res []map[string]any
	if err := s.Client.send(ctx, http.MethodGet, path, nil, &res); err != nil {
		return nil, fmt.Errorf("pocketbase: list external auths: %w", err)
	}
	return res, nil
}

// UnlinkExternalAuth unlinks the specified external auth provider from the record.
func (s *LegacyService) UnlinkExternalAuth(ctx context.Context, collection, recordID, provider string) error {
	path := fmt.Sprintf("/api/collections/%s/records/%s/external-auths/%s", url.PathEscape(collection), url.PathEscape(recordID), url.PathEscape(provider))
	if err := s.Client.send(ctx, http.MethodDelete, path, nil, nil); err != nil {
		return fmt.Errorf("pocketbase: unlink external auth: %w", err)
	}
	return nil
}

// OAuth2Request represents parameters for OAuth2 authentication.
type OAuth2Request struct {
	Provider     string         `json:"provider"`
	Code         string         `json:"code"`
	CodeVerifier string         `json:"codeVerifier,omitempty"`
	RedirectURL  string         `json:"redirectUrl"`
	CreateData   map[string]any `json:"createData,omitempty"`
}

// AuthWithOAuth2 authenticates a record using OAuth2 provider.
func (s *LegacyService) AuthWithOAuth2(ctx context.Context, collection string, req *OAuth2Request) (*AuthResponse, error) {
	path := fmt.Sprintf("/api/collections/%s/auth-with-oauth2", url.PathEscape(collection))
	var res AuthResponse
	if err := s.Client.send(ctx, http.MethodPost, path, req, &res); err != nil {
		return nil, fmt.Errorf("pocketbase: auth with oauth2: %w", err)
	}
	if res.Admin != nil {
		s.Client.AuthStore.Set(res.Token, res.Admin)
	} else {
		s.Client.AuthStore.Set(res.Token, res.Record)
	}
	return &res, nil
}
