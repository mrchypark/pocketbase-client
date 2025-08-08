// filename: errors_ext.go
// Package pocketbase provides extended error handling utilities for PocketBase API interactions.

// This file defines a richer error type (`Error`) than the basic `APIError`.
// It classifies errors based on HTTP status codes, parses structured JSON error bodies,
// and normalizes server error messages to stable alias codes. The alias set is derived
// from the PocketBase documentation for main and v0.22.x releases.

// Consumers of this client can use the `ParseAPIError` helper to convert an HTTP response
// into an `*Error`, then use classification helpers such as `IsValidation()` or
// `IsAuth()` to branch on the category of failure.

package pocketbase

import (
    "encoding/json"
    "errors"
    "fmt"
    "net/http"
    "strconv"
    "strings"
    "time"
)

// ErrorKind enumerates coarse categories of errors encountered when calling the PocketBase API.
// The values are derived from common HTTP status codes.
type ErrorKind int

const (
    // ErrUnknown is returned when the error kind cannot be determined.
    ErrUnknown ErrorKind = iota
    // ErrBadRequest corresponds to 400 Bad Request responses.
    ErrBadRequest
    // ErrUnauthorized corresponds to 401 Unauthorized responses.
    ErrUnauthorized
    // ErrForbidden corresponds to 403 Forbidden responses.
    ErrForbidden
    // ErrNotFound corresponds to 404 Not Found responses.
    ErrNotFound
    // ErrPayloadTooLarge corresponds to 413 Request Entity Too Large responses.
    ErrPayloadTooLarge
    // ErrTooManyRequests corresponds to 429 Too Many Requests responses.
    ErrTooManyRequests
    // ErrInternal corresponds to 500 Internal Server Error responses.
    ErrInternal
)

// FieldError captures validation failures for a specific field in the PocketBase API.
// The Code is a machine-readable identifier and Message is a human-readable explanation.
type FieldError struct {
    Code    string `json:"code"`
    Message string `json:"message"`
}

// Error represents a normalized PocketBase API error. It contains the HTTP status, a generic
// message, optional per-field errors, and derived metadata such as the error kind and retry hints.
type Error struct {
    // Status is the HTTP status code.
    Status int `json:"status"`
    // Code is an optional alias derived from the server message (see messageAliases).
    Code string `json:"code"`
    // Message is the top-level error message returned by the server.
    Message string `json:"message"`
    // Data contains per-field validation errors, keyed by field name.
    Data map[string]FieldError `json:"data"`
    // Kind classifies the error based on Status.
    Kind ErrorKind `json:"-"`
    // Endpoint optionally records the API endpoint used, for debugging.
    Endpoint string `json:"-"`
    // RetryAfter contains the duration to wait before retrying, if provided via Retry-After header.
    RetryAfter *time.Duration `json:"-"`
    // RawHeaders captures the original response headers.
    RawHeaders http.Header `json:"-"`
    // RawBody captures the original response body bytes.
    RawBody []byte `json:"-"`
}

// Error implements the error interface for *Error.
func (e *Error) Error() string {
    if e == nil {
        return "<nil>"
    }
    base := fmt.Sprintf("pocketbase: %d %s", e.Status, http.StatusText(e.Status))
    if e.Code != "" {
        base += " code=" + e.Code
    }
    if e.Message != "" {
        base += " msg=" + e.Message
    }
    if len(e.Data) > 0 {
        base += fmt.Sprintf(" data=%d field error(s)", len(e.Data))
    }
    if e.Endpoint != "" {
        base += " at=" + e.Endpoint
    }
    return base
}

// Unwrap satisfies errors.Unwrap but returns nil because Error is the root error type.
func (e *Error) Unwrap() error { return nil }

// IsValidation reports whether the error represents a validation failure (HTTP 400 with field errors).
func (e *Error) IsValidation() bool { return e.Kind == ErrBadRequest && len(e.Data) > 0 }

// IsAuth reports whether the error indicates authentication failure (401).
func (e *Error) IsAuth() bool { return e.Kind == ErrUnauthorized }

// IsForbidden reports whether the error indicates an authorization failure (403).
func (e *Error) IsForbidden() bool { return e.Kind == ErrForbidden }

// IsNotFound reports whether the error corresponds to a missing resource (404).
func (e *Error) IsNotFound() bool { return e.Kind == ErrNotFound }

// IsRateLimited reports whether the error corresponds to a rate limit violation (429).
func (e *Error) IsRateLimited() bool { return e.Kind == ErrTooManyRequests }

// IsInternal reports whether the error corresponds to an internal server error (500).
func (e *Error) IsInternal() bool { return e.Kind == ErrInternal }

// Retryable indicates whether the error could succeed if retried after a delay.
// Currently only rate limited errors are considered retryable.
func (e *Error) Retryable() bool { return e.IsRateLimited() }

// RetryAfterDuration returns the suggested backoff duration for retryable errors.
func (e *Error) RetryAfterDuration() time.Duration {
    if e.RetryAfter != nil {
        return *e.RetryAfter
    }
    return 0
}

// rawPocketBaseError matches the JSON structure returned by the PocketBase API.
// It is used internally to decode error bodies.
type rawPocketBaseError struct {
    Code    int `json:"code"`
    Message string `json:"message"`
    Data    map[string]struct {
        Code    string `json:"code"`
        Message string `json:"message"`
    } `json:"data"`
}

