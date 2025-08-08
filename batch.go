package pocketbase

import (
	"context"
	"net/http"

	"github.com/goccy/go-json"
)

// BatchServiceAPI defines the functionality for executing batch operations.
// BatchServiceAPI defines the functionality for executing batch operations.
type BatchServiceAPI interface {
	Execute(ctx context.Context, requests []*BatchRequest) ([]*BatchResponse, error)
}

// BatchService interacts with the /api/batch endpoint.
type BatchService struct {
	client *Client
}

var _ BatchServiceAPI = (*BatchService)(nil)

// BatchRequest represents a single batch operation.
// BatchRequest represents a single batch operation.
type BatchRequest struct {
	Method string         `json:"method"`
	URL    string         `json:"url"`
	Body   map[string]any `json:"body,omitempty"`
}

// BatchResponse contains the result of an individual batch operation.
type BatchResponse struct {
	Status int `json:"status"`
	Body   any `json:"body"`

	// ParsedError contains structured error information from a failed response.
	ParsedError *Error `json:"-"`
}

// Execute sends the given requests to /api/batch.
func (s *BatchService) Execute(ctx context.Context, requests []*BatchRequest) ([]*BatchResponse, error) {
	type rawBatchResponse struct {
		Status int             `json:"status"`
		Body   json.RawMessage `json:"body"`
	}

	var rawResponses []*rawBatchResponse
	if err := s.client.send(ctx, http.MethodPost, "/api/batch", map[string]any{"requests": requests}, &rawResponses); err != nil {
		return nil, err
	}

	responses := make([]*BatchResponse, len(rawResponses))
	for i, rawRes := range rawResponses {
		res := &BatchResponse{Status: rawRes.Status}
		if rawRes.Status >= http.StatusBadRequest {
			if apiErr := ParseAPIError(rawRes.Status, rawRes.Body); apiErr != nil {
				res.ParsedError = apiErr.(*Error)
			}
			res.Body = rawRes.Body
		} else {
			if err := json.Unmarshal(rawRes.Body, &res.Body); err != nil {
				res.Body = string(rawRes.Body)
			}
		}
		responses[i] = res
	}

	return responses, nil
}
