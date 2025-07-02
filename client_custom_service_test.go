package pocketbase

import (
	"context"
	"testing"
)

type mockRecordService struct{}

func (m *mockRecordService) GetList(ctx context.Context, collection string, opts *ListOptions) (*ListResult, error) {
	return &ListResult{}, nil
}
func (m *mockRecordService) GetOne(ctx context.Context, collection, recordID string, opts *GetOneOptions) (*Record, error) {
	return &Record{}, nil
}
func (m *mockRecordService) Create(ctx context.Context, collection string, body interface{}) (*Record, error) {
	return &Record{}, nil
}
func (m *mockRecordService) CreateWithOptions(ctx context.Context, collection string, body interface{}, opts *WriteOptions) (*Record, error) {
	return &Record{}, nil
}
func (m *mockRecordService) Update(ctx context.Context, collection, recordID string, body interface{}) (*Record, error) {
	return &Record{}, nil
}
func (m *mockRecordService) UpdateWithOptions(ctx context.Context, collection, recordID string, body interface{}, opts *WriteOptions) (*Record, error) {
	return &Record{}, nil
}
func (m *mockRecordService) Delete(ctx context.Context, collection, recordID string) error {
	return nil
}
func (m *mockRecordService) NewCreateRequest(collection string, body map[string]any) (*BatchRequest, error) {
	return &BatchRequest{}, nil
}
func (m *mockRecordService) NewUpdateRequest(collection, recordID string, body map[string]any) (*BatchRequest, error) {
	return &BatchRequest{}, nil
}
func (m *mockRecordService) NewDeleteRequest(collection, recordID string) (*BatchRequest, error) {
	return &BatchRequest{}, nil
}
func (m *mockRecordService) NewUpsertRequest(collection string, body map[string]any) (*BatchRequest, error) {
	return &BatchRequest{}, nil
}

// TestClientAllowsCustomRecordService tests that a custom RecordService can be set on the client.
func TestClientAllowsCustomRecordService(t *testing.T) {
	c := NewClient("http://example.com")
	mock := &mockRecordService{}
	c.Records = mock
	if c.Records != mock {
		t.Fatalf("custom service not set")
	}
}
