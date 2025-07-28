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

// --- Test helpers for pagination ---

// generateTestRecords creates test records for pagination testing
func generateTestRecords(count int, prefix string) []map[string]interface{} {
	records := make([]map[string]interface{}, count)
	for i := 0; i < count; i++ {
		records[i] = map[string]interface{}{
			"id":             fmt.Sprintf("%s_%d", prefix, i),
			"collectionId":   "test_col",
			"collectionName": "test",
			"created":        "2025-07-02T10:30:00.123Z",
			"updated":        "2025-07-02T10:30:00.456Z",
			"title":          fmt.Sprintf("Test Record %d", i),
			"status":         "active",
		}
	}
	return records
}

// createPaginationMockServer creates a mock server that simulates pagination responses
func createPaginationMockServer(t *testing.T, totalRecords int, pageSize int) (*httptest.Server, *int) {
	requestCount := 0

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		requestCount++

		// Parse query parameters
		page := 1
		perPage := pageSize
		if pageStr := r.URL.Query().Get("page"); pageStr != "" {
			if n, err := fmt.Sscanf(pageStr, "%d", &page); err != nil || n != 1 {
				page = 1
			}
		}
		if perPageStr := r.URL.Query().Get("perPage"); perPageStr != "" {
			if n, err := fmt.Sscanf(perPageStr, "%d", &perPage); err != nil || n != 1 {
				perPage = pageSize
			}
		}

		// Calculate pagination
		totalPages := (totalRecords + perPage - 1) / perPage
		startIdx := (page - 1) * perPage
		endIdx := startIdx + perPage
		if endIdx > totalRecords {
			endIdx = totalRecords
		}

		// Generate records for this page
		var items []map[string]interface{}
		if startIdx < totalRecords {
			items = generateTestRecords(endIdx-startIdx, fmt.Sprintf("rec%d", startIdx))
		}

		response := map[string]interface{}{
			"page":       page,
			"perPage":    perPage,
			"totalItems": totalRecords,
			"totalPages": totalPages,
			"items":      items,
		}

		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(response)
	}))

	return srv, &requestCount
}

// createErrorMockServer creates a mock server that returns errors after a certain number of requests
func createErrorMockServer(t *testing.T, successfulRequests int, errorCode int, errorMessage string) (*httptest.Server, *int) {
	requestCount := 0

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		requestCount++

		if requestCount <= successfulRequests {
			// Return successful response
			response := map[string]interface{}{
				"page":       requestCount,
				"perPage":    100,
				"totalItems": 300, // Set to have 3 pages
				"totalPages": 3,
				"items":      generateTestRecords(1, fmt.Sprintf("rec%d", requestCount)),
			}
			w.Header().Set("Content-Type", "application/json")
			_ = json.NewEncoder(w).Encode(response)
		} else {
			// Return error
			w.WriteHeader(errorCode)
			_ = json.NewEncoder(w).Encode(APIError{
				Code:    errorCode,
				Message: errorMessage,
			})
		}
	}))

	return srv, &requestCount
}

