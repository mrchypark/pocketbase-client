package pocketbase

import (
	"context"
	"fmt"
	"net/http"
	"net/url"
)

// UserServiceAPI defines the API operations related to regular user accounts.
// UserServiceAPI defines the API operations related to regular user accounts.
type UserServiceAPI interface {
	RequestPasswordReset(ctx context.Context, collection, email string) error
	ConfirmPasswordReset(ctx context.Context, collection, token, newPassword, newPasswordConfirm string) error
	RequestVerification(ctx context.Context, collection, email string) error
	ConfirmVerification(ctx context.Context, collection, token string) error
	GetOAuth2Providers(ctx context.Context, collection string) (map[string]interface{}, error)
	AuthWithOAuth2(ctx context.Context, collection, provider, code, verifier, redirect string, createData map[string]any) (*AuthResponse, error)
	AuthRefresh(ctx context.Context, collection string) (*AuthResponse, error)
	RequestOTP(ctx context.Context, collection, email string) (map[string]string, error)
	AuthWithOTP(ctx context.Context, collection, otpID, password string) (*AuthResponse, error)
	RequestEmailChange(ctx context.Context, collection, newEmail string) error
	ConfirmEmailChange(ctx context.Context, collection, token, password string) error
	Impersonate(ctx context.Context, collection, id string, duration int) (*Client, error)
}

// UserService provides API related to regular user accounts.
type UserService struct {
	Client *Client
}

var _ UserServiceAPI = (*UserService)(nil)

// RequestPasswordReset sends a password reset email.
func (s *UserService) RequestPasswordReset(ctx context.Context, collection, email string) error {
	path := fmt.Sprintf("/api/collections/%s/request-password-reset", url.PathEscape(collection))
	body := map[string]string{"email": email}
	if err := s.Client.send(ctx, http.MethodPost, path, body, nil); err != nil {
		return fmt.Errorf("pocketbase: request password reset: %w", err)
	}
	return nil
}

// ConfirmPasswordReset completes the password reset process.
func (s *UserService) ConfirmPasswordReset(ctx context.Context, collection, token, newPassword, newPasswordConfirm string) error {
	path := fmt.Sprintf("/api/collections/%s/confirm-password-reset", url.PathEscape(collection))
	body := map[string]string{"token": token, "password": newPassword, "passwordConfirm": newPasswordConfirm}
	if err := s.Client.send(ctx, http.MethodPost, path, body, nil); err != nil {
		return fmt.Errorf("pocketbase: confirm password reset: %w", err)
	}
	return nil
}

// RequestVerification sends an email verification email.
func (s *UserService) RequestVerification(ctx context.Context, collection, email string) error {
	path := fmt.Sprintf("/api/collections/%s/request-verification", url.PathEscape(collection))
	body := map[string]string{"email": email}
	if err := s.Client.send(ctx, http.MethodPost, path, body, nil); err != nil {
		return fmt.Errorf("pocketbase: request verification: %w", err)
	}
	return nil
}

// ConfirmVerification confirms an email verification token.
func (s *UserService) ConfirmVerification(ctx context.Context, collection, token string) error {
	path := fmt.Sprintf("/api/collections/%s/confirm-verification", url.PathEscape(collection))
	body := map[string]string{"token": token}
	if err := s.Client.send(ctx, http.MethodPost, path, body, nil); err != nil {
		return fmt.Errorf("pocketbase: confirm verification: %w", err)
	}
	return nil
}

// GetOAuth2Providers retrieves OAuth2 provider information.
func (s *UserService) GetOAuth2Providers(ctx context.Context, collection string) (map[string]interface{}, error) {
	path := fmt.Sprintf("/api/collections/%s/auth-methods", url.PathEscape(collection))
	var result map[string]interface{}
	err := s.Client.send(ctx, http.MethodGet, path, nil, &result)
	if err != nil {
		return nil, fmt.Errorf("pocketbase: get oauth2 providers: %w", err)
	}
	return result, nil
}

