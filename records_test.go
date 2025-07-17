package pocketbase

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

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
				"totalItems": 300, // 3페이지가 있다고 설정
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

// TestRecordServiceGetAll tests the GetAll method of RecordService.
func TestRecordServiceGetAll(t *testing.T) {
	t.Run("정상 케이스 - 여러 페이지", func(t *testing.T) {
		srv, requestCount := createPaginationMockServer(t, 250, 100)
		defer srv.Close()

		c := NewClient(srv.URL)
		records, err := c.Records.GetAll(context.Background(), "test", &ListOptions{
			Filter: "status = 'active'",
		})

		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		if len(records) != 250 {
			t.Errorf("expected 250 records, got %d", len(records))
		}

		if *requestCount != 3 { // 3 pages: 100, 100, 50
			t.Errorf("expected 3 requests, got %d", *requestCount)
		}

		// 첫 번째와 마지막 레코드 검증
		if records[0].ID != "rec0_0" {
			t.Errorf("expected first record ID to be 'rec0_0', got '%s'", records[0].ID)
		}
		if records[249].ID != "rec200_49" {
			t.Errorf("expected last record ID to be 'rec200_49', got '%s'", records[249].ID)
		}
	})

	t.Run("빈 결과", func(t *testing.T) {
		srv, requestCount := createPaginationMockServer(t, 0, 100)
		defer srv.Close()

		c := NewClient(srv.URL)
		records, err := c.Records.GetAll(context.Background(), "test", nil)

		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		if len(records) != 0 {
			t.Errorf("expected 0 records, got %d", len(records))
		}

		if *requestCount != 1 {
			t.Errorf("expected 1 request, got %d", *requestCount)
		}
	})

	t.Run("단일 페이지", func(t *testing.T) {
		srv, requestCount := createPaginationMockServer(t, 50, 100)
		defer srv.Close()

		c := NewClient(srv.URL)
		records, err := c.Records.GetAll(context.Background(), "test", nil)

		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		if len(records) != 50 {
			t.Errorf("expected 50 records, got %d", len(records))
		}

		if *requestCount != 1 {
			t.Errorf("expected 1 request, got %d", *requestCount)
		}
	})

	t.Run("컨텍스트 취소", func(t *testing.T) {
		requestCount := 0
		srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			requestCount++
			if requestCount == 1 {
				// 첫 번째 요청은 성공
				response := map[string]interface{}{
					"page":       1,
					"perPage":    100,
					"totalItems": 200,
					"totalPages": 2,
					"items":      generateTestRecords(100, "rec0"),
				}
				_ = json.NewEncoder(w).Encode(response)
			} else {
				// 두 번째 요청에서는 지연을 주어 컨텍스트 취소가 발생하도록 함
				time.Sleep(100 * time.Millisecond)
				response := map[string]interface{}{
					"page":       2,
					"perPage":    100,
					"totalItems": 200,
					"totalPages": 2,
					"items":      generateTestRecords(100, "rec100"),
				}
				_ = json.NewEncoder(w).Encode(response)
			}
		}))
		defer srv.Close()

		c := NewClient(srv.URL)
		ctx, cancel := context.WithCancel(context.Background())

		// 첫 번째 요청 후 잠시 기다렸다가 컨텍스트 취소
		go func() {
			time.Sleep(50 * time.Millisecond) // 첫 번째 요청이 완료될 시간을 줌
			cancel()
		}()

		records, err := c.Records.GetAll(ctx, "test", nil)

		// 부분 데이터와 함께 에러가 반환되어야 함
		if err == nil {
			t.Fatal("expected error due to context cancellation")
		}

		var paginationErr *PaginationError
		if !errors.As(err, &paginationErr) {
			t.Fatalf("expected PaginationError, got %T", err)
		}

		if paginationErr.Operation != "GetAll" {
			t.Errorf("expected operation 'GetAll', got '%s'", paginationErr.Operation)
		}

		// 부분 데이터가 있어야 함 (첫 번째 페이지)
		if len(records) == 0 {
			t.Error("expected partial data to be returned")
		}
	})

	t.Run("네트워크 에러", func(t *testing.T) {
		srv, requestCount := createErrorMockServer(t, 1, http.StatusInternalServerError, "Internal Server Error")
		defer srv.Close()

		c := NewClient(srv.URL)
		records, err := c.Records.GetAll(context.Background(), "test", nil)

		if err == nil {
			t.Fatal("expected error due to network failure")
		}

		var paginationErr *PaginationError
		if !errors.As(err, &paginationErr) {
			t.Fatalf("expected PaginationError, got %T", err)
		}

		if paginationErr.Operation != "GetAll" {
			t.Errorf("expected operation 'GetAll', got '%s'", paginationErr.Operation)
		}

		if paginationErr.Page != 2 {
			t.Errorf("expected error at page 2, got page %d", paginationErr.Page)
		}

		// 부분 데이터가 있어야 함 (첫 번째 페이지)
		if len(records) == 0 {
			t.Error("expected partial data to be returned")
		}

		if *requestCount < 2 {
			t.Errorf("expected at least 2 requests, got %d", *requestCount)
		}
	})

	t.Run("ListOptions 전달 확인", func(t *testing.T) {
		srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			// 쿼리 파라미터 검증
			if r.URL.Query().Get("filter") != "status='active'" {
				t.Errorf("expected filter='status='active'', got '%s'", r.URL.Query().Get("filter"))
			}
			if r.URL.Query().Get("sort") != "-created" {
				t.Errorf("expected sort='-created', got '%s'", r.URL.Query().Get("sort"))
			}
			if r.URL.Query().Get("expand") != "author" {
				t.Errorf("expected expand='author', got '%s'", r.URL.Query().Get("expand"))
			}

			response := map[string]interface{}{
				"page":       1,
				"perPage":    100,
				"totalItems": 1,
				"totalPages": 1,
				"items":      generateTestRecords(1, "rec0"),
			}
			_ = json.NewEncoder(w).Encode(response)
		}))
		defer srv.Close()

		c := NewClient(srv.URL)
		_, err := c.Records.GetAll(context.Background(), "test", &ListOptions{
			Filter: "status='active'",
			Sort:   "-created",
			Expand: "author",
		})

		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
	})
}

