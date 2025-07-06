package pocketbase

import (
	"context"
	"fmt"
	"net/http"
	"net/url"

	"github.com/goccy/go-json"
)

// RecordServiceAPI defines the API operations related to records.
// RecordServiceAPI defines the API operations related to records.
type RecordServiceAPI interface {
	GetList(ctx context.Context, collection string, opts *ListOptions) (*ListResult, error)
	GetOne(ctx context.Context, collection, recordID string, opts *GetOneOptions) (*Record, error)
	Create(ctx context.Context, collection string, body interface{}) (*Record, error)
	CreateWithOptions(ctx context.Context, collection string, body interface{}, opts *WriteOptions) (*Record, error)
	Update(ctx context.Context, collection, recordID string, body interface{}) (*Record, error)
	UpdateWithOptions(ctx context.Context, collection, recordID string, body interface{}, opts *WriteOptions) (*Record, error)
	Delete(ctx context.Context, collection, recordID string) error
	NewCreateRequest(collection string, body map[string]any) (*BatchRequest, error)
	NewUpdateRequest(collection, recordID string, body map[string]any) (*BatchRequest, error)
	NewDeleteRequest(collection, recordID string) (*BatchRequest, error)
	NewUpsertRequest(collection string, body map[string]any) (*BatchRequest, error)
}

// RecordService handles record-related API operations.
type RecordService struct {
	Client *Client
}

var _ RecordServiceAPI = (*RecordService)(nil)

// GetList retrieves a list of records from a collection.
func (s *RecordService) GetList(ctx context.Context, collection string, opts *ListOptions) (*ListResult, error) {
	path := fmt.Sprintf("/api/collections/%s/records", url.PathEscape(collection))
	q := url.Values{}
	applyListOptions(q, opts)
	if qs := q.Encode(); qs != "" {
		path += "?" + qs
	}
	var result ListResult
	if err := s.Client.send(ctx, http.MethodGet, path, nil, &result); err != nil {
		return nil, fmt.Errorf("pocketbase: fetch records list: %w", err)
	}
	return &result, nil
}

// GetOne retrieves a single record.
func (s *RecordService) GetOne(ctx context.Context, collection, recordID string, opts *GetOneOptions) (*Record, error) {
	path := fmt.Sprintf("/api/collections/%s/records/%s", url.PathEscape(collection), url.PathEscape(recordID))
	q := url.Values{}
	if opts != nil {
		if opts.Expand != "" {
			q.Set("expand", opts.Expand)
		}
		if opts.Fields != "" {
			q.Set("fields", opts.Fields)
		}
	}
	if qs := q.Encode(); qs != "" {
		path += "?" + qs
	}
	var rec Record
	if err := s.Client.send(ctx, http.MethodGet, path, nil, &rec); err != nil {
		return nil, fmt.Errorf("pocketbase: fetch record: %w", err)
	}
	return &rec, nil
}

// Create creates a new record.
func (s *RecordService) Create(ctx context.Context, collection string, body interface{}) (*Record, error) {
	return s.CreateWithOptions(ctx, collection, body, nil)
}

// CreateWithOptions creates a new record with additional query parameters.
func (s *RecordService) CreateWithOptions(ctx context.Context, collection string, body interface{}, opts *WriteOptions) (*Record, error) {
	path := fmt.Sprintf("/api/collections/%s/records", url.PathEscape(collection))
	q := url.Values{}
	if opts != nil {
		if opts.Expand != "" {
			q.Set("expand", opts.Expand)
		}
		if opts.Fields != "" {
			q.Set("fields", opts.Fields)
		}
	}
	if qs := q.Encode(); qs != "" {
		path += "?" + qs
	}
	var rec Record
	if err := s.Client.send(ctx, http.MethodPost, path, body, &rec); err != nil {
		return nil, fmt.Errorf("pocketbase: create record: %w", err)
	}
	return &rec, nil
}

// Update modifies an existing record.
func (s *RecordService) Update(ctx context.Context, collection, recordID string, body interface{}) (*Record, error) {
	return s.UpdateWithOptions(ctx, collection, recordID, body, nil)
}

