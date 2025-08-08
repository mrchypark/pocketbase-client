// Package pocketbase provides extended error handling utilities for PocketBase API interactions.
//
// This file defines a richer error type (`Error`) than the basic `APIError`.
// It classifies errors based on HTTP status codes, parses structured JSON error bodies,
// and normalizes server error messages to stable alias codes. The alias set is derived
// from the PocketBase documentation for main and v0.22.x releases.
//
// Consumers of this client can use the `ParseAPIError` helper to convert an HTTP response
// into an `*Error`, then use classification helpers such as `IsValidation()` or
// `IsAuth()` to branch on the category of failure.
package pocketbase

import (
	"errors"
	"fmt"
	"maps"
	"net/http"
	"strings"
	"sync"

	"github.com/goccy/go-json"
)

// =============================================================================
// Core Types
// =============================================================================

// HTTPStatus wraps HTTP status codes to make them compatible with errors.Is.
// This allows direct comparison with HTTP status codes using the standard
// errors.Is pattern.
//
// Example:
//
//	if errors.Is(err, HTTPStatus(http.StatusNotFound)) {
//	    // Handle 404 error
//	}
type HTTPStatus int

// Error implements the error interface for HTTPStatus.
func (s HTTPStatus) Error() string {
	if text := http.StatusText(int(s)); text != "" {
		return fmt.Sprintf("HTTP %d %s", int(s), text)
	}
	return fmt.Sprintf("HTTP %d", int(s))
}

// Common HTTP status constants for convenience.
const (
	StatusBadRequest            = HTTPStatus(http.StatusBadRequest)
	StatusUnauthorized          = HTTPStatus(http.StatusUnauthorized)
	StatusForbidden             = HTTPStatus(http.StatusForbidden)
	StatusNotFound              = HTTPStatus(http.StatusNotFound)
	StatusRequestEntityTooLarge = HTTPStatus(http.StatusRequestEntityTooLarge)
	StatusTooManyRequests       = HTTPStatus(http.StatusTooManyRequests)
	StatusInternalServerError   = HTTPStatus(http.StatusInternalServerError)
)

// FieldError captures validation failures for a specific field in the PocketBase API.
// The Code is a machine-readable identifier and Message is a human-readable explanation.
type FieldError struct {
	Code    string `json:"code"`
	Message string `json:"message"`
}

// DebugInfo contains debugging information for PocketBase API errors.
// This information is only populated when debug mode is enabled.
type DebugInfo struct {
	// Endpoint records the API endpoint used, for debugging.
	Endpoint string
	// RawHeaders captures the original response headers.
	RawHeaders http.Header
	// RawBody captures the original response body bytes.
	RawBody []byte
}

// Error represents a normalized PocketBase API error. It contains the HTTP status,
// a generic message, optional per-field errors, and derived metadata such as the
// error alias code and optional debugging information.
type Error struct {
	// Status is the HTTP status code.
	Status int `json:"status"`
	// Code is an optional alias derived from the server message (see messageAliases).
	Code string `json:"code"`
	// Message is the top-level error message returned by the server.
	Message string `json:"message"`
	// Data contains per-field validation errors, keyed by field name.
	Data map[string]FieldError `json:"data"`

	// Debug contains optional debugging information (not serialized)
	Debug *DebugInfo `json:"-"`
}

// =============================================================================
// Error Interface Implementation
// =============================================================================

// Error implements the error interface for *Error.
func (e *Error) Error() string {
	if e == nil {
		return "<nil>"
	}

	var parts []string

	// Base HTTP status
	if statusText := http.StatusText(e.Status); statusText != "" {
		parts = append(parts, fmt.Sprintf("pocketbase: %d %s", e.Status, statusText))
	} else {
		parts = append(parts, fmt.Sprintf("pocketbase: %d", e.Status))
	}

	// Add code if present
	if e.Code != "" {
		parts = append(parts, "code="+e.Code)
	}

	// Add message if present
	if e.Message != "" {
		parts = append(parts, "msg="+e.Message)
	}

	// Add field error count if present
	if len(e.Data) > 0 {
		parts = append(parts, fmt.Sprintf("data=%d field error(s)", len(e.Data)))
	}

	// Add endpoint if present (debugging)
	if e.Debug != nil && e.Debug.Endpoint != "" {
		parts = append(parts, "at="+e.Debug.Endpoint)
	}

	return strings.Join(parts, " ")
}

// Unwrap satisfies errors.Unwrap but returns nil because Error is the root error type.
func (e *Error) Unwrap() error { return nil }