// TestRecordServiceGetAllWithBatchSize tests the GetAllWithBatchSize method.
func TestRecordServiceGetAllWithBatchSize(t *testing.T) {
	t.Run("정상 케이스 - 지정된 배치 크기", func(t *testing.T) {
		srv, requestCount := createPaginationMockServer(t, 150, 50)
		defer srv.Close()

		c := NewClient(srv.URL)
		records, err := c.Records.GetAllWithBatchSize(context.Background(), "test", nil, 50)

		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		if len(records) != 150 {
			t.Errorf("expected 150 records, got %d", len(records))
		}

		if *requestCount != 3 { // 3 pages: 50, 50, 50
			t.Errorf("expected 3 requests, got %d", *requestCount)
		}
	})

	t.Run("잘못된 배치 크기 - 0", func(t *testing.T) {
		srv, _ := createPaginationMockServer(t, 100, 100) // 기본 배치 크기 사용됨
		defer srv.Close()

		c := NewClient(srv.URL)
		records, err := c.Records.GetAllWithBatchSize(context.Background(), "test", nil, 0)

		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		if len(records) != 100 {
			t.Errorf("expected 100 records, got %d", len(records))
		}
	})

	t.Run("잘못된 배치 크기 - 음수", func(t *testing.T) {
		srv, _ := createPaginationMockServer(t, 100, 100) // 기본 배치 크기 사용됨
		defer srv.Close()

		c := NewClient(srv.URL)
		records, err := c.Records.GetAllWithBatchSize(context.Background(), "test", nil, -10)

		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		if len(records) != 100 {
			t.Errorf("expected 100 records, got %d", len(records))
		}
	})

	t.Run("잘못된 배치 크기 - 너무 큰 값", func(t *testing.T) {
		srv, _ := createPaginationMockServer(t, 100, 100) // 기본 배치 크기 사용됨
		defer srv.Close()

		c := NewClient(srv.URL)
		records, err := c.Records.GetAllWithBatchSize(context.Background(), "test", nil, 2000)

		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		if len(records) != 100 {
			t.Errorf("expected 100 records, got %d", len(records))
		}
	})

	t.Run("작은 배치 크기", func(t *testing.T) {
		srv, requestCount := createPaginationMockServer(t, 25, 5)
		defer srv.Close()

		c := NewClient(srv.URL)
		records, err := c.Records.GetAllWithBatchSize(context.Background(), "test", nil, 5)

		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		if len(records) != 25 {
			t.Errorf("expected 25 records, got %d", len(records))
		}

		if *requestCount != 5 { // 5 pages: 5, 5, 5, 5, 5
			t.Errorf("expected 5 requests, got %d", *requestCount)
		}
	})

	t.Run("마지막 페이지 부분 데이터", func(t *testing.T) {
		srv, requestCount := createPaginationMockServer(t, 127, 50)
		defer srv.Close()

		c := NewClient(srv.URL)
		records, err := c.Records.GetAllWithBatchSize(context.Background(), "test", nil, 50)

		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		if len(records) != 127 {
			t.Errorf("expected 127 records, got %d", len(records))
		}

		if *requestCount != 3 { // 3 pages: 50, 50, 27
			t.Errorf("expected 3 requests, got %d", *requestCount)
		}
	})
}

