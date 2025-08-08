package pocketbase

import (
	"errors"
	"fmt"
	"net/http"
	"strings"
	"sync"
	"testing"
)

func TestError_Error(t *testing.T) {
	err := &Error{
		Status:  404,
		Code:    "resource_not_found",
		Message: "Not Found",
		Data:    map[string]FieldError{},
	}

	expected := "pocketbase: 404 Not Found code=resource_not_found msg=Not Found"
	if err.Error() != expected {
		t.Errorf("Expected error message '%s', got '%s'", expected, err.Error())
	}
}

func TestError_IsValidation(t *testing.T) {
	// Validation error (400 with field errors)
	validationErr := &Error{
		Status: 400,
		Data: map[string]FieldError{
			"email": {Code: "required", Message: "Email is required"},
		},
	}

	if !validationErr.IsValidation() {
		t.Error("Expected IsValidation() to return true for validation error")
	}

	// Non-validation error (400 without field errors)
	badRequestErr := &Error{
		Status: 400,
		Data:   map[string]FieldError{},
	}

	if badRequestErr.IsValidation() {
		t.Error("Expected IsValidation() to return false for non-validation error")
	}
}

func TestError_IsAuth(t *testing.T) {
	authErr := &Error{
		Status: 401,
	}

	if !authErr.IsAuth() {
		t.Error("Expected IsAuth() to return true for auth error")
	}

	notAuthErr := &Error{
		Status: 404,
	}

	if notAuthErr.IsAuth() {
		t.Error("Expected IsAuth() to return false for non-auth error")
	}
}

func TestParseAPIError(t *testing.T) {
	// Test successful response (no error)
	successResp := &http.Response{StatusCode: 200}
	err := ParseAPIError(successResp, []byte("{}"), "/test")
	if err != nil {
		t.Errorf("Expected no error for successful response, got %v", err)
	}

	// Test error response
	errorResp := &http.Response{
		StatusCode: 404,
		Header:     make(http.Header),
	}
	body := []byte(`{"code": 404, "message": "Record not found."}`)

	err = ParseAPIError(errorResp, body, "/api/collections/users/records/123")
	if err == nil {
		t.Fatal("Expected error for 404 response")
	}

	var pbErr *Error
	if !errors.As(err, &pbErr) {
		t.Fatalf("Expected *Error, got %T", err)
	}

	if pbErr.Status != 404 {
		t.Errorf("Expected status 404, got %d", pbErr.Status)
	}

	if pbErr.Code != "record_not_found" {
		t.Errorf("Expected code 'record_not_found', got '%s'", pbErr.Code)
	}

	if !pbErr.IsNotFound() {
		t.Error("Expected error to be a not found error")
	}
}

func TestHelperFunctions(t *testing.T) {
	// Test IsNotFoundError
	notFoundErr := &Error{Status: 404}
	if !IsNotFoundError(notFoundErr) {
		t.Error("Expected IsNotFoundError to return true")
	}

	// Test IsAuthError
	authErr := &Error{Status: 401}
	if !IsAuthError(authErr) {
		t.Error("Expected IsAuthError to return true")
	}

	// Test IsValidationError
	validationErr := &Error{
		Status: 400,
		Data:   map[string]FieldError{"email": {Code: "required", Message: "required"}},
	}
	if !IsValidationError(validationErr) {
		t.Error("Expected IsValidationError to return true")
	}

	// Test with non-PocketBase error
	genericErr := errors.New("generic error")
	if IsNotFoundError(genericErr) {
		t.Error("Expected IsNotFoundError to return false for generic error")
	}
}

func TestGetFieldErrors(t *testing.T) {
	validationErr := &Error{
		Status: 400,
		Data: map[string]FieldError{
			"email": {Code: "required", Message: "Email is required"},
			"name":  {Code: "invalid", Message: "Name is invalid"},
		},
	}

	fieldErrors := GetFieldErrors(validationErr)
	if fieldErrors == nil {
		t.Fatal("Expected field errors, got nil")
	}

	if len(fieldErrors) != 2 {
		t.Errorf("Expected 2 field errors, got %d", len(fieldErrors))
	}

	if fieldErrors["email"].Code != "required" {
		t.Errorf("Expected email error code 'required', got '%s'", fieldErrors["email"].Code)
	}

	// Test with non-validation error
	nonValidationErr := &Error{Status: 404}
	fieldErrors = GetFieldErrors(nonValidationErr)
	if fieldErrors != nil {
		t.Error("Expected nil field errors for non-validation error")
	}
}