// Is implements the errors.Is semantics for *Error.
//
// It allows callers to use standard library errors.Is to compare PocketBase
// errors by their HTTP status and Code alias. Supported target types:
//   - *Error: matches by Status (if non-zero) and Code (if non-empty) - all specified fields must match
//   - HTTPStatus: matches by HTTP status code
//
// Other fields such as Message, Data, Endpoint, RawHeaders and RawBody are
// ignored during comparisons.
func (e *Error) Is(target error) bool {
	if e == nil {
		return target == nil
	}

	switch t := target.(type) {
	case *Error:
		// An empty target error shouldn't match anything.
		if t.Status == 0 && t.Code == "" {
			return false
		}
		// If target specifies a status, it must match.
		if t.Status != 0 && e.Status != t.Status {
			return false
		}
		// If target specifies a code, it must match.
		if t.Code != "" && e.Code != t.Code {
			return false
		}
		// All specified fields in the target matched.
		return true
	case HTTPStatus:
		// Match by HTTP status code
		return e.Status == int(t)
	}
	return false
}

// =============================================================================
// Error Classification Methods
// =============================================================================

// IsValidation reports whether the error represents a validation failure (HTTP 400 with field errors).
func (e *Error) IsValidation() bool {
	return e != nil && e.Status == http.StatusBadRequest && len(e.Data) > 0
}

// IsAuth reports whether the error indicates authentication failure (401).
func (e *Error) IsAuth() bool {
	return e != nil && e.Status == http.StatusUnauthorized
}

// IsForbidden reports whether the error indicates an authorization failure (403).
func (e *Error) IsForbidden() bool {
	return e != nil && e.Status == http.StatusForbidden
}

// IsNotFound reports whether the error corresponds to a missing resource (404).
func (e *Error) IsNotFound() bool {
	return e != nil && e.Status == http.StatusNotFound
}

// IsRateLimited reports whether the error corresponds to a rate limit violation (429).
func (e *Error) IsRateLimited() bool {
	return e != nil && e.Status == http.StatusTooManyRequests
}

// IsInternal reports whether the error corresponds to an internal server error (500).
func (e *Error) IsInternal() bool {
	return e != nil && e.Status == http.StatusInternalServerError
}

// IsBadRequest reports whether the error corresponds to a bad request (400).
func (e *Error) IsBadRequest() bool {
	return e != nil && e.Status == http.StatusBadRequest
}

// =============================================================================
// Error Parsing
// =============================================================================

// rawPocketBaseError matches the JSON structure returned by the PocketBase API.
// It is used internally to decode error bodies.
type rawPocketBaseError struct {
	Code    int                   `json:"code"`
	Message string                `json:"message"`
	Data    map[string]FieldError `json:"data"`
}

// ParseAPIError converts an HTTP response and body into an *Error. If the response
// status code is < 400, nil is returned.
func ParseAPIError(statusCode int, headers http.Header, body []byte, endpoint string) error {
	if statusCode < 400 {
		return nil
	}

	var wire rawPocketBaseError
	_ = json.Unmarshal(body, &wire) // tolerate invalid JSON

	e := &Error{
		Status:  statusCode,
		Message: strings.TrimSpace(wire.Message),
		Data:    wire.Data,
		Debug: &DebugInfo{
			Endpoint:   endpoint,
			RawHeaders: headers.Clone(),
			RawBody:    append([]byte(nil), body...), // defensive copy
		},
	}

	if e.Data == nil {
		e.Data = make(map[string]FieldError)
	}

	// Apply message alias mapping
	applyKnownAliases(e)
	return e
}

// ParseAPIErrorFromResponse is a convenience wrapper around ParseAPIError that accepts
// an *http.Response. This maintains backward compatibility with existing code.
func ParseAPIErrorFromResponse(resp *http.Response, body []byte, endpoint string) error {
	if resp == nil {
		return errors.New("pocketbase: nil http.Response")
	}
	return ParseAPIError(resp.StatusCode, resp.Header, body, endpoint)
}

// =============================================================================
// Message Alias System
// =============================================================================

// messageAliases maps server-provided error messages to stable alias codes.
// This map is populated at package initialization and can be extended at runtime.
var messageAliases = make(map[string]string)

// aliasCache provides thread-safe caching for alias lookups to improve performance.
var aliasCache = struct {
	mu    sync.RWMutex
	cache map[string]string
}{
	cache: make(map[string]string),
}

// RegisterMessageAlias adds or overrides a server message → alias mapping.
// Applications may call this during initialization to customize alias names.
// This function is thread-safe.
func RegisterMessageAlias(serverMessage, alias string) {
	aliasCache.mu.Lock()
	defer aliasCache.mu.Unlock()

	messageAliases[serverMessage] = alias
	// Invalidate cache entry if it exists
	delete(aliasCache.cache, serverMessage)
}