// TestRecordIterator_Next tests the Next method of RecordIterator.
func TestRecordIterator_Next(t *testing.T) {
	t.Run("정상 순회 - 여러 페이지", func(t *testing.T) {
		srv, requestCount := createPaginationMockServer(t, 250, 100)
		defer srv.Close()

		c := NewClient(srv.URL)
		iterator := c.Records.Iterate(context.Background(), "test", nil)

		recordCount := 0
		pageCount := 0
		batchRecordCount := 0
		for iterator.Next() {
			record := iterator.Record()
			if record == nil {
				t.Error("expected record, got nil")
			}
			recordCount++
			batchRecordCount++

			// 페이지 변경 감지를 위한 로직 (간단한 방법)
			if recordCount%100 == 1 {
				pageCount++
				t.Logf("Processing page %d, record count: %d, batch records: %d", pageCount, recordCount, batchRecordCount)
				batchRecordCount = 1 // 새 페이지 시작
			}
		}

		t.Logf("Last batch had %d records", batchRecordCount)

		if err := iterator.Error(); err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		t.Logf("Final record count: %d, request count: %d", recordCount, *requestCount)

		if recordCount != 250 {
			t.Errorf("expected 250 records, got %d", recordCount)
		}

		if *requestCount != 3 { // 3 pages
			t.Errorf("expected 3 requests, got %d", *requestCount)
		}
	})

	t.Run("빈 결과", func(t *testing.T) {
		srv, requestCount := createPaginationMockServer(t, 0, 100)
		defer srv.Close()

		c := NewClient(srv.URL)
		iterator := c.Records.Iterate(context.Background(), "test", nil)

		recordCount := 0
		for iterator.Next() {
			record := iterator.Record()
			if record == nil {
				t.Error("expected record, got nil")
			}
			recordCount++
		}

		if err := iterator.Error(); err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		if recordCount != 0 {
			t.Errorf("expected 0 records, got %d", recordCount)
		}

		if *requestCount != 1 {
			t.Errorf("expected 1 request, got %d", *requestCount)
		}
	})

	t.Run("단일 페이지", func(t *testing.T) {
		srv, requestCount := createPaginationMockServer(t, 50, 100)
		defer srv.Close()

		c := NewClient(srv.URL)
		iterator := c.Records.Iterate(context.Background(), "test", nil)

		recordCount := 0
		for iterator.Next() {
			record := iterator.Record()
			if record == nil {
				t.Error("expected record, got nil")
			}
			recordCount++
		}

		if err := iterator.Error(); err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		if recordCount != 50 {
			t.Errorf("expected 50 records, got %d", recordCount)
		}

		if *requestCount != 1 {
			t.Errorf("expected 1 request, got %d", *requestCount)
		}
	})

	t.Run("컨텍스트 취소", func(t *testing.T) {
		srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			response := map[string]interface{}{
				"page":       1,
				"perPage":    100,
				"totalItems": 200,
				"totalPages": 2,
				"items":      generateTestRecords(100, "rec0"),
			}
			_ = json.NewEncoder(w).Encode(response)
		}))
		defer srv.Close()

		c := NewClient(srv.URL)
		ctx, cancel := context.WithCancel(context.Background())
		iterator := c.Records.Iterate(ctx, "test", nil)

		recordCount := 0
		for iterator.Next() {
			record := iterator.Record()
			if record == nil {
				t.Error("expected record, got nil")
			}
			recordCount++
			if recordCount == 50 {
				cancel() // 중간에 컨텍스트 취소
			}
		}

		if err := iterator.Error(); err == nil {
			t.Fatal("expected error due to context cancellation")
		}

		// 부분적으로 처리된 레코드가 있어야 함
		if recordCount < 50 {
			t.Errorf("expected at least 50 records processed, got %d", recordCount)
		}
	})

	t.Run("네트워크 에러", func(t *testing.T) {
		srv, _ := createErrorMockServer(t, 1, http.StatusInternalServerError, "Internal Server Error")
		defer srv.Close()

		c := NewClient(srv.URL)
		iterator := c.Records.Iterate(context.Background(), "test", nil)

		recordCount := 0
		for iterator.Next() {
			record := iterator.Record()
			if record == nil {
				t.Error("expected record, got nil")
			}
			recordCount++
		}

		if err := iterator.Error(); err == nil {
			t.Fatal("expected error due to network failure")
		}

		// 첫 번째 페이지는 성공했어야 함
		if recordCount == 0 {
			t.Error("expected some records from first successful page")
		}
	})
}