func TestHasErrorCode(t *testing.T) {
	err := &Error{
		Status: 404,
		Code:   "collection_not_found",
	}

	if !HasErrorCode(err, "collection_not_found") {
		t.Error("Expected HasErrorCode to return true for matching code")
	}

	if HasErrorCode(err, "record_not_found") {
		t.Error("Expected HasErrorCode to return false for non-matching code")
	}

	// Test with non-PocketBase error
	genericErr := errors.New("generic error")
	if HasErrorCode(genericErr, "any_code") {
		t.Error("Expected HasErrorCode to return false for generic error")
	}
}
func TestHTTPStatusErrors(t *testing.T) {
	// Test errors.Is with HTTPStatus values
	notFoundErr := &Error{Status: 404}
	if !errors.Is(notFoundErr, StatusNotFound) {
		t.Error("Expected errors.Is to return true for StatusNotFound")
	}

	authErr := &Error{Status: 401}
	if !errors.Is(authErr, StatusUnauthorized) {
		t.Error("Expected errors.Is to return true for StatusUnauthorized")
	}

	// Test with http package constants
	if !errors.Is(notFoundErr, HTTPStatus(http.StatusNotFound)) {
		t.Error("Expected errors.Is to return true for HTTPStatus(http.StatusNotFound)")
	}

	// Test that different statuses don't match
	if errors.Is(notFoundErr, StatusUnauthorized) {
		t.Error("Expected errors.Is to return false for different status codes")
	}

	// Test with generic error
	genericErr := errors.New("generic error")
	if errors.Is(genericErr, StatusNotFound) {
		t.Error("Expected errors.Is to return false for generic error")
	}
}

func TestErrorIs_ByCode(t *testing.T) {
	// Test matching by error code
	err := &Error{
		Status: 404,
		Code:   "collection_not_found",
	}

	// Create a target error with the same code
	target := &Error{Code: "collection_not_found"}

	if !errors.Is(err, target) {
		t.Error("Expected errors.Is to return true for matching error code")
	}

	// Test with different code
	differentTarget := &Error{Code: "record_not_found"}
	if errors.Is(err, differentTarget) {
		t.Error("Expected errors.Is to return false for different error code")
	}
}

// Additional tests