// getMessageAlias retrieves the alias for a message with caching for performance.
func getMessageAlias(message string) (string, bool) {
	// Fast path: check cache first
	aliasCache.mu.RLock()
	if alias, ok := aliasCache.cache[message]; ok {
		aliasCache.mu.RUnlock()
		return alias, true
	}
	aliasCache.mu.RUnlock()

	// Slow path: check main map and update cache
	aliasCache.mu.Lock()
	defer aliasCache.mu.Unlock()

	// Double-check in case another goroutine updated it
	if alias, ok := aliasCache.cache[message]; ok {
		return alias, true
	}

	if alias, ok := messageAliases[message]; ok {
		aliasCache.cache[message] = alias
		return alias, true
	}

	// Cache negative result to avoid repeated lookups
	aliasCache.cache[message] = ""
	return "", false
}

// applyKnownAliases stores the alias derived from messageAliases back into e.Code.
func applyKnownAliases(e *Error) {
	if e == nil || e.Message == "" {
		return
	}
	if alias, ok := getMessageAlias(e.Message); ok && alias != "" {
		e.Code = alias
	}
}

// initializeMessageAliases populates the messageAliases map with known mappings.
func initializeMessageAliases() {
	aliases := map[string]string{
		// Authentication errors
		"Missing or invalid authentication.":             "invalid_authentication",
		"Missing or invalid authentication token.":       "invalid_auth_token",
		"Missing or invalid admin authorization token.":  "invalid_admin_auth_token",
		"Missing or invalid record authorization token.": "invalid_record_auth_token",

		// Authorization errors
		"You are not allowed to perform this request.":                 "forbidden_generic",
		"The authorized record is not allowed to perform this action.": "record_forbidden",
		"The authorized admin is not allowed to perform this action.":  "admin_forbidden",

		// Access control errors
		"The request can be accessed only by guests.":                "only_guests",
		"The request can be accessed only by authenticated admins.":  "only_admins",
		"The request can be accessed only by authenticated records.": "only_records",
		"The request requires valid admin authorization token.":      "require_admin_token",
		"The request requires valid record authorization token.":     "require_record_token",

		// Parameter validation errors
		"Invalid \"sort\" parameter format.":      "invalid_sort_param",
		"Invalid \"expand\" parameter format.":    "invalid_expand_param",
		"Invalid \"filter\" parameter format.":    "invalid_filter_param",
		"Invalid \"fields\" parameter format.":    "invalid_fields_param",
		"Invalid \"page\" parameter format.":      "invalid_page_param",
		"Invalid \"perPage\" parameter format.":   "invalid_perpage_param",
		"Invalid \"skipTotal\" parameter format.": "invalid_skiptotal_param",

		// Missing parameter errors
		"Missing required \"expand\" parameter.": "missing_expand_param",
		"Missing required \"filter\" parameter.": "missing_filter_param",
		"Missing required \"fields\" parameter.": "missing_fields_param",
		"Missing required \"sort\" parameter.":   "missing_sort_param",

		// Request format errors
		"Unsupported Content-Type": "unsupported_content_type",
		"Invalid request payload.": "invalid_request_payload",
		"Invalid request body.":    "invalid_request_body",
		"Invalid or missing file.": "invalid_or_missing_file",

		// Field validation errors
		"Invalid or missing field":                   "invalid_or_missing_field",
		"Invalid or missing record id.":              "invalid_record_id",
		"Invalid or missing password reset token.":   "invalid_password_reset_token",
		"Invalid or missing verification token.":     "invalid_verification_token",
		"Invalid or missing file token.":             "invalid_file_token",
		"Invalid or missing OAuth2 state parameter.": "invalid_oauth2_state",
		"Invalid or missing redirect URL.":           "invalid_redirect_url",
		"Invalid redirect status code.":              "invalid_redirect_status_code",

		// Resource not found errors
		"The requested resource wasn't found.": "resource_not_found",
		"File not found.":                      "file_not_found",
		"Collection not found.":                "collection_not_found",
		"Record not found.":                    "record_not_found",

		// Context errors
		"Missing or invalid collection context.": "collection_context_invalid",

		// Service errors
		"Failed to fetch admins info.":       "failed_fetch_admins",
		"Failed to fetch records info.":      "failed_fetch_records",
		"Failed to fetch collection schema.": "failed_fetch_schema",

		// Rate limiting and size errors
		"Request entity too large": "entity_too_large",
		"Too Many Requests.":       "too_many_requests",

		// Internal server errors
		"Something went wrong while processing your request.": "internal_generic",
	}

	// Copy all aliases to messageAliases map using maps.Copy (Go 1.21+)
	maps.Copy(messageAliases, aliases)
}

// Initialize message aliases at package load time
func init() {
	initializeMessageAliases()
}

// =============================================================================
// Convenience Functions
// =============================================================================

// checkErrorType is a generic helper for error type checking.
// It extracts a *Error from any error and applies the checker function.
func checkErrorType(err error, checker func(*Error) bool) bool {
	var pbErr *Error
	return errors.As(err, &pbErr) && checker(pbErr)
}