// TestRecordIterator_Record tests the Record method of RecordIterator.
func TestRecordIterator_Record(t *testing.T) {
	t.Run("정상 레코드 반환", func(t *testing.T) {
		srv, _ := createPaginationMockServer(t, 5, 100)
		defer srv.Close()

		c := NewClient(srv.URL)
		iterator := c.Records.Iterate(context.Background(), "test", nil)

		var records []*Record
		for iterator.Next() {
			record := iterator.Record()
			if record == nil {
				t.Error("expected record, got nil")
				continue
			}
			records = append(records, record)
		}

		if err := iterator.Error(); err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		if len(records) != 5 {
			t.Errorf("expected 5 records, got %d", len(records))
		}

		// 레코드 순서 확인
		for i, record := range records {
			expectedID := fmt.Sprintf("rec0_%d", i)
			if record.ID != expectedID {
				t.Errorf("record[%d] ID mismatch: expected %s, got %s", i, expectedID, record.ID)
			}
		}
	})

	t.Run("Next() 호출 전 Record() 호출", func(t *testing.T) {
		srv, _ := createPaginationMockServer(t, 5, 100)
		defer srv.Close()

		c := NewClient(srv.URL)
		iterator := c.Records.Iterate(context.Background(), "test", nil)

		// Next() 호출 전에 Record() 호출
		record := iterator.Record()
		if record != nil {
			t.Error("expected nil record before calling Next()")
		}
	})

	t.Run("마지막 레코드 이후 Record() 호출", func(t *testing.T) {
		srv, _ := createPaginationMockServer(t, 1, 100)
		defer srv.Close()

		c := NewClient(srv.URL)
		iterator := c.Records.Iterate(context.Background(), "test", nil)

		// 모든 레코드 순회
		for iterator.Next() {
			iterator.Record()
		}

		// 마지막 레코드 이후 Record() 호출
		record := iterator.Record()
		if record != nil {
			t.Error("expected nil record after iteration completed")
		}
	})
}