func TestError_AllMethods(t *testing.T) {
	tests := []struct {
		name     string
		err      *Error
		expected map[string]bool
	}{
		{
			name: "BadRequest validation error",
			err: &Error{
				Status: 400,
				Data:   map[string]FieldError{"email": {Code: "required", Message: "required"}},
			},
			expected: map[string]bool{
				"IsValidation":  true,
				"IsAuth":        false,
				"IsForbidden":   false,
				"IsNotFound":    false,
				"IsRateLimited": false,
				"IsInternal":    false,
			},
		},
		{
			name: "Unauthorized error",
			err: &Error{
				Status: 401,
			},
			expected: map[string]bool{
				"IsValidation":  false,
				"IsAuth":        true,
				"IsForbidden":   false,
				"IsNotFound":    false,
				"IsRateLimited": false,
				"IsInternal":    false,
			},
		},
		{
			name: "Forbidden error",
			err: &Error{
				Status: 403,
			},
			expected: map[string]bool{
				"IsValidation":  false,
				"IsAuth":        false,
				"IsForbidden":   true,
				"IsNotFound":    false,
				"IsRateLimited": false,
				"IsInternal":    false,
			},
		},
		{
			name: "NotFound error",
			err: &Error{
				Status: 404,
			},
			expected: map[string]bool{
				"IsValidation":  false,
				"IsAuth":        false,
				"IsForbidden":   false,
				"IsNotFound":    true,
				"IsRateLimited": false,
				"IsInternal":    false,
			},
		},
		{
			name: "RateLimited error",
			err: &Error{
				Status: 429,
			},
			expected: map[string]bool{
				"IsValidation":  false,
				"IsAuth":        false,
				"IsForbidden":   false,
				"IsNotFound":    false,
				"IsRateLimited": true,
				"IsInternal":    false,
			},
		},
		{
			name: "Internal error",
			err: &Error{
				Status: 500,
			},
			expected: map[string]bool{
				"IsValidation":  false,
				"IsAuth":        false,
				"IsForbidden":   false,
				"IsNotFound":    false,
				"IsRateLimited": false,
				"IsInternal":    true,
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.err.IsValidation() != tt.expected["IsValidation"] {
				t.Errorf("IsValidation() = %v, want %v", tt.err.IsValidation(), tt.expected["IsValidation"])
			}
			if tt.err.IsAuth() != tt.expected["IsAuth"] {
				t.Errorf("IsAuth() = %v, want %v", tt.err.IsAuth(), tt.expected["IsAuth"])
			}
			if tt.err.IsForbidden() != tt.expected["IsForbidden"] {
				t.Errorf("IsForbidden() = %v, want %v", tt.err.IsForbidden(), tt.expected["IsForbidden"])
			}
			if tt.err.IsNotFound() != tt.expected["IsNotFound"] {
				t.Errorf("IsNotFound() = %v, want %v", tt.err.IsNotFound(), tt.expected["IsNotFound"])
			}
			if tt.err.IsRateLimited() != tt.expected["IsRateLimited"] {
				t.Errorf("IsRateLimited() = %v, want %v", tt.err.IsRateLimited(), tt.expected["IsRateLimited"])
			}
			if tt.err.IsInternal() != tt.expected["IsInternal"] {
				t.Errorf("IsInternal() = %v, want %v", tt.err.IsInternal(), tt.expected["IsInternal"])
			}
		})
	}
}

func TestAllHTTPStatusErrors(t *testing.T) {
	tests := []struct {
		name   string
		err    *Error
		status HTTPStatus
	}{
		{"BadRequest", &Error{Status: 400}, StatusBadRequest},
		{"Unauthorized", &Error{Status: 401}, StatusUnauthorized},
		{"Forbidden", &Error{Status: 403}, StatusForbidden},
		{"NotFound", &Error{Status: 404}, StatusNotFound},
		{"PayloadTooLarge", &Error{Status: 413}, StatusRequestEntityTooLarge},
		{"TooManyRequests", &Error{Status: 429}, StatusTooManyRequests},
		{"Internal", &Error{Status: 500}, StatusInternalServerError},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if !errors.Is(tt.err, tt.status) {
				t.Errorf("errors.Is(%v, %v) = false, want true", tt.err, tt.status)
			}
		})
	}
}

func TestGetErrorCode(t *testing.T) {
	// Test with PocketBase error
	pbErr := &Error{Code: "collection_not_found"}
	if code := GetErrorCode(pbErr); code != "collection_not_found" {
		t.Errorf("Expected code 'collection_not_found', got '%s'", code)
	}

	// Test with generic error
	genericErr := errors.New("generic error")
	if code := GetErrorCode(genericErr); code != "" {
		t.Errorf("Expected empty code for generic error, got '%s'", code)
	}

	// Test with nil
	if code := GetErrorCode(nil); code != "" {
		t.Errorf("Expected empty code for nil error, got '%s'", code)
	}
}

func TestParseAPIError_EdgeCases(t *testing.T) {
	// Test with nil response
	err := ParseAPIError(nil, []byte("{}"), "/test")
	if err == nil {
		t.Error("Expected error for nil response")
	}

	// Test with invalid JSON
	resp := &http.Response{
		StatusCode: 400,
		Header:     make(http.Header),
	}
	invalidJSON := []byte(`{"invalid": json}`)

	err = ParseAPIError(resp, invalidJSON, "/test")
	if err == nil {
		t.Error("Expected error for invalid JSON")
	}

	var pbErr *Error
	if !errors.As(err, &pbErr) {
		t.Fatalf("Expected *Error, got %T", err)
	}

	if pbErr.Status != 400 {
		t.Errorf("Expected status 400, got %d", pbErr.Status)
	}
}

