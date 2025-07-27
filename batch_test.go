package pocketbase

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/goccy/go-json"
)

// TestBatchExecuteSuccess tests the successful execution of batch requests.
func TestBatchExecuteSuccess(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			t.Fatalf("unexpected method: %s", r.Method)
		}
		if r.URL.Path != "/api/batch" {
			t.Fatalf("unexpected path: %s", r.URL.Path)
		}
		body := struct {
			Requests []*BatchRequest `json:"requests"`
		}{}
		if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
			t.Fatalf("invalid body: %v", err)
		}
		if len(body.Requests) != 4 {
			t.Fatalf("unexpected request count: %d", len(body.Requests))
		}
		responses := []*BatchResponse{
			{Status: http.StatusCreated, Body: map[string]any{"id": "1"}},
			{Status: http.StatusOK, Body: map[string]any{"id": "2"}},
			{Status: http.StatusNoContent},
			{Status: http.StatusOK, Body: map[string]any{"id": "4"}},
		}
		_ = json.NewEncoder(w).Encode(responses)
	}))
	defer srv.Close()

	c := NewClient(srv.URL)
	createRecord := &Record{}
	createRecord.Set("title", "a")
	createReq, _ := c.Records("posts").NewCreateRequest(createRecord)

	updateRecord := &Record{}
	updateRecord.Set("title", "b")
	updateReq, _ := c.Records("posts").NewUpdateRequest("2", updateRecord)

	deleteReq, _ := c.Records("posts").NewDeleteRequest("3")

	upsertRecord := &Record{}
	upsertRecord.Set("id", "4")
	upsertRecord.Set("title", "c")
	upsertReq, _ := c.Records("posts").NewUpsertRequest(upsertRecord)

	res, err := c.Batch.Execute(context.Background(), []*BatchRequest{createReq, updateReq, deleteReq, upsertReq})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(res) != 4 {
		t.Fatalf("unexpected response count: %d", len(res))
	}
	if res[0].Status != http.StatusCreated || res[0].Body.(map[string]any)["id"] != "1" {
		t.Fatalf("unexpected create response: %+v", res[0])
	}
	if res[2].Status != http.StatusNoContent {
		t.Fatalf("unexpected delete response: %+v", res[2])
	}
}

// TestBatchExecutePartialFailure tests batch execution with a partial failure.
func TestBatchExecutePartialFailure(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		responses := []*BatchResponse{
			{
				Status: http.StatusBadRequest,
				Body:   json.RawMessage(`{"code":400,"message":"invalid","data":{"title":"required"}}`),
				ParsedError: &APIError{
					Code:    400,
					Message: "invalid",
					Data: map[string]any{
						"title": "required",
					},
				},
			},
			{Status: http.StatusNoContent},
		}
		_ = json.NewEncoder(w).Encode(responses)
	}))
	defer srv.Close()

	c := NewClient(srv.URL)
	createRecord := &Record{}
	createRecord.Set("title", "a")
	createReq, _ := c.Records("posts").NewCreateRequest(createRecord)
	deleteReq, _ := c.Records("posts").NewDeleteRequest("1")

	res, err := c.Batch.Execute(context.Background(), []*BatchRequest{createReq, deleteReq})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if res[0].Status != http.StatusBadRequest {
		t.Fatalf("unexpected status: %d", res[0].Status)
	}
	if res[0].ParsedError == nil || res[0].ParsedError.Code != 400 {
		t.Fatalf("expected parsed error: %+v", res[0])
	}
	if res[1].Status != http.StatusNoContent {
		t.Fatalf("unexpected status: %d", res[1].Status)
	}
}

// TestBatchExecuteFailure tests batch execution with a complete failure.
func TestBatchExecuteFailure(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusUnauthorized)
		_ = json.NewEncoder(w).Encode(APIError{Code: 401, Message: "unauthorized"})
	}))
	defer srv.Close()

	c := NewClient(srv.URL)
	_, err := c.Batch.Execute(context.Background(), []*BatchRequest{})
	if err == nil {
		t.Fatal("expected error")
	}
}

// TestBatchExecuteParsedError tests batch execution where a parsed error is returned.
func TestBatchExecuteParsedError(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		responses := []*BatchResponse{
			{
				Status: http.StatusBadRequest,
				Body:   json.RawMessage(`{"code":400,"message":"validation failed","data":{"title":"required"}}`),
				ParsedError: &APIError{
					Code:    400,
					Message: "validation failed",
					Data:    map[string]any{"title": "required"},
				},
			},
		}
		_ = json.NewEncoder(w).Encode(responses)
	}))
	defer srv.Close()

	c := NewClient(srv.URL)
	record := &Record{}
	record.Set("title", "")
	req, _ := c.Records("posts").NewCreateRequest(record)

	res, err := c.Batch.Execute(context.Background(), []*BatchRequest{req})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if res[0].ParsedError == nil {
		t.Fatalf("expected parsed error: %+v", res[0])
	}
	if res[0].ParsedError.Code != 400 {
		t.Fatalf("unexpected code: %d", res[0].ParsedError.Code)
	}
}

// TestBatchExecuteForbidden tests batch execution when forbidden.
func TestBatchExecuteForbidden(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusForbidden)
		_ = json.NewEncoder(w).Encode(APIError{Code: 403, Message: "Batch requests are not allowed."})
	}))
	defer srv.Close()

	c := NewClient(srv.URL)
	_, err := c.Batch.Execute(context.Background(), []*BatchRequest{})
	if err == nil {
		t.Fatal("expected error")
	}
	apiErr := &APIError{}
	if !errors.As(err, &apiErr) {
		t.Fatalf("expected APIError, got %T", err)
	}
	if apiErr.Code != 403 {
		t.Fatalf("expected status code 403, got %d", apiErr.Code)
	}
	if apiErr.Message != "Batch requests are not allowed." {
		t.Fatalf("expected message 'Batch requests are not allowed.', got %s", apiErr.Message)
	}
}

// TestNewUpsertRequestMissingID tests the NewUpsertRequest with a missing ID.
func TestNewUpsertRequestMissingID(t *testing.T) {
	c := NewClient("http://example.com")
	record := &Record{}
	record.Set("title", "a")
	if _, err := c.Records("posts").NewUpsertRequest(record); err == nil {
		t.Fatal("expected error for missing id")
	}
}