// TestRecordIterator_Error tests the Error method of RecordIterator.
func TestRecordIterator_Error(t *testing.T) {
	t.Run("에러 없는 경우", func(t *testing.T) {
		srv, _ := createPaginationMockServer(t, 5, 100)
		defer srv.Close()

		c := NewClient(srv.URL)
		iterator := c.Records.Iterate(context.Background(), "test", nil)

		for iterator.Next() {
			iterator.Record()
		}

		if err := iterator.Error(); err != nil {
			t.Errorf("expected no error, got %v", err)
		}
	})

	t.Run("네트워크 에러", func(t *testing.T) {
		srv, _ := createErrorMockServer(t, 0, http.StatusInternalServerError, "Internal Server Error")
		defer srv.Close()

		c := NewClient(srv.URL)
		iterator := c.Records.Iterate(context.Background(), "test", nil)

		hasNext := iterator.Next()
		if hasNext {
			t.Error("expected Next() to return false due to error")
		}

		if err := iterator.Error(); err == nil {
			t.Fatal("expected error due to network failure")
		}
	})

	t.Run("컨텍스트 취소 에러", func(t *testing.T) {
		srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			// 응답을 지연시켜 컨텍스트 취소가 발생하도록 함
			response := map[string]interface{}{
				"page":       1,
				"perPage":    100,
				"totalItems": 100,
				"totalPages": 1,
				"items":      generateTestRecords(100, "rec0"),
			}
			_ = json.NewEncoder(w).Encode(response)
		}))
		defer srv.Close()

		c := NewClient(srv.URL)
		ctx, cancel := context.WithCancel(context.Background())
		iterator := c.Records.Iterate(ctx, "test", nil)

		cancel() // 즉시 취소

		hasNext := iterator.Next()
		if hasNext {
			t.Error("expected Next() to return false due to context cancellation")
		}

		err := iterator.Error()
		if err == nil {
			t.Fatal("expected error due to context cancellation")
		}

		if !errors.Is(err, context.Canceled) {
			t.Errorf("expected context.Canceled error, got %v", err)
		}
	})
}

// TestRecordIterator_WithBatchSize tests the IterateWithBatchSize method.
func TestRecordIterator_WithBatchSize(t *testing.T) {
	t.Run("지정된 배치 크기", func(t *testing.T) {
		srv, requestCount := createPaginationMockServer(t, 150, 25)
		defer srv.Close()

		c := NewClient(srv.URL)
		iterator := c.Records.IterateWithBatchSize(context.Background(), "test", nil, 25)

		recordCount := 0
		for iterator.Next() {
			record := iterator.Record()
			if record == nil {
				t.Error("expected record, got nil")
			}
			recordCount++
		}

		if err := iterator.Error(); err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		if recordCount != 150 {
			t.Errorf("expected 150 records, got %d", recordCount)
		}

		if *requestCount != 6 { // 6 pages: 25 each
			t.Errorf("expected 6 requests, got %d", *requestCount)
		}
	})

	t.Run("잘못된 배치 크기 - 기본값 사용", func(t *testing.T) {
		srv, _ := createPaginationMockServer(t, 100, 100) // 기본 배치 크기 사용됨
		defer srv.Close()

		c := NewClient(srv.URL)
		iterator := c.Records.IterateWithBatchSize(context.Background(), "test", nil, 0)

		recordCount := 0
		for iterator.Next() {
			record := iterator.Record()
			if record == nil {
				t.Error("expected record, got nil")
			}
			recordCount++
		}

		if err := iterator.Error(); err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		if recordCount != 100 {
			t.Errorf("expected 100 records, got %d", recordCount)
		}
	})
}