func TestRegisterMessageAlias(t *testing.T) {
	// Backup original alias
	originalAlias := messageAliases["Test message"]

	// Register new alias
	RegisterMessageAlias("Test message", "test_alias")

	if messageAliases["Test message"] != "test_alias" {
		t.Error("RegisterMessageAlias did not set the alias correctly")
	}

	// Cleanup (restore original state)
	if originalAlias == "" {
		delete(messageAliases, "Test message")
	} else {
		messageAliases["Test message"] = originalAlias
	}
}

func TestError_ErrorMessage_EdgeCases(t *testing.T) {
	// Test nil error
	var nilErr *Error
	if nilErr.Error() != "<nil>" {
		t.Errorf("Expected '<nil>' for nil error, got '%s'", nilErr.Error())
	}

	// Test error with endpoint
	err := &Error{
		Status:   404,
		Code:     "not_found",
		Message:  "Not Found",
		Endpoint: "/api/test",
	}

	expected := "pocketbase: 404 Not Found code=not_found msg=Not Found at=/api/test"
	if err.Error() != expected {
		t.Errorf("Expected '%s', got '%s'", expected, err.Error())
	}

	// Test error with field data
	errWithData := &Error{
		Status:  400,
		Message: "Validation failed",
		Data: map[string]FieldError{
			"email": {Code: "required", Message: "required"},
			"name":  {Code: "invalid", Message: "invalid"},
		},
	}

	result := errWithData.Error()
	if !strings.Contains(result, "data=2 field error(s)") {
		t.Errorf("Expected error message to contain field count, got '%s'", result)
	}
}

// Tests for new features

func TestHTTPStatus(t *testing.T) {
	// Test HTTPStatus Error method
	status := HTTPStatus(404)
	expected := "HTTP 404 Not Found"
	if status.Error() != expected {
		t.Errorf("HTTPStatus.Error() = %q, want %q", status.Error(), expected)
	}

	// Test unknown status
	unknownStatus := HTTPStatus(999)
	expected = "HTTP 999"
	if unknownStatus.Error() != expected {
		t.Errorf("HTTPStatus.Error() = %q, want %q", unknownStatus.Error(), expected)
	}
}

func TestHasHTTPStatus(t *testing.T) {
	err := &Error{Status: 404}

	// Test with http package constant
	if !HasHTTPStatus(err, http.StatusNotFound) {
		t.Error("Expected HasHTTPStatus to return true for http.StatusNotFound")
	}

	// Test with different status
	if HasHTTPStatus(err, http.StatusUnauthorized) {
		t.Error("Expected HasHTTPStatus to return false for different status")
	}

	// Test with non-PocketBase error
	genericErr := errors.New("generic error")
	if HasHTTPStatus(genericErr, http.StatusNotFound) {
		t.Error("Expected HasHTTPStatus to return false for generic error")
	}
}

func TestGetHTTPStatus(t *testing.T) {
	// Test with PocketBase error
	pbErr := &Error{Status: 404}
	if status := GetHTTPStatus(pbErr); status != 404 {
		t.Errorf("Expected status 404, got %d", status)
	}

	// Test with generic error
	genericErr := errors.New("generic error")
	if status := GetHTTPStatus(genericErr); status != 0 {
		t.Errorf("Expected status 0 for generic error, got %d", status)
	}

	// Test with nil
	if status := GetHTTPStatus(nil); status != 0 {
		t.Errorf("Expected status 0 for nil error, got %d", status)
	}
}

func TestNewTestError(t *testing.T) {
	err := NewTestError(404, "not_found", "Resource not found")

	if err.Status != 404 {
		t.Errorf("Expected status 404, got %d", err.Status)
	}
	if err.Code != "not_found" {
		t.Errorf("Expected code 'not_found', got %q", err.Code)
	}
	if err.Message != "Resource not found" {
		t.Errorf("Expected message 'Resource not found', got %q", err.Message)
	}
	if err.Data == nil {
		t.Error("Expected Data to be initialized")
	}
}

