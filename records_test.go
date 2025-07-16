package pocketbase

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/goccy/go-json"
)

// ListResult definition for benchmarks (Lazy version)
type ListResultLazy struct {
	Page       int           `json:"page"`
	PerPage    int           `json:"perPage"`
	TotalItems int           `json:"totalItems"`
	TotalPages int           `json:"totalPages"`
	Items      []*RecordLazy `json:"items"`
}

// --- Benchmark helpers ---

// generateBenchListResponse generates large record list JSON for benchmark tests.
func generateBenchListResponse(numRecords int) []byte {
	items := make([]map[string]interface{}, numRecords)
	for i := 0; i < numRecords; i++ {
		items[i] = map[string]interface{}{
			"id":             fmt.Sprintf("rec_%d", i),
			"collectionId":   "posts_col",
			"collectionName": "posts",
			"created":        "2025-07-02T10:30:00.123Z",
			"updated":        "2025-07-02T10:30:00.456Z",
			"title":          fmt.Sprintf("Title %d", i),
			"is_published":   true,
			"view_count":     i * 10,
		}
	}
	response := map[string]interface{}{
		"page":       1,
		"perPage":    numRecords,
		"totalItems": numRecords,
		"totalPages": 1,
		"items":      items,
	}
	data, _ := json.Marshal(response)
	return data
}

// --- Integrated benchmark tests ---

var benchResponse = generateBenchListResponse(100) // Test with 100 records

// ✅ BenchmarkGetListLazy: Measure performance of existing approach (lazy parsing) for list retrieval
func BenchmarkGetListLazy(b *testing.B) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.Write(benchResponse)
	}))
	defer srv.Close()

	c := NewClient(srv.URL)
	b.ResetTimer()
	b.ReportAllocs()

	for i := 0; i < b.N; i++ {
		// Similar logic to GetList but modified to use ListResultLazy
		path := fmt.Sprintf("/api/collections/%s/records", "posts")
		var result ListResultLazy // Use lazy version
		if err := c.send(context.Background(), http.MethodGet, path, nil, &result); err != nil {
			b.Fatalf("unexpected error: %v", err)
		}
	}
}

// ✅ BenchmarkGetListEager: Measure performance of new approach (eager parsing) for list retrieval
func BenchmarkGetListEager(b *testing.B) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.Write(benchResponse)
	}))
	defer srv.Close()

	c := NewClient(srv.URL)
	b.ResetTimer()
	b.ReportAllocs()

	for i := 0; i < b.N; i++ {
		// Use actual Records.GetList method
		_, err := c.Records.GetList(context.Background(), "posts", &ListOptions{})
		if err != nil {
			b.Fatalf("unexpected error: %v", err)
		}
	}
}

// --- Existing unit tests (no modifications) ---
// (All existing Test... functions below are kept as-is)

// TestRecordServiceGetList tests the GetList method of RecordService.
func TestRecordServiceGetList(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/api/collections/posts/records" {
			t.Fatalf("unexpected path: %s", r.URL.Path)
		}
		if r.URL.Query().Get("page") != "2" {
			t.Fatalf("unexpected query: %s", r.URL.RawQuery)
		}
		_ = json.NewEncoder(w).Encode(ListResult{Items: []*Record{}})
	}))
	defer srv.Close()

	c := NewClient(srv.URL)
	res, err := c.Records.GetList(context.Background(), "posts", &ListOptions{Page: 2})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(res.Items) != 0 {
		t.Fatalf("unexpected items: %v", res.Items)
	}
}

// TestRecordServiceGetListFieldsSkipTotal tests the GetList method with fields and skipTotal options.
func TestRecordServiceGetListFieldsSkipTotal(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Query().Get("fields") != "id,title" {
			t.Fatalf("unexpected fields: %s", r.URL.Query().Get("fields"))
		}
		if r.URL.Query().Get("skipTotal") != "1" {
			t.Fatalf("unexpected skipTotal: %s", r.URL.Query().Get("skipTotal"))
		}
		_ = json.NewEncoder(w).Encode(ListResult{Items: []*Record{}})
	}))
	defer srv.Close()

	c := NewClient(srv.URL)
	opts := &ListOptions{Fields: "id,title", SkipTotal: true}
	if _, err := c.Records.GetList(context.Background(), "posts", opts); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

// TestRecordServiceGetOne tests the GetOne method of RecordService.
func TestRecordServiceGetOne(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/api/collections/posts/records/1" {
			t.Fatalf("unexpected path: %s", r.URL.Path)
		}
		_ = json.NewEncoder(w).Encode(Record{BaseModel: BaseModel{ID: "1"}})
	}))
	defer srv.Close()

	c := NewClient(srv.URL)
	rec, err := c.Records.GetOne(context.Background(), "posts", "1", nil)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if rec.ID != "1" {
		t.Fatalf("unexpected id: %s", rec.ID)
	}
}

