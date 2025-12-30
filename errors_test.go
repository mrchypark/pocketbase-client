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
	err := ParseAPIErrorFromResponse(successResp, []byte("{}"))
	if err != nil {
		t.Errorf("Expected no error for successful response, got %v", err)
	}

	// Test error response
	errorResp := &http.Response{
		StatusCode: 404,
		Header:     make(http.Header),
	}
	body := []byte(`{"code": 404, "message": "Record not found."}`)

	err = ParseAPIErrorFromResponse(errorResp, body)
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

func TestParseAPIError_Direct(t *testing.T) {
	// Test successful response (no error)
	err := ParseAPIError(200, []byte("{}"))
	if err != nil {
		t.Errorf("Expected no error for successful response, got %v", err)
	}

	// Test error response
	headers := make(http.Header)
	headers.Set("Content-Type", "application/json")
	body := []byte(`{"code": 404, "message": "Record not found.", "data": {"id": {"code": "validation_required", "message": "Missing required value."}}}`)

	err = ParseAPIError(404, body)
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
	if pbErr.Message != "Record not found." {
		t.Errorf("Expected message 'Record not found.', got '%s'", pbErr.Message)
	}

	// Test field errors
	if len(pbErr.Data) != 1 {
		t.Errorf("Expected 1 field error, got %d", len(pbErr.Data))
	}
	if fieldErr, ok := pbErr.Data["id"]; ok {
		if fieldErr.Code != "validation_required" {
			t.Errorf("Expected field error code 'validation_required', got '%s'", fieldErr.Code)
		}
		if fieldErr.Message != "Missing required value." {
			t.Errorf("Expected field error message 'Missing required value.', got '%s'", fieldErr.Message)
		}
	} else {
		t.Error("Expected field error for 'id' field")
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
	err := ParseAPIErrorFromResponse(nil, []byte("{}"))
	if err == nil {
		t.Error("Expected error for nil response")
	}

	// Test with invalid JSON
	resp := &http.Response{
		StatusCode: 400,
		Header:     make(http.Header),
	}
	invalidJSON := []byte(`{"invalid": json}`)

	err = ParseAPIErrorFromResponse(resp, invalidJSON)
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

func TestNewErrorAliases(t *testing.T) {
	tests := []struct {
		message      string
		expectedCode string
	}{
		{"You are not allowed to perform this action.", "action_forbidden"},
		{"The new email address is already in use.", "email_already_in_use"},
		{"The provided old password is not valid.", "invalid_old_password"},
	}

	for _, tt := range tests {
		t.Run(tt.expectedCode, func(t *testing.T) {
			// Create error with the message
			err := &Error{
				Status:  400,
				Message: tt.message,
				Data:    make(map[string]FieldError),
			}

			// Apply aliases
			applyKnownAliases(err)

			if err.Code != tt.expectedCode {
				t.Errorf("Expected code '%s', got '%s'", tt.expectedCode, err.Code)
			}
		})
	}
}

func TestError_ErrorMessage_EdgeCases(t *testing.T) {
	// Test nil error
	var nilErr *Error
	if nilErr.Error() != "<nil>" {
		t.Errorf("Expected '<nil>' for nil error, got '%s'", nilErr.Error())
	}

	// Test basic error without endpoint
	err := &Error{
		Status:  404,
		Code:    "not_found",
		Message: "Not Found",
	}

	expected := "pocketbase: 404 Not Found code=not_found msg=Not Found"
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

	for i := range numGoroutines {
		go func(id int) {
			defer wg.Done()
			for j := range numOperations {
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

func TestGetMessageAlias_CacheBehavior(t *testing.T) {
	// Clear cache for this test
	aliasCache.mu.Lock()
	aliasCache.cache = make(map[string]string)
	aliasCache.mu.Unlock()

	// Test message that doesn't exist
	nonExistentMessage := "This message does not exist in aliases"

	// First call should return false
	alias1, found1 := getMessageAlias(nonExistentMessage)
	if alias1 != "" || found1 {
		t.Errorf("First call: expected ('', false), got ('%s', %v)", alias1, found1)
	}

	// Second call should also return false (cached miss)
	alias2, found2 := getMessageAlias(nonExistentMessage)
	if alias2 != "" || found2 {
		t.Errorf("Second call: expected ('', false), got ('%s', %v)", alias2, found2)
	}

	// Test message that exists
	existingMessage := "Record not found."
	expectedAlias := "record_not_found"

	// First call should return true
	alias3, found3 := getMessageAlias(existingMessage)
	if alias3 != expectedAlias || !found3 {
		t.Errorf("First call for existing: expected ('%s', true), got ('%s', %v)", expectedAlias, alias3, found3)
	}

	// Second call should also return true (cached hit)
	alias4, found4 := getMessageAlias(existingMessage)
	if alias4 != expectedAlias || !found4 {
		t.Errorf("Second call for existing: expected ('%s', true), got ('%s', %v)", expectedAlias, alias4, found4)
	}
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

func TestError_LogFields(t *testing.T) {
	// Test nil error
	var nilErr *Error
	fields := nilErr.LogFields()
	if fields["error"] != "nil" {
		t.Errorf("Expected nil error to have 'error': 'nil', got %v", fields)
	}

	// Test basic error
	err := &Error{
		Status:  404,
		Code:    "not_found",
		Message: "Resource not found",
	}

	fields = err.LogFields()
	expectedFields := map[string]any{
		"status":  404,
		"code":    "not_found",
		"message": "Resource not found",
	}

	for key, expected := range expectedFields {
		if fields[key] != expected {
			t.Errorf("Expected field %s to be %v, got %v", key, expected, fields[key])
		}
	}

	// Test error with field errors
	errWithData := &Error{
		Status:  400,
		Message: "Validation failed",
		Data: map[string]FieldError{
			"email": {Code: "required", Message: "Email is required"},
			"name":  {Code: "invalid", Message: "Name is invalid"},
		},
	}

	fields = errWithData.LogFields()
	if fields["status"] != 400 {
		t.Errorf("Expected status 400, got %v", fields["status"])
	}
	if fields["message"] != "Validation failed" {
		t.Errorf("Expected message 'Validation failed', got %v", fields["message"])
	}
	if fields["field_errors"] != 2 {
		t.Errorf("Expected field_errors 2, got %v", fields["field_errors"])
	}

	// Test error without optional fields
	simpleErr := &Error{
		Status:  500,
		Message: "Internal error",
	}

	fields = simpleErr.LogFields()
	if _, hasCode := fields["code"]; hasCode {
		t.Error("Expected no code field for error without code")
	}
	if _, hasFieldErrors := fields["field_errors"]; hasFieldErrors {
		t.Error("Expected no field_errors field for error without field errors")
	}
}
func TestValidationErrorCodes(t *testing.T) {
	// Test that validation error codes are properly handled
	validationCodes := []string{
		"validation_is_required",
		"validation_nil_or_not_empty",
		"validation_in_invalid",
		"validation_not_in_invalid",
		"validation_length_too_short",
		"validation_length_too_long",
		"validation_length_invalid",
		"validation_min_greater_equal",
		"validation_min_invalid",
		"validation_max_less_equal",
		"validation_max_invalid",
		"validation_match_invalid",
		"validation_invalid_email",
		"validation_invalid_url",
		"validation_invalid_ip",
		"validation_invalid_ipv4",
		"validation_invalid_ipv6",
		"validation_date_invalid",
		"validation_date_too_early",
		"validation_date_too_late",
		"validation_invalid_alphanumeric",
		"validation_invalid_json",
		"validation_invalid_slug",
		"validation_not_slug",
		"validation_invalid_system_collection_name",
		"validation_unique",
		"validation_values_exist",
	}

	for _, code := range validationCodes {
		t.Run(code, func(t *testing.T) {
			// Create a validation error with field data
			fieldError := FieldError{
				Code:    code,
				Message: "Test validation error",
			}

			err := &Error{
				Status:  400,
				Message: "Validation failed",
				Data:    map[string]FieldError{"testField": fieldError},
			}

			// Test that it's recognized as a validation error
			if !err.IsValidation() {
				t.Error("Expected error to be recognized as validation error")
			}

			// Test that field errors are accessible
			fieldErrors := GetFieldErrors(err)
			if fieldErrors == nil {
				t.Fatal("Expected field errors to be accessible")
			}

			if fieldErrors["testField"].Code != code {
				t.Errorf("Expected field error code '%s', got '%s'", code, fieldErrors["testField"].Code)
			}
		})
	}
}

func TestAllKnownAliases(t *testing.T) {
	// Test that all known aliases are properly applied
	knownMessages := []struct {
		message string
		alias   string
	}{
		{"Failed to authenticate.", "authentication_failed"},
		{"You are not allowed to perform this action.", "action_forbidden"},
		{"The new email address is already in use.", "email_already_in_use"},
		{"The provided old password is not valid.", "invalid_old_password"},
		{"Invalid \"sort\" parameter format.", "invalid_sort_param"},
		{"Invalid \"expand\" parameter format.", "invalid_expand_param"},
		{"Invalid \"filter\" parameter format.", "invalid_filter_param"},
		{"Invalid \"fields\" parameter format.", "invalid_fields_param"},
		{"Invalid \"page\" parameter format.", "invalid_page_param"},
		{"Invalid \"perPage\" parameter format.", "invalid_perpage_param"},
		{"Invalid \"skipTotal\" parameter format.", "invalid_skiptotal_param"},
		{"Unsupported Content-Type", "unsupported_content_type"},
		{"Invalid request payload.", "invalid_request_payload"},
		{"Invalid request body.", "invalid_request_body"},
		{"Invalid or missing file.", "invalid_or_missing_file"},
		{"Invalid or missing record id.", "invalid_record_id"},
		{"Invalid or missing password reset token.", "invalid_password_reset_token"},
		{"Invalid or missing verification token.", "invalid_verification_token"},
		{"Invalid or missing file token.", "invalid_file_token"},
		{"Invalid or missing OAuth2 state parameter.", "invalid_oauth2_state"},
		{"Invalid or missing redirect URL.", "invalid_redirect_url"},
		{"Invalid redirect status code.", "invalid_redirect_status_code"},
		{"The requested resource wasn't found.", "resource_not_found"},
		{"File not found.", "file_not_found"},
		{"Collection not found.", "collection_not_found"},
		{"Record not found.", "record_not_found"},
		{"Missing or invalid collection context.", "collection_context_invalid"},
		{"Request entity too large", "entity_too_large"},
		{"Missing or invalid authentication token.", "invalid_auth_token"},
		{"Too Many Requests.", "too_many_requests"},
		{"Something went wrong while processing your request.", "internal_generic"},
	}

	for _, test := range knownMessages {
		t.Run(test.alias, func(t *testing.T) {
			err := &Error{
				Status:  400,
				Message: test.message,
				Data:    make(map[string]FieldError),
			}

			applyKnownAliases(err)

			if err.Code != test.alias {
				t.Errorf("Expected alias '%s' for message '%s', got '%s'", test.alias, test.message, err.Code)
			}
		})
	}
}

func TestIsAuthenticationFailed(t *testing.T) {
	// Test with 401 error
	authErr := &Error{Status: 401}
	if !IsAuthenticationFailed(authErr) {
		t.Error("Expected IsAuthenticationFailed to return true for 401 error")
	}

	// Test with 400 error with authentication_failed code
	authFailedErr := &Error{
		Status:  400,
		Code:    "authentication_failed",
		Message: "Failed to authenticate.",
	}
	if !IsAuthenticationFailed(authFailedErr) {
		t.Error("Expected IsAuthenticationFailed to return true for authentication_failed code")
	}

	// Test with other error
	otherErr := &Error{Status: 404}
	if IsAuthenticationFailed(otherErr) {
		t.Error("Expected IsAuthenticationFailed to return false for other error")
	}

	// Test with generic error
	genericErr := errors.New("generic error")
	if IsAuthenticationFailed(genericErr) {
		t.Error("Expected IsAuthenticationFailed to return false for generic error")
	}
}

func TestSentinelErrors(t *testing.T) {
	t.Run("record not found", func(t *testing.T) {
		err := ParseAPIError(http.StatusNotFound, []byte(`{"message":"Record not found.","data":{}}`))
		if err == nil {
			t.Fatal("expected error")
		}
		if !errors.Is(err, ErrRecordNotFound) {
			t.Fatalf("expected errors.Is(err, ErrRecordNotFound) to be true; got %v", err)
		}
	})

	t.Run("authentication failed", func(t *testing.T) {
		err := ParseAPIError(http.StatusBadRequest, []byte(`{"message":"Failed to authenticate.","data":{}}`))
		if err == nil {
			t.Fatal("expected error")
		}
		if !errors.Is(err, ErrAuthenticationFailed) {
			t.Fatalf("expected errors.Is(err, ErrAuthenticationFailed) to be true; got %v", err)
		}
	})
}
