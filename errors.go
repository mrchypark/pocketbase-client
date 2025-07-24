// filename: errors.go
package pocketbase

import (
	"errors"
	"fmt"
)

// APIError represents a structured error returned from the PocketBase API.
type APIError struct {
	Code    int            `json:"code"`
	Message string         `json:"message"`
	Data    map[string]any `json:"data"`
}

func (e *APIError) Error() string {
	return fmt.Sprintf("pocketbase: API error (code=%d): %s", e.Code, e.Message)
}

// ClientError is a custom error type that wraps both a general client-side error
// and the original detailed error from the API.
type ClientError struct {
	// BaseErr is the general, client-side error (e.g., ErrNotFound).
	BaseErr error
	// OriginalErr is the original, detailed error from the API.
	OriginalErr *APIError
}

func (e *ClientError) Error() string {
	return fmt.Sprintf("%s: %s", e.BaseErr, e.OriginalErr)
}

// Unwrap allows errors.As to find the underlying *APIError.
func (e *ClientError) Unwrap() error {
	return e.OriginalErr
}

// Is allows errors.Is to check against the base error type.
func (e *ClientError) Is(target error) bool {
	return errors.Is(e.BaseErr, target)
}

// --- Predefined Client Errors ---
// (This part remains the same)
var (
	ErrBadRequest            = errors.New("the request is invalid")
	ErrForbidden             = errors.New("you are not allowed to perform this request")
	ErrNotFound              = errors.New("the requested resource is not found")
	ErrUnauthorized          = errors.New("the request requires a valid auth token")
	ErrInvalidCredentials    = errors.New("invalid email/username or password")
	ErrUserNotFound          = errors.New("auth record not found")
	ErrUserNotVerified       = errors.New("the auth record is not verified")
	ErrTokenInvalidOrExpired = errors.New("the provided token is invalid or has expired")
	ErrRecordInvalidData     = errors.New("failed to load the submitted data")
	ErrFileTooLarge          = errors.New("the uploaded file is too large")
	ErrInvalidFileType       = errors.New("the uploaded file type is not allowed")
	ErrUnknown               = errors.New("an unexpected server error occurred")
)
