package pocketbase

import "fmt"

// APIError represents a structured error returned from the PocketBase API.
type APIError struct {
	Code    int                    `json:"code"`
	Message string                 `json:"message"`
	Data    map[string]interface{} `json:"data"`
}

// Error implements the error interface.
func (e *APIError) Error() string {
	return fmt.Sprintf("pocketbase: API error (code=%d): %s", e.Code, e.Message)
}
