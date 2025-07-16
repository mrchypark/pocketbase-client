package pocketbase

import (
	"errors"
	"fmt"
	"testing"
)

func TestAPIError_Error(t *testing.T) {
	err := &APIError{
		Code:    404,
		Message: "Not Found",
		Data:    nil,
	}
	expected := "pocketbase: API error (code=404): Not Found"
	if err.Error() != expected {
		t.Errorf("Expected error message '%s', got '%s'", expected, err.Error())
	}
}

func TestClientError_Error(t *testing.T) {
	apiErr := &APIError{
		Code:    400,
		Message: "Invalid data",
	}
	clientErr := &ClientError{
		BaseErr:     ErrBadRequest,
		OriginalErr: apiErr,
	}
	expected := fmt.Sprintf("%s: %s", ErrBadRequest, apiErr)
	if clientErr.Error() != expected {
		t.Errorf("Expected error message '%s', got '%s'", expected, clientErr.Error())
	}
}

func TestClientError_Unwrap(t *testing.T) {
	apiErr := &APIError{Code: 500, Message: "Server error"}
	clientErr := &ClientError{
		BaseErr:     ErrUnknown,
		OriginalErr: apiErr,
	}

	unwrappedErr := errors.Unwrap(clientErr)
	if unwrappedErr != apiErr {
		t.Errorf("Expected unwrapped error to be the original APIError, but it was not")
	}

	var target *APIError
	if !errors.As(clientErr, &target) {
		t.Errorf("errors.As should have found an *APIError in the chain")
	}
	if target != apiErr {
		t.Errorf("errors.As extracted the wrong error")
	}
}

func TestClientError_Is(t *testing.T) {
	clientErr := &ClientError{
		BaseErr:     ErrNotFound,
		OriginalErr: &APIError{Code: 404, Message: "Resource not found"},
	}

	if !errors.Is(clientErr, ErrNotFound) {
		t.Errorf("errors.Is should have returned true for the wrapped ErrNotFound")
	}

	if errors.Is(clientErr, ErrBadRequest) {
		t.Errorf("errors.Is should have returned false for a different error type")
	}
}