// UpdateWithOptions modifies an existing record with additional query parameters.
func (s *RecordService) UpdateWithOptions(ctx context.Context, collection, recordID string, body interface{}, opts *WriteOptions) (*Record, error) {
	path := fmt.Sprintf("/api/collections/%s/records/%s", url.PathEscape(collection), url.PathEscape(recordID))
	q := url.Values{}
	if opts != nil {
		if opts.Expand != "" {
			q.Set("expand", opts.Expand)
		}
		if opts.Fields != "" {
			q.Set("fields", opts.Fields)
		}
	}
	if qs := q.Encode(); qs != "" {
		path += "?" + qs
	}
	var rec Record
	if err := s.Client.send(ctx, http.MethodPatch, path, body, &rec); err != nil {
		return nil, fmt.Errorf("pocketbase: update record: %w", err)
	}
	return &rec, nil
}

// Delete deletes a record.
func (s *RecordService) Delete(ctx context.Context, collection, recordID string) error {
	path := fmt.Sprintf("/api/collections/%s/records/%s", url.PathEscape(collection), url.PathEscape(recordID))
	if err := s.Client.send(ctx, http.MethodDelete, path, nil, nil); err != nil {
		return fmt.Errorf("pocketbase: delete record: %w", err)
	}
	return nil
}

// NewCreateRequest creates a create request for batch processing.
func (s *RecordService) NewCreateRequest(collection string, body map[string]any) (*BatchRequest, error) {
	return &BatchRequest{
		Method: http.MethodPost,
		URL:    fmt.Sprintf("/api/collections/%s/records", url.PathEscape(collection)),
		Body:   body,
	}, nil
}

// NewUpdateRequest creates an update request for batch processing.
func (s *RecordService) NewUpdateRequest(collection, recordID string, body map[string]any) (*BatchRequest, error) {
	return &BatchRequest{
		Method: http.MethodPatch,
		URL:    fmt.Sprintf("/api/collections/%s/records/%s", url.PathEscape(collection), url.PathEscape(recordID)),
		Body:   body,
	}, nil
}

// NewDeleteRequest creates a delete request for batch processing.
func (s *RecordService) NewDeleteRequest(collection, recordID string) (*BatchRequest, error) {
	return &BatchRequest{
		Method: http.MethodDelete,
		URL:    fmt.Sprintf("/api/collections/%s/records/%s", url.PathEscape(collection), url.PathEscape(recordID)),
	}, nil
}

// NewUpsertRequest creates an upsert request that updates a record if it exists, or creates it if it doesn't.
// It returns an error if the "id" field is missing from the body.
func (s *RecordService) NewUpsertRequest(collection string, body map[string]any) (*BatchRequest, error) {
	if _, ok := body["id"]; !ok {
		return nil, fmt.Errorf("upsert error: 'id' field is required in the body")
	}
	return &BatchRequest{
		Method: http.MethodPut,
		URL:    fmt.Sprintf("/api/collections/%s/records", url.PathEscape(collection)),
		Body:   body,
	}, nil
}

// GetListAs RecordService를 사용하여 레코드 목록을 가져오고, 지정된 타입의 슬라이스로 변환합니다.
func GetListAs[T any](ctx context.Context, s *RecordService, collection string, opts *ListOptions) ([]*T, error) {
	listResult, err := s.GetList(ctx, collection, opts)
	if err != nil {
		return nil, fmt.Errorf("pocketbase: get list failed: %w", err)
	}

	results := make([]*T, 0, len(listResult.Items))
	for _, record := range listResult.Items {
		goStruct, err := ToGoStruct[T](record)
		if err != nil {
			return nil, fmt.Errorf("pocketbase: failed to convert record ID '%s': %w", record.ID, err)
		}
		results = append(results, goStruct)
	}

	return results, nil
}

// GetOneAs RecordService를 사용하여 단일 레코드를 가져오고, 지정된 타입으로 변환합니다.
func GetOneAs[T any](ctx context.Context, s *RecordService, collection, recordID string, opts *GetOneOptions) (*T, error) {
	record, err := s.GetOne(ctx, collection, recordID, opts)
	if err != nil {
		return nil, fmt.Errorf("pocketbase: get one failed: %w", err)
	}

	return ToGoStruct[T](record)
}

// ToGoStruct Record의 Data 필드를 지정된 Go 구조체로 변환합니다.
func ToGoStruct[T any](record *Record) (*T, error) {
	// 1. record.Data (map[string]interface{})를 JSON 바이트로 마샬링합니다.
	jsonData, err := json.Marshal(record.Data)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal record data to JSON: %w", err)
	}

	// 2. 생성된 JSON 바이트를 목표 구조체(T)의 인스턴스로 언마샬링합니다.
	var result T
	if err := json.Unmarshal(jsonData, &result); err != nil {
		return nil, fmt.Errorf("failed to unmarshal JSON to target struct: %w", err)
	}

	return &result, nil
}
