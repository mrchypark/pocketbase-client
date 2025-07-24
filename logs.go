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
	GetRequest(ctx context.Context, requestID string) (map[string]any, error)
	GetStats(ctx context.Context) (*LogStats, error)
}

// LogService provides functionality for viewing server request logs.
type LogService struct {
	Client *Client
}

var _ LogServiceAPI = (*LogService)(nil)

// GetRequestsList retrieves a list of request logs.
func (s *LogService) GetRequestsList(ctx context.Context, opts *ListOptions) (*ListResult, error) {
	path := "/api/logs/requests"
	q := url.Values{}
	ApplyListOptions(q, opts)
	if qs := q.Encode(); qs != "" {
		path += "?" + qs
	}
	var res ListResult
	if err := s.Client.send(ctx, http.MethodGet, path, nil, &res); err != nil {
		return nil, fmt.Errorf("pocketbase: fetch logs list: %w", err)
	}
	return &res, nil
}

// GetRequest retrieves detailed information for a specific request log.
func (s *LogService) GetRequest(ctx context.Context, requestID string) (map[string]any, error) {
	path := fmt.Sprintf("/api/logs/requests/%s", url.PathEscape(requestID))
	var result map[string]any
	if err := s.Client.send(ctx, http.MethodGet, path, nil, &result); err != nil {
		return nil, fmt.Errorf("pocketbase: fetch log request: %w", err)
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
	path := "/api/logs/stats"
	var stats LogStats
	if err := s.Client.send(ctx, http.MethodGet, path, nil, &stats); err != nil {
		return nil, fmt.Errorf("pocketbase: fetch log stats: %w", err)
	}
	return &stats, nil
}