func TestNewTestValidationError(t *testing.T) {
	fieldErrors := map[string]FieldError{
		"email": {Code: "required", Message: "Email is required"},
		"name":  {Code: "invalid", Message: "Name is invalid"},
	}

	err := NewTestValidationError(fieldErrors)

	if err.Status != 400 {
		t.Errorf("Expected status 400, got %d", err.Status)
	}
	if !err.IsValidation() {
		t.Error("Expected error to be a validation error")
	}
	if len(err.Data) != 2 {
		t.Errorf("Expected 2 field errors, got %d", len(err.Data))
	}
}

func TestErrorEquals(t *testing.T) {
	err1 := &Error{
		Status:  404,
		Code:    "not_found",
		Message: "Not found",
		Data:    map[string]FieldError{"field": {Code: "invalid", Message: "Invalid"}},
	}

	err2 := &Error{
		Status:  404,
		Code:    "not_found",
		Message: "Not found",
		Data:    map[string]FieldError{"field": {Code: "invalid", Message: "Invalid"}},
		// Different debugging fields should be ignored
		Endpoint: "/different/endpoint",
	}

	if !err1.Equals(err2) {
		t.Error("Expected errors to be equal")
	}

	// Test with different core fields
	err3 := &Error{
		Status:  500,
		Code:    "not_found",
		Message: "Not found",
		Data:    map[string]FieldError{"field": {Code: "invalid", Message: "Invalid"}},
	}

	if err1.Equals(err3) {
		t.Error("Expected errors to be different")
	}

	// Test with nil
	if err1.Equals(nil) {
		t.Error("Expected error to not equal nil")
	}

	var nilErr *Error
	if !nilErr.Equals(nil) {
		t.Error("Expected nil errors to be equal")
	}
}

func TestRegisterMessageAlias_ThreadSafety(t *testing.T) {
	// Test concurrent access to RegisterMessageAlias
	const numGoroutines = 10
	const numOperations = 100

	var wg sync.WaitGroup
	wg.Add(numGoroutines)

	for i := 0; i < numGoroutines; i++ {
		go func(id int) {
			defer wg.Done()
			for j := 0; j < numOperations; j++ {
				message := fmt.Sprintf("Test message %d-%d", id, j)
				alias := fmt.Sprintf("test_alias_%d_%d", id, j)
				RegisterMessageAlias(message, alias)

				// Test retrieval
				if _, ok := getMessageAlias(message); !ok {
					t.Errorf("Failed to retrieve alias for message: %s", message)
				}
			}
		}(i)
	}

	wg.Wait()
}

func TestHelperFunctionsWithErrorsIs(t *testing.T) {
	// Test that helper functions work the same as errors.Is
	notFoundErr := &Error{Status: 404}

	// Both should return the same result
	if IsNotFoundError(notFoundErr) != errors.Is(notFoundErr, StatusNotFound) {
		t.Error("IsNotFoundError and errors.Is should return the same result")
	}

	authErr := &Error{Status: 401}
	if IsAuthError(authErr) != errors.Is(authErr, StatusUnauthorized) {
		t.Error("IsAuthError and errors.Is should return the same result")
	}

	badRequestErr := &Error{Status: 400}
	if IsBadRequestError(badRequestErr) != errors.Is(badRequestErr, StatusBadRequest) {
		t.Error("IsBadRequestError and errors.Is should return the same result")
	}
}

func TestErrorNilSafety(t *testing.T) {
	var nilErr *Error

	// Test all methods are nil-safe
	if nilErr.IsValidation() {
		t.Error("Expected nil error IsValidation to return false")
	}
	if nilErr.IsAuth() {
		t.Error("Expected nil error IsAuth to return false")
	}
	if nilErr.IsForbidden() {
		t.Error("Expected nil error IsForbidden to return false")
	}
	if nilErr.IsNotFound() {
		t.Error("Expected nil error IsNotFound to return false")
	}
	if nilErr.IsRateLimited() {
		t.Error("Expected nil error IsRateLimited to return false")
	}
	if nilErr.IsInternal() {
		t.Error("Expected nil error IsInternal to return false")
	}
	if nilErr.IsBadRequest() {
		t.Error("Expected nil error IsBadRequest to return false")
	}

	// Test Is method
	if nilErr.Is(StatusNotFound) {
		t.Error("Expected nil error Is to return false")
	}
	if !nilErr.Is(nil) {
		t.Error("Expected nil error to equal nil")
	}
}
