package pocketbase

import (
	"context"
	"fmt"
	"net/http"
	"net/url"
)

// LogServiceAPI defines the API operations for viewing server logs.
type LogServiceAPI interface {
	GetRequestsList(ctx context.Context, opts *ListOptions) (*ListResult, error)
	GetRequest(ctx context.Context, requestID string) (map[string]interface{}, error)
	GetStats(ctx context.Context) (*LogStats, error)
}

// LogService provides functionality for viewing server request logs.
type LogService struct {
	Client *Client
}

var _ LogServiceAPI = (*LogService)(nil)

// GetRequestsList retrieves a list of request logs.
func (s *LogService) GetRequestsList(ctx context.Context, opts *ListOptions) (*ListResult, error) {
	path := buildPathWithQuery("/api/logs/requests", buildQueryString(opts))
	var res ListResult
	if err := s.Client.send(ctx, http.MethodGet, path, nil, &res); err != nil {
		return nil, wrapError("fetch", "logs list", err)
	}
	return &res, nil
}

// GetRequest retrieves detailed information for a specific request log.
func (s *LogService) GetRequest(ctx context.Context, requestID string) (map[string]interface{}, error) {
	path := fmt.Sprintf("/api/logs/requests/%s", url.PathEscape(requestID))
	var result map[string]interface{}
	if err := s.Client.send(ctx, http.MethodGet, path, nil, &result); err != nil {
		return nil, wrapError("fetch", "log request", err)
	}
	return result, nil
}

// LogStatItem represents the number of requests within a specific time period.
// LogStatItem represents the number of requests within a specific time period.
type LogStatItem struct {
	Time  string `json:"time"`
	Count int    `json:"count"`
}

// LogStats contains request statistics.
type LogStats struct {
	Total int           `json:"total"`
	Items []LogStatItem `json:"items"`
}

// GetStats retrieves request log statistics.
func (s *LogService) GetStats(ctx context.Context) (*LogStats, error) {
	var stats LogStats
	if err := s.Client.send(ctx, http.MethodGet, "/api/logs/stats", nil, &stats); err != nil {
		return nil, wrapError("fetch", "log stats", err)
	}
	return &stats, nil
}