// ParseAPIError converts an HTTP response and body into an *Error. If the response
// status code is < 400, nil is returned.
func ParseAPIError(resp *http.Response, body []byte, endpoint string) error {
    if resp == nil {
        return errors.New("pocketbase: nil http.Response")
    }
    if resp.StatusCode < 400 {
        return nil
    }
    kind := classifyStatus(resp.StatusCode)
    var wire rawPocketBaseError
    _ = json.Unmarshal(body, &wire) // tolerate invalid JSON
    e := &Error{
        Status:     resp.StatusCode,
        Message:    strings.TrimSpace(wire.Message),
        Data:       map[string]FieldError{},
        Kind:       kind,
        Endpoint:   endpoint,
        RawHeaders: resp.Header.Clone(),
        RawBody:    append([]byte(nil), body...),
    }
    for k, v := range wire.Data {
        e.Data[k] = FieldError{Code: v.Code, Message: v.Message}
    }
    // For rate limited responses, parse Retry-After header.
    if resp.StatusCode == http.StatusTooManyRequests {
        if d := parseRetryAfter(resp.Header.Get("Retry-After")); d > 0 {
            e.RetryAfter = &d
        }
    }
    applyKnownAliases(e)
    return e
}

// classifyStatus maps HTTP status codes to ErrorKind values.
func classifyStatus(s int) ErrorKind {
    switch s {
    case 400:
        return ErrBadRequest
    case 401:
        return ErrUnauthorized
    case 403:
        return ErrForbidden
    case 404:
        return ErrNotFound
    case 413:
        return ErrPayloadTooLarge
    case 429:
        return ErrTooManyRequests
    case 500:
        return ErrInternal
    default:
        return ErrUnknown
    }
}

// parseRetryAfter parses the Retry-After header value, supporting both seconds and HTTP-date.
func parseRetryAfter(v string) time.Duration {
    if v == "" {
        return 0
    }
    // seconds format
    if secs, err := strconv.Atoi(v); err == nil && secs >= 0 {
        return time.Duration(secs) * time.Second
    }
    // HTTP-date format
    if t, err := http.ParseTime(v); err == nil {
        d := time.Until(t)
        if d < 0 {
            return 0
        }
        return d
    }
    return 0
}

// messageAliases maps server-provided error messages to stable alias codes.
// Applications may register additional aliases at runtime.
var messageAliases = map[string]string{
    "Unsupported Content-Type":                         "unsupported_content_type",
    "The request can be accessed only by guests.":      "only_guests",
    "The request can be accessed only by authenticated admins.":  "only_admins",
    "The request can be accessed only by authenticated records.": "only_records",
    "The request requires valid admin authorization token.":      "require_admin_token",
    "The request requires valid record authorization token.":     "require_record_token",
    "Missing or invalid collection context.":           "collection_context_invalid",
    "Missing required \"expand\" parameter.":          "missing_expand_param",
    "Missing required \"filter\" parameter.":          "missing_filter_param",
    "Missing required \"fields\" parameter.":          "missing_fields_param",
    "Missing required \"sort\" parameter.":            "missing_sort_param",
    "Invalid \"sort\" parameter format.":              "invalid_sort_param",
    "Invalid \"expand\" parameter format.":            "invalid_expand_param",
    "Invalid \"filter\" parameter format.":            "invalid_filter_param",
    "Invalid \"fields\" parameter format.":            "invalid_fields_param",
    "Invalid \"page\" parameter format.":              "invalid_page_param",
    "Invalid \"perPage\" parameter format.":           "invalid_perpage_param",
    "Invalid \"skipTotal\" parameter format.":         "invalid_skiptotal_param",
    "Invalid request payload.":                        "invalid_request_payload",
    "Invalid or missing file.":                        "invalid_or_missing_file",
    "Invalid request body.":                           "invalid_request_body",
    "Invalid or missing field":                        "invalid_or_missing_field",
    "Invalid or missing record id.":                   "invalid_record_id",
    "Invalid or missing password reset token.":        "invalid_password_reset_token",
    "Invalid or missing verification token.":          "invalid_verification_token",
    "Invalid or missing file token.":                  "invalid_file_token",
    "Invalid or missing OAuth2 state parameter.":      "invalid_oauth2_state",
    "Invalid or missing redirect URL.":                "invalid_redirect_url",
    "Invalid redirect status code.":                   "invalid_redirect_status_code",
    "Failed to fetch admins info.":                    "failed_fetch_admins",
    "Failed to fetch records info.":                   "failed_fetch_records",
    "Failed to fetch collection schema.":              "failed_fetch_schema",
    "Missing or invalid authentication.":              "invalid_authentication",
    "Missing or invalid authentication token.":        "invalid_auth_token",
    "Missing or invalid admin authorization token.":   "invalid_admin_auth_token",
    "Missing or invalid record authorization token.":  "invalid_record_auth_token",
    "You are not allowed to perform this request.":    "forbidden_generic",
    "The authorized record is not allowed to perform this action.": "record_forbidden",
    "The authorized admin is not allowed to perform this action.":  "admin_forbidden",
    "The requested resource wasn't found.":            "resource_not_found",
    "File not found.":                                 "file_not_found",
    "Collection not found.":                           "collection_not_found",
    "Record not found.":                               "record_not_found",
    "Request entity too large":                        "entity_too_large",
    "Too Many Requests.":                              "too_many_requests",
    "Something went wrong while processing your request.": "internal_generic",
}

// RegisterMessageAlias adds or overrides a server message → alias mapping.
// Applications may call this during initialization to customize alias names.
func RegisterMessageAlias(serverMessage, alias string) {
    messageAliases[serverMessage] = alias
}

// applyKnownAliases stores the alias derived from messageAliases back into e.Code.
func applyKnownAliases(e *Error) {
    if e == nil {
        return
    }
    if alias, ok := messageAliases[e.Message]; ok {
        e.Code = alias
    }
}

// init registers the default alias set. It runs automatically when this package is imported.
func init() {
    for msg, alias := range messageAliases {
        RegisterMessageAlias(msg, alias)
    }
}
