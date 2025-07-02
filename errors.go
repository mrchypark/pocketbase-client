package pocketbase

import "fmt"

// APIError represents a structured error returned from the PocketBase API.
type APIError struct {
	Code    int                    `json:"code"`
	Message string                 `json:"message"`
	Data    map[string]interface{} `json:"data"`
}

// Error는 error 인터페이스를 구현합니다.
func (e *APIError) Error() string {
	return fmt.Sprintf("pocketbase: API error (code=%d): %s", e.Code, e.Message)
}