// IsValidationError checks if the error is a validation error (400 with field errors).
// This is equivalent to checking if the error is a bad request with field data.
func IsValidationError(err error) bool {
	return checkErrorType(err, (*Error).IsValidation)
}

// IsAuthError checks if the error is an authentication error (401).
// This is equivalent to: errors.Is(err, StatusUnauthorized)
func IsAuthError(err error) bool {
	return errors.Is(err, StatusUnauthorized)
}

// IsForbiddenError checks if the error is a forbidden error (403).
// This is equivalent to: errors.Is(err, StatusForbidden)
func IsForbiddenError(err error) bool {
	return errors.Is(err, StatusForbidden)
}

// IsNotFoundError checks if the error is a not found error (404).
// This is equivalent to: errors.Is(err, StatusNotFound)
func IsNotFoundError(err error) bool {
	return errors.Is(err, StatusNotFound)
}

// IsRateLimitedError checks if the error is a rate limited error (429).
// This is equivalent to: errors.Is(err, StatusTooManyRequests)
func IsRateLimitedError(err error) bool {
	return errors.Is(err, StatusTooManyRequests)
}

// IsInternalError checks if the error is an internal server error (500).
// This is equivalent to: errors.Is(err, StatusInternalServerError)
func IsInternalError(err error) bool {
	return errors.Is(err, StatusInternalServerError)
}

// IsBadRequestError checks if the error is a bad request error (400).
// This is equivalent to: errors.Is(err, StatusBadRequest)
func IsBadRequestError(err error) bool {
	return errors.Is(err, StatusBadRequest)
}

// HasHTTPStatus checks if the error has a specific HTTP status code.
// This is a generic function that works with any HTTP status code.
//
// Example:
//
//	if HasHTTPStatus(err, http.StatusNotFound) { ... }
func HasHTTPStatus(err error, status int) bool {
	return errors.Is(err, HTTPStatus(status))
}

// HasErrorCode checks if the error has a specific alias code.
func HasErrorCode(err error, code string) bool {
	var pbErr *Error
	return errors.As(err, &pbErr) && pbErr.Code == code
}

// GetErrorCode returns the alias code of the error, or empty string if not a PocketBase error.
func GetErrorCode(err error) string {
	var pbErr *Error
	if errors.As(err, &pbErr) {
		return pbErr.Code
	}
	return ""
}

// GetHTTPStatus returns the HTTP status code of the error, or 0 if not a PocketBase error.
func GetHTTPStatus(err error) int {
	var pbErr *Error
	if errors.As(err, &pbErr) {
		return pbErr.Status
	}
	return 0
}

// GetFieldErrors returns the field validation errors, or nil if not a validation error.
func GetFieldErrors(err error) map[string]FieldError {
	var pbErr *Error
	if errors.As(err, &pbErr) && pbErr.IsValidation() {
		return pbErr.Data
	}
	return nil
}

// =============================================================================
// Test Utilities
// =============================================================================

// NewTestError creates a new Error for testing purposes.
// This function is primarily intended for use in tests.
func NewTestError(status int, code, message string) *Error {
	return &Error{
		Status:  status,
		Code:    code,
		Message: message,
		Data:    make(map[string]FieldError),
	}
}

// NewTestValidationError creates a new validation Error for testing purposes.
// This function is primarily intended for use in tests.
func NewTestValidationError(fieldErrors map[string]FieldError) *Error {
	return &Error{
		Status:  http.StatusBadRequest,
		Message: "Validation failed",
		Data:    fieldErrors,
	}
}

// Equals compares two Error instances for equality, ignoring debugging fields.
// This function is primarily intended for use in tests.
func (e *Error) Equals(other *Error) bool {
	if e == nil && other == nil {
		return true
	}
	if e == nil || other == nil {
		return false
	}

	if e.Status != other.Status || e.Code != other.Code || e.Message != other.Message {
		return false
	}

	if len(e.Data) != len(other.Data) {
		return false
	}

	for k, v := range e.Data {
		if otherV, ok := other.Data[k]; !ok || v != otherV {
			return false
		}
	}

	// Debug fields are intentionally ignored for equality comparison
	return true
}

// LogFields returns structured fields for logging purposes.
// This provides a logging-friendly representation of the error.
func (e *Error) LogFields() map[string]any {
	if e == nil {
		return map[string]any{"error": "nil"}
	}

	fields := map[string]any{
		"status":  e.Status,
		"message": e.Message,
	}

	if e.Code != "" {
		fields["code"] = e.Code
	}

	if e.Debug != nil && e.Debug.Endpoint != "" {
		fields["endpoint"] = e.Debug.Endpoint
	}

	if len(e.Data) > 0 {
		fields["field_errors"] = len(e.Data)
	}

	return fields
}