// assertRecordsEqual compares two slices of records for equality
func assertRecordsEqual(t *testing.T, expected, actual []*Record) {
	t.Helper()

	if len(expected) != len(actual) {
		t.Errorf("record count mismatch: expected %d, got %d", len(expected), len(actual))
		return
	}

	for i, expectedRecord := range expected {
		actualRecord := actual[i]
		if expectedRecord.ID != actualRecord.ID {
			t.Errorf("record[%d] ID mismatch: expected %s, got %s", i, expectedRecord.ID, actualRecord.ID)
		}
		if expectedRecord.CollectionName != actualRecord.CollectionName {
			t.Errorf("record[%d] CollectionName mismatch: expected %s, got %s", i, expectedRecord.CollectionName, actualRecord.CollectionName)
		}
	}
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
		// Use actual Records method
		_, err := c.Records("posts").GetList(context.Background(), &ListOptions{})
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
	res, err := c.Records("posts").GetList(context.Background(), &ListOptions{Page: 2})
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
	if _, err := c.Records("posts").GetList(context.Background(), opts); err != nil {
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
	rec, err := c.Records("posts").GetOne(context.Background(), "1", nil)
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
	if _, err := c.Records("posts").GetOne(context.Background(), "1", opts); err != nil {
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
	record := &Record{}
	record.Set("title", "hi")
	rec, err := c.Records("posts").Create(context.Background(), record)
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
	record := &Record{}
	record.Set("title", "hi")
	if _, err := c.Records("posts").CreateWithOptions(context.Background(), record, opts); err != nil {
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
	record := &Record{}
	record.Set("title", "new")
	rec, err := c.Records("posts").Update(context.Background(), "1", record)
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
	record := &Record{}
	record.Set("title", "new")
	if _, err := c.Records("posts").UpdateWithOptions(context.Background(), "1", record, opts); err != nil {
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
	if err := c.Records("posts").Delete(context.Background(), "1"); err != nil {
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

	if _, err := c.Records("posts").GetList(context.Background(), nil); err != nil {
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
	_, err := c.Records("posts").GetOne(context.Background(), "missing", nil)
	if err == nil {
		t.Fatal("expected error")
	}
	var apiErr *APIError
	if !errors.As(err, &apiErr) || apiErr.Code != 404 {
		t.Fatalf("unexpected error: %v", err)
	}
}

// --- Backward Compatibility Tests ---

// TestBackwardCompatibility ensures new pagination helpers don't break existing functionality
func TestBackwardCompatibility(t *testing.T) {
	t.Run("Verify existing GetList method behavior", func(t *testing.T) {
		// Verify that existing GetList behavior has not changed
		srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			// Verify request path
			expectedPath := "/api/collections/posts/records"
			if r.URL.Path != expectedPath {
				t.Errorf("Expected path %s, got %s", expectedPath, r.URL.Path)
			}

			// Verify query parameters
			query := r.URL.Query()
			if query.Get("page") != "2" {
				t.Errorf("Expected page=2, got %s", query.Get("page"))
			}
			if query.Get("perPage") != "50" {
				t.Errorf("Expected perPage=50, got %s", query.Get("perPage"))
			}
			if query.Get("filter") != "status='active'" {
				t.Errorf("Expected filter=status='active', got %s", query.Get("filter"))
			}
			if query.Get("sort") != "-created" {
				t.Errorf("Expected sort=-created, got %s", query.Get("sort"))
			}
			if query.Get("expand") != "author" {
				t.Errorf("Expected expand=author, got %s", query.Get("expand"))
			}

			// Same response structure as before
			response := ListResult{
				Page:       2,
				PerPage:    50,
				TotalItems: 150,
				TotalPages: 3,
				Items: []*Record{
					{
						BaseModel: BaseModel{
							ID:             "test1",
							CollectionID:   "posts_col",
							CollectionName: "posts",
						},
					},
					{
						BaseModel: BaseModel{
							ID:             "test2",
							CollectionID:   "posts_col",
							CollectionName: "posts",
						},
					},
				},
			}

			w.Header().Set("Content-Type", "application/json")
			json.NewEncoder(w).Encode(response)
		}))
		defer srv.Close()

		client := NewClient(srv.URL)

		// Call GetList using the new approach
		result, err := client.Records("posts").GetList(context.Background(), &ListOptions{
			Page:    2,
			PerPage: 50,
			Filter:  "status='active'",
			Sort:    "-created",
			Expand:  "author",
		})

		if err != nil {
			t.Fatalf("GetList failed: %v", err)
		}

		// Verify response structure
		if result.Page != 2 {
			t.Errorf("Expected page 2, got %d", result.Page)
		}
		if result.PerPage != 50 {
			t.Errorf("Expected perPage 50, got %d", result.PerPage)
		}
		if result.TotalItems != 150 {
			t.Errorf("Expected totalItems 150, got %d", result.TotalItems)
		}
		if result.TotalPages != 3 {
			t.Errorf("Expected totalPages 3, got %d", result.TotalPages)
		}
		if len(result.Items) != 2 {
			t.Errorf("Expected 2 items, got %d", len(result.Items))
		}

		// Verify record structure
		if result.Items[0].ID != "test1" {
			t.Errorf("Expected first record ID 'test1', got '%s'", result.Items[0].ID)
		}
		if result.Items[0].CollectionName != "posts" {
			t.Errorf("Expected collection name 'posts', got '%s'", result.Items[0].CollectionName)
		}
	})
}

// min helper function for Go versions that don't have it built-in
func min(a, b, c int) int {
	if a < b {
		if a < c {
			return a
		}
		return c
	}
	if b < c {
		return b
	}
	return c
}

// --- Model definitions for testing ---
// These are structs that would actually be in the "models" package.

type TestUser struct {
	Name   string `json:"name"`
	Avatar string `json:"avatar"`
}