// AuthWithOAuth2 authenticates with an OAuth2 code.
func (s *UserService) AuthWithOAuth2(ctx context.Context, collection, provider, code, verifier, redirect string, createData map[string]any) (*AuthResponse, error) {
	path := fmt.Sprintf("/api/collections/%s/auth-with-oauth2", url.PathEscape(collection))
	body := map[string]any{
		"provider":     provider,
		"code":         code,
		"codeVerifier": verifier,
		"redirectUrl":  redirect,
	}
	if createData != nil {
		body["createData"] = createData
	}
	var res AuthResponse
	if err := s.Client.send(ctx, http.MethodPost, path, body, &res); err != nil {
		return nil, fmt.Errorf("pocketbase: auth with oauth2: %w", err)
	}

	return &res, nil
}

// AuthRefresh refreshes the stored token.
func (s *UserService) AuthRefresh(ctx context.Context, collection string) (*AuthResponse, error) {
	path := fmt.Sprintf("/api/collections/%s/auth-refresh", url.PathEscape(collection))
	var res AuthResponse
	if err := s.Client.send(ctx, http.MethodPost, path, nil, &res); err != nil {
		return nil, fmt.Errorf("pocketbase: auth refresh: %w", err)
	}

	return &res, nil
}

// RequestOTP requests a one-time password.
func (s *UserService) RequestOTP(ctx context.Context, collection, email string) (map[string]string, error) {
	path := fmt.Sprintf("/api/collections/%s/request-otp", url.PathEscape(collection))
	body := map[string]string{"email": email}
	var res map[string]string
	if err := s.Client.send(ctx, http.MethodPost, path, body, &res); err != nil {
		return nil, fmt.Errorf("pocketbase: request otp: %w", err)
	}
	return res, nil
}

// AuthWithOTP authenticates with an OTP ID and password.
func (s *UserService) AuthWithOTP(ctx context.Context, collection, otpID, password string) (*AuthResponse, error) {
	path := fmt.Sprintf("/api/collections/%s/auth-with-otp", url.PathEscape(collection))
	body := map[string]string{"otpId": otpID, "password": password}
	var res AuthResponse
	if err := s.Client.send(ctx, http.MethodPost, path, body, &res); err != nil {
		return nil, fmt.Errorf("pocketbase: auth with otp: %w", err)
	}

	return &res, nil
}

// RequestEmailChange requests an email change.
func (s *UserService) RequestEmailChange(ctx context.Context, collection, newEmail string) error {
	path := fmt.Sprintf("/api/collections/%s/request-email-change", url.PathEscape(collection))
	body := map[string]string{"newEmail": newEmail}
	if err := s.Client.send(ctx, http.MethodPost, path, body, nil); err != nil {
		return fmt.Errorf("pocketbase: request email change: %w", err)
	}
	return nil
}

// ConfirmEmailChange confirms an email change.
func (s *UserService) ConfirmEmailChange(ctx context.Context, collection, token, password string) error {
	path := fmt.Sprintf("/api/collections/%s/confirm-email-change", url.PathEscape(collection))
	body := map[string]string{"token": token, "password": password}
	if err := s.Client.send(ctx, http.MethodPost, path, body, nil); err != nil {
		return fmt.Errorf("pocketbase: confirm email change: %w", err)
	}
	return nil
}

// Impersonate returns a new client that can impersonate another user.
func (s *UserService) Impersonate(ctx context.Context, collection, id string, duration int) (*Client, error) {
	path := fmt.Sprintf("/api/collections/%s/impersonate/%s", url.PathEscape(collection), url.PathEscape(id))
	body := map[string]int{"duration": duration}
	var res AuthResponse
	if err := s.Client.send(ctx, http.MethodPost, path, body, &res); err != nil {
		return nil, fmt.Errorf("pocketbase: impersonate user: %w", err)
	}
	client := NewClient(s.Client.BaseURL)
	client.UseAuthResponse(&res)
	return client, nil
}