// TestRecordServiceGetOneFields tests the GetOne method with fields option.
func TestRecordServiceGetOneFields(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Query().Get("fields") != "id,title" {
			t.Fatalf("unexpected fields: %s", r.URL.Query().Get("fields"))
		}
		_ = json.NewEncoder(w).Encode(Record{BaseModel: BaseModel{ID: "1"}})
	}))
	defer srv.Close()

	c := NewClient(srv.URL)
	opts := &GetOneOptions{Fields: "id,title"}
	if _, err := c.Records.GetOne(context.Background(), "posts", "1", opts); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

// TestRecordServiceCreate tests the Create method of RecordService.
func TestRecordServiceCreate(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			t.Fatalf("unexpected method: %s", r.Method)
		}
		if r.URL.Path != "/api/collections/posts/records" {
			t.Fatalf("unexpected path: %s", r.URL.Path)
		}
		_ = json.NewEncoder(w).Encode(Record{BaseModel: BaseModel{ID: "1"}})
	}))
	defer srv.Close()

	c := NewClient(srv.URL)
	rec, err := c.Records.Create(context.Background(), "posts", map[string]string{"title": "hi"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if rec.ID != "1" {
		t.Fatalf("unexpected id: %s", rec.ID)
	}
}

// TestRecordServiceCreateWithQuery tests the CreateWithOptions method with query parameters.
func TestRecordServiceCreateWithQuery(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Query().Get("expand") != "rel" {
			t.Fatalf("unexpected expand: %s", r.URL.Query().Get("expand"))
		}
		if r.URL.Query().Get("fields") != "id" {
			t.Fatalf("unexpected fields: %s", r.URL.Query().Get("fields"))
		}
		_ = json.NewEncoder(w).Encode(Record{BaseModel: BaseModel{ID: "1"}})
	}))
	defer srv.Close()

	c := NewClient(srv.URL)
	opts := &WriteOptions{Expand: "rel", Fields: "id"}
	if _, err := c.Records.CreateWithOptions(context.Background(), "posts", map[string]string{"title": "hi"}, opts); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

// TestRecordServiceUpdate tests the Update method of RecordService.
func TestRecordServiceUpdate(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPatch {
			t.Fatalf("unexpected method: %s", r.Method)
		}
		if r.URL.Path != "/api/collections/posts/records/1" {
			t.Fatalf("unexpected path: %s", r.URL.Path)
		}
		_ = json.NewEncoder(w).Encode(Record{BaseModel: BaseModel{ID: "1"}})
	}))
	defer srv.Close()

	c := NewClient(srv.URL)
	rec, err := c.Records.Update(context.Background(), "posts", "1", map[string]string{"title": "new"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if rec.ID != "1" {
		t.Fatalf("unexpected id: %s", rec.ID)
	}
}

func TestRecordServiceUpdateWithQuery(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Query().Get("expand") != "rel" {
			t.Fatalf("unexpected expand: %s", r.URL.Query().Get("expand"))
		}
		if r.URL.Query().Get("fields") != "id" {
			t.Fatalf("unexpected fields: %s", r.URL.Query().Get("fields"))
		}
		_ = json.NewEncoder(w).Encode(Record{BaseModel: BaseModel{ID: "1"}})
	}))
	defer srv.Close()

	c := NewClient(srv.URL)
	opts := &WriteOptions{Expand: "rel", Fields: "id"}
	if _, err := c.Records.UpdateWithOptions(context.Background(), "posts", "1", map[string]string{"title": "new"}, opts); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestRecordServiceDelete(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodDelete {
			t.Fatalf("unexpected method: %s", r.Method)
		}
		if r.URL.Path != "/api/collections/posts/records/1" {
			t.Fatalf("unexpected path: %s", r.URL.Path)
		}
		w.WriteHeader(http.StatusNoContent)
	}))
	defer srv.Close()

	c := NewClient(srv.URL)
	if err := c.Records.Delete(context.Background(), "posts", "1"); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestRecordServiceAuthHeader(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Verify headers
		if r.Header.Get("Authorization") != "tok" {
			t.Fatalf("missing auth header: got %q, want 'tok'", r.Header.Get("Authorization"))
		}
		w.WriteHeader(http.StatusOK) // Explicit success response
		_ = json.NewEncoder(w).Encode(ListResult{})
	}))
	defer srv.Close()

	c := NewClient(srv.URL)
	// ✨ Modified part: Set auth token with WithToken
	c.WithToken("tok")

	if _, err := c.Records.GetList(context.Background(), "posts", nil); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestRecordServiceNotFound(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusNotFound)
		_ = json.NewEncoder(w).Encode(APIError{Code: 404, Message: "no"})
	}))
	defer srv.Close()

	c := NewClient(srv.URL)
	_, err := c.Records.GetOne(context.Background(), "posts", "missing", nil)
	if err == nil {
		t.Fatal("expected error")
	}
	var apiErr *APIError
	if !errors.As(err, &apiErr) || apiErr.Code != 404 {
		t.Fatalf("unexpected error: %v", err)
	}
}

// --- Model definitions for testing ---
// These are structs that would actually be in the "models" package.

type TestUser struct {
	Name   string `json:"name"`
	Avatar string `json:"avatar"`
}