// TestPaginationErrorHandling tests comprehensive error handling scenarios.
func TestPaginationErrorHandling(t *testing.T) {
	t.Run("부분 데이터 보존 - GetAll", func(t *testing.T) {
		srv, _ := createErrorMockServer(t, 2, http.StatusInternalServerError, "Internal Server Error")
		defer srv.Close()

		c := NewClient(srv.URL)
		records, err := c.Records.GetAll(context.Background(), "test", nil)

		if err == nil {
			t.Fatal("expected error")
		}

		var paginationErr *PaginationError
		if !errors.As(err, &paginationErr) {
			t.Fatalf("expected PaginationError, got %T", err)
		}

		// 부분 데이터 확인
		partialData := paginationErr.GetPartialData()
		if len(partialData) != 2 { // 2개의 성공한 요청
			t.Errorf("expected 2 partial records, got %d", len(partialData))
		}

		// 반환된 records도 부분 데이터와 같아야 함
		if len(records) != len(partialData) {
			t.Errorf("returned records length mismatch: expected %d, got %d", len(partialData), len(records))
		}

		// PaginationError 필드 검증
		if paginationErr.Operation != "GetAll" {
			t.Errorf("expected operation 'GetAll', got '%s'", paginationErr.Operation)
		}

		if paginationErr.Page != 3 {
			t.Errorf("expected error at page 3, got page %d", paginationErr.Page)
		}

		if paginationErr.OriginalErr == nil {
			t.Error("expected original error to be set")
		}
	})

	t.Run("재시도 로직 - 일시적 에러", func(t *testing.T) {
		// 현재 구현에서는 재시도 로직이 없으므로 첫 번째 에러에서 바로 실패
		retryCount := 0
		srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			retryCount++

			// 항상 실패
			w.WriteHeader(http.StatusServiceUnavailable)
			_ = json.NewEncoder(w).Encode(APIError{
				Code:    503,
				Message: "Service Unavailable",
			})
		}))
		defer srv.Close()

		c := NewClient(srv.URL)
		_, err := c.Records.GetAll(context.Background(), "test", nil)

		if err == nil {
			t.Fatal("expected error due to service unavailable")
		}

		// 재시도 로직이 없으므로 1번만 요청
		if retryCount != 1 {
			t.Errorf("expected 1 request (no retries), got %d", retryCount)
		}
	})

	t.Run("비재시도 에러 - 인증 에러", func(t *testing.T) {
		requestCount := 0
		srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			requestCount++

			w.WriteHeader(http.StatusUnauthorized)
			_ = json.NewEncoder(w).Encode(APIError{
				Code:    401,
				Message: "Unauthorized",
			})
		}))
		defer srv.Close()

		c := NewClient(srv.URL)
		_, err := c.Records.GetAll(context.Background(), "test", nil)

		if err == nil {
			t.Fatal("expected error")
		}

		// 재시도하지 않았는지 확인 (1번만 요청)
		if requestCount != 1 {
			t.Errorf("expected 1 request (no retries), got %d", requestCount)
		}

		var apiErr *APIError
		if errors.As(err, &apiErr) && apiErr.Code == 401 {
			// 올바른 에러 타입
		} else {
			t.Errorf("expected APIError with code 401, got %v", err)
		}
	})

	t.Run("비재시도 에러 - 권한 에러", func(t *testing.T) {
		requestCount := 0
		srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			requestCount++

			w.WriteHeader(http.StatusForbidden)
			_ = json.NewEncoder(w).Encode(APIError{
				Code:    403,
				Message: "Forbidden",
			})
		}))
		defer srv.Close()

		c := NewClient(srv.URL)
		_, err := c.Records.GetAll(context.Background(), "test", nil)

		if err == nil {
			t.Fatal("expected error")
		}

		// 재시도하지 않았는지 확인
		if requestCount != 1 {
			t.Errorf("expected 1 request (no retries), got %d", requestCount)
		}
	})

	t.Run("비재시도 에러 - 잘못된 요청", func(t *testing.T) {
		requestCount := 0
		srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			requestCount++

			w.WriteHeader(http.StatusBadRequest)
			_ = json.NewEncoder(w).Encode(APIError{
				Code:    400,
				Message: "Bad Request",
			})
		}))
		defer srv.Close()

		c := NewClient(srv.URL)
		_, err := c.Records.GetAll(context.Background(), "test", nil)

		if err == nil {
			t.Fatal("expected error")
		}

		// 재시도하지 않았는지 확인
		if requestCount != 1 {
			t.Errorf("expected 1 request (no retries), got %d", requestCount)
		}
	})

	t.Run("비재시도 에러 - 리소스 없음", func(t *testing.T) {
		requestCount := 0
		srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			requestCount++

			w.WriteHeader(http.StatusNotFound)
			_ = json.NewEncoder(w).Encode(APIError{
				Code:    404,
				Message: "Not Found",
			})
		}))
		defer srv.Close()

		c := NewClient(srv.URL)
		_, err := c.Records.GetAll(context.Background(), "test", nil)

		if err == nil {
			t.Fatal("expected error")
		}

		// 재시도하지 않았는지 확인
		if requestCount != 1 {
			t.Errorf("expected 1 request (no retries), got %d", requestCount)
		}
	})

	t.Run("최대 재시도 횟수 초과", func(t *testing.T) {
		// 현재 구현에서는 재시도 로직이 없으므로 1번만 요청
		requestCount := 0
		srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			requestCount++

			// 항상 일시적 에러 반환
			w.WriteHeader(http.StatusServiceUnavailable)
			_ = json.NewEncoder(w).Encode(APIError{
				Code:    503,
				Message: "Service Unavailable",
			})
		}))
		defer srv.Close()

		c := NewClient(srv.URL)
		_, err := c.Records.GetAll(context.Background(), "test", nil)

		if err == nil {
			t.Fatal("expected error")
		}

		// 재시도 로직이 없으므로 1번만 요청
		if requestCount != 1 {
			t.Errorf("expected 1 request (no retries), got %d", requestCount)
		}
	})

	t.Run("Iterator 에러 처리", func(t *testing.T) {
		srv, _ := createErrorMockServer(t, 1, http.StatusInternalServerError, "Internal Server Error")
		defer srv.Close()

		c := NewClient(srv.URL)
		iterator := c.Records.Iterate(context.Background(), "test", nil)

		recordCount := 0
		for iterator.Next() {
			record := iterator.Record()
			if record == nil {
				t.Error("expected record, got nil")
			}
			recordCount++
		}

		if err := iterator.Error(); err == nil {
			t.Fatal("expected error from iterator")
		}

		// 첫 번째 페이지는 성공했어야 함
		if recordCount == 0 {
			t.Error("expected some records from first successful page")
		}
	})

	t.Run("PaginationError 메서드 테스트", func(t *testing.T) {
		originalErr := errors.New("network error")
		partialData := []*Record{
			{BaseModel: BaseModel{ID: "test1"}},
			{BaseModel: BaseModel{ID: "test2"}},
		}

		paginationErr := &PaginationError{
			Operation:   "GetAll",
			Page:        3,
			PartialData: partialData,
			OriginalErr: originalErr,
		}

		// Error() 메서드 테스트
		errorMsg := paginationErr.Error()
		if !strings.Contains(errorMsg, "GetAll") {
			t.Errorf("error message should contain operation: %s", errorMsg)
		}
		if !strings.Contains(errorMsg, "page 3") {
			t.Errorf("error message should contain page number: %s", errorMsg)
		}

		// Unwrap() 메서드 테스트
		if !errors.Is(paginationErr, originalErr) {
			t.Error("Unwrap() should return original error")
		}

		// GetPartialData() 메서드 테스트
		retrievedData := paginationErr.GetPartialData()
		if len(retrievedData) != 2 {
			t.Errorf("expected 2 partial records, got %d", len(retrievedData))
		}
		if retrievedData[0].ID != "test1" {
			t.Errorf("expected first record ID 'test1', got '%s'", retrievedData[0].ID)
		}
	})
}

// --- Model definitions for testing ---
// These are structs that would actually be in the "models" package.

type TestUser struct {
	Name   string `json:"name"`
	Avatar string `json:"avatar"`
}
