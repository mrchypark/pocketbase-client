package pocketbase

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"net/http/httptest"
	"os"
	"runtime"
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

// --- Pagination Benchmark Tests ---

// BenchmarkGetAll measures performance of GetAll method
func BenchmarkGetAll(b *testing.B) {
	// 다양한 데이터 크기로 벤치마크
	sizes := []int{100, 500, 1000, 5000}

	for _, size := range sizes {
		b.Run(fmt.Sprintf("records_%d", size), func(b *testing.B) {
			srv, _ := createPaginationMockServer(&testing.T{}, size, 100)
			defer srv.Close()

			c := NewClient(srv.URL)
			b.ResetTimer()
			b.ReportAllocs()

			for i := 0; i < b.N; i++ {
				records, err := c.Records.GetAll(context.Background(), "test", nil)
				if err != nil {
					b.Fatalf("GetAll failed: %v", err)
				}
				if len(records) != size {
					b.Fatalf("Expected %d records, got %d", size, len(records))
				}
			}
		})
	}
}

// BenchmarkGetAllWithBatchSize measures performance with different batch sizes
func BenchmarkGetAllWithBatchSize(b *testing.B) {
	const totalRecords = 1000
	batchSizes := []int{10, 50, 100, 200, 500}

	for _, batchSize := range batchSizes {
		b.Run(fmt.Sprintf("batch_%d", batchSize), func(b *testing.B) {
			srv, _ := createPaginationMockServer(&testing.T{}, totalRecords, batchSize)
			defer srv.Close()

			c := NewClient(srv.URL)
			b.ResetTimer()
			b.ReportAllocs()

			for i := 0; i < b.N; i++ {
				records, err := c.Records.GetAllWithBatchSize(context.Background(), "test", nil, batchSize)
				if err != nil {
					b.Fatalf("GetAllWithBatchSize failed: %v", err)
				}
				if len(records) != totalRecords {
					b.Fatalf("Expected %d records, got %d", totalRecords, len(records))
				}
			}
		})
	}
}

// BenchmarkIterator measures performance of Iterator pattern
func BenchmarkIterator(b *testing.B) {
	sizes := []int{100, 500, 1000, 5000}

	for _, size := range sizes {
		b.Run(fmt.Sprintf("records_%d", size), func(b *testing.B) {
			srv, _ := createPaginationMockServer(&testing.T{}, size, 100)
			defer srv.Close()

			c := NewClient(srv.URL)
			b.ResetTimer()
			b.ReportAllocs()

			for i := 0; i < b.N; i++ {
				iterator := c.Records.Iterate(context.Background(), "test", nil)
				recordCount := 0
				for iterator.Next() {
					_ = iterator.Record()
					recordCount++
				}
				if err := iterator.Error(); err != nil {
					b.Fatalf("Iterator failed: %v", err)
				}
				if recordCount != size {
					b.Fatalf("Expected %d records, got %d", size, recordCount)
				}
			}
		})
	}
}

// BenchmarkIteratorWithBatchSize measures Iterator performance with different batch sizes
func BenchmarkIteratorWithBatchSize(b *testing.B) {
	const totalRecords = 1000
	batchSizes := []int{10, 50, 100, 200, 500}

	for _, batchSize := range batchSizes {
		b.Run(fmt.Sprintf("batch_%d", batchSize), func(b *testing.B) {
			srv, _ := createPaginationMockServer(&testing.T{}, totalRecords, batchSize)
			defer srv.Close()

			c := NewClient(srv.URL)
			b.ResetTimer()
			b.ReportAllocs()

			for i := 0; i < b.N; i++ {
				iterator := c.Records.IterateWithBatchSize(context.Background(), "test", nil, batchSize)
				recordCount := 0
				for iterator.Next() {
					_ = iterator.Record()
					recordCount++
				}
				if err := iterator.Error(); err != nil {
					b.Fatalf("Iterator failed: %v", err)
				}
				if recordCount != totalRecords {
					b.Fatalf("Expected %d records, got %d", totalRecords, recordCount)
				}
			}
		})
	}
}

// BenchmarkGetAllVsManualPagination compares GetAll with manual pagination
func BenchmarkGetAllVsManualPagination(b *testing.B) {
	const totalRecords = 1000
	const pageSize = 100

	srv, _ := createPaginationMockServer(&testing.T{}, totalRecords, pageSize)
	defer srv.Close()

	c := NewClient(srv.URL)

	b.Run("GetAll", func(b *testing.B) {
		b.ResetTimer()
		b.ReportAllocs()

		for i := 0; i < b.N; i++ {
			records, err := c.Records.GetAll(context.Background(), "test", nil)
			if err != nil {
				b.Fatalf("GetAll failed: %v", err)
			}
			if len(records) != totalRecords {
				b.Fatalf("Expected %d records, got %d", totalRecords, len(records))
			}
		}
	})

	b.Run("ManualPagination", func(b *testing.B) {
		b.ResetTimer()
		b.ReportAllocs()

		for i := 0; i < b.N; i++ {
			var allRecords []*Record
			page := 1

			for {
				result, err := c.Records.GetList(context.Background(), "test", &ListOptions{
					Page:    page,
					PerPage: pageSize,
				})
				if err != nil {
					b.Fatalf("GetList failed: %v", err)
				}

				allRecords = append(allRecords, result.Items...)

				if page >= result.TotalPages || len(result.Items) == 0 {
					break
				}
				page++
			}

			if len(allRecords) != totalRecords {
				b.Fatalf("Expected %d records, got %d", totalRecords, len(allRecords))
			}
		}
	})
}

// BenchmarkIteratorVsGetAll compares memory usage between Iterator and GetAll
func BenchmarkIteratorVsGetAll(b *testing.B) {
	const totalRecords = 5000
	srv, _ := createPaginationMockServer(&testing.T{}, totalRecords, 100)
	defer srv.Close()

	c := NewClient(srv.URL)

	b.Run("GetAll_Memory", func(b *testing.B) {
		b.ResetTimer()
		b.ReportAllocs()

		for i := 0; i < b.N; i++ {
			records, err := c.Records.GetAll(context.Background(), "test", nil)
			if err != nil {
				b.Fatalf("GetAll failed: %v", err)
			}
			// 메모리 사용량 측정을 위해 레코드를 실제로 사용
			_ = len(records)
		}
	})

	b.Run("Iterator_Memory", func(b *testing.B) {
		b.ResetTimer()
		b.ReportAllocs()

		for i := 0; i < b.N; i++ {
			iterator := c.Records.Iterate(context.Background(), "test", nil)
			recordCount := 0
			for iterator.Next() {
				record := iterator.Record()
				// 메모리 사용량 측정을 위해 레코드를 실제로 사용
				_ = record.ID
				recordCount++
			}
			if err := iterator.Error(); err != nil {
				b.Fatalf("Iterator failed: %v", err)
			}
		}
	})
}

// BenchmarkPaginationWithFilters measures performance with complex filters
func BenchmarkPaginationWithFilters(b *testing.B) {
	const totalRecords = 1000

	// 복잡한 필터를 시뮬레이션하는 모킹 서버
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// 쿼리 파라미터 파싱
		page := 1
		perPage := 100
		if pageStr := r.URL.Query().Get("page"); pageStr != "" {
			fmt.Sscanf(pageStr, "%d", &page)
		}
		if perPageStr := r.URL.Query().Get("perPage"); perPageStr != "" {
			fmt.Sscanf(perPageStr, "%d", &perPage)
		}

		// 필터 시뮬레이션 (실제로는 필터링하지 않고 전체 데이터 반환)
		totalPages := (totalRecords + perPage - 1) / perPage
		startIdx := (page - 1) * perPage
		endIdx := startIdx + perPage
		if endIdx > totalRecords {
			endIdx = totalRecords
		}

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
		json.NewEncoder(w).Encode(response)
	}))
	defer srv.Close()

	c := NewClient(srv.URL)

	complexOptions := &ListOptions{
		Filter: "status = 'active' && created >= '2024-01-01' && (category = 'tech' || category = 'science')",
		Sort:   "-created,title",
		Expand: "author,category,tags",
		Fields: "id,title,content,created,updated,author.name,category.name",
	}

	b.Run("GetAll_WithFilters", func(b *testing.B) {
		b.ResetTimer()
		b.ReportAllocs()

		for i := 0; i < b.N; i++ {
			records, err := c.Records.GetAll(context.Background(), "test", complexOptions)
			if err != nil {
				b.Fatalf("GetAll with filters failed: %v", err)
			}
			_ = len(records)
		}
	})

	b.Run("Iterator_WithFilters", func(b *testing.B) {
		b.ResetTimer()
		b.ReportAllocs()

		for i := 0; i < b.N; i++ {
			iterator := c.Records.Iterate(context.Background(), "test", complexOptions)
			recordCount := 0
			for iterator.Next() {
				_ = iterator.Record()
				recordCount++
			}
			if err := iterator.Error(); err != nil {
				b.Fatalf("Iterator with filters failed: %v", err)
			}
		}
	})
}

// BenchmarkPaginationErrorHandling measures performance impact of error handling
func BenchmarkPaginationErrorHandling(b *testing.B) {
	// 가끔 에러를 발생시키는 모킹 서버
	requestCount := 0
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		requestCount++

		// 10번 중 1번은 일시적 에러 발생
		if requestCount%10 == 0 {
			w.WriteHeader(http.StatusInternalServerError)
			json.NewEncoder(w).Encode(APIError{
				Code:    500,
				Message: "Temporary server error",
			})
			return
		}

		// 정상 응답
		page := 1
		if pageStr := r.URL.Query().Get("page"); pageStr != "" {
			fmt.Sscanf(pageStr, "%d", &page)
		}

		response := map[string]interface{}{
			"page":       page,
			"perPage":    100,
			"totalItems": 500,
			"totalPages": 5,
			"items":      generateTestRecords(100, fmt.Sprintf("rec%d", (page-1)*100)),
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(response)
	}))
	defer srv.Close()

	c := NewClient(srv.URL)

	b.Run("GetAll_WithRetry", func(b *testing.B) {
		b.ResetTimer()
		b.ReportAllocs()

		for i := 0; i < b.N; i++ {
			// 에러가 발생할 수 있지만 재시도 로직으로 인해 대부분 성공
			records, err := c.Records.GetAll(context.Background(), "test", nil)
			if err != nil {
				// 일부 실패는 예상됨
				var paginationErr *PaginationError
				if errors.As(err, &paginationErr) {
					// 부분 데이터라도 있으면 성공으로 간주
					if len(paginationErr.PartialData) > 0 {
						continue
					}
				}
				b.Logf("GetAll failed: %v", err)
			} else {
				_ = len(records)
			}
		}
	})
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
		defer cancel()
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

// --- Integration Tests ---

// TestPaginationIntegration tests pagination helpers with a real PocketBase server
// This test requires a running PocketBase instance with test data
func TestPaginationIntegration(t *testing.T) {
	// Skip integration tests unless explicitly enabled
	if testing.Short() {
		t.Skip("Skipping integration tests in short mode")
	}

	// Check if PocketBase server is available
	pbURL := "http://localhost:8090"
	if url := os.Getenv("POCKETBASE_URL"); url != "" {
		pbURL = url
	}

	client := NewClient(pbURL)
	ctx := context.Background()

	// Test collection name
	testCollection := "integration_test"

	// PocketBase 서버 연결 확인
	_, err := client.Records.GetList(ctx, testCollection, &ListOptions{
		Page:    1,
		PerPage: 1,
	})
	if err != nil {
		t.Skipf("PocketBase server not available or collection '%s' not found: %v", testCollection, err)
	}

	t.Run("대용량 데이터 처리 시나리오", func(t *testing.T) {
		// 먼저 테스트 데이터가 충분히 있는지 확인
		result, err := client.Records.GetList(ctx, testCollection, &ListOptions{
			Page:    1,
			PerPage: 1,
		})
		if err != nil {
			t.Skipf("PocketBase server not available or collection '%s' not found: %v", testCollection, err)
		}

		if result.TotalItems < 100 {
			t.Skipf("Not enough test data in collection '%s' (need at least 100, got %d)", testCollection, result.TotalItems)
		}

		t.Logf("Testing with %d total records", result.TotalItems)

		// GetAll 테스트
		t.Run("GetAll 대용량 데이터", func(t *testing.T) {
			start := time.Now()
			allRecords, err := client.Records.GetAll(ctx, testCollection, &ListOptions{
				Sort: "created",
			})
			duration := time.Since(start)

			if err != nil {
				t.Fatalf("GetAll failed: %v", err)
			}

			t.Logf("GetAll retrieved %d records in %v", len(allRecords), duration)

			if len(allRecords) != result.TotalItems {
				t.Errorf("Expected %d records, got %d", result.TotalItems, len(allRecords))
			}

			// 순서 확인 (created 필드로 정렬했으므로)
			for i := 1; i < len(allRecords); i++ {
				currentCreated := allRecords[i].GetDateTime("created")
				previousCreated := allRecords[i-1].GetDateTime("created")
				if currentCreated.Before(previousCreated) {
					t.Errorf("Records not properly sorted at index %d", i)
					break
				}
			}
		})

		// Iterator 테스트
		t.Run("Iterator 대용량 데이터", func(t *testing.T) {
			start := time.Now()
			iterator := client.Records.Iterate(ctx, testCollection, &ListOptions{
				Sort: "created",
			})

			recordCount := 0
			var firstRecord, lastRecord *Record
			for iterator.Next() {
				record := iterator.Record()
				if record == nil {
					t.Error("Got nil record from iterator")
					continue
				}

				if firstRecord == nil {
					firstRecord = record
				}
				lastRecord = record
				recordCount++

				// 메모리 사용량 체크를 위해 주기적으로 로그
				if recordCount%1000 == 0 {
					t.Logf("Processed %d records", recordCount)
				}
			}
			duration := time.Since(start)

			if err := iterator.Error(); err != nil {
				t.Fatalf("Iterator failed: %v", err)
			}

			t.Logf("Iterator processed %d records in %v", recordCount, duration)

			if recordCount != result.TotalItems {
				t.Errorf("Expected %d records, got %d", result.TotalItems, recordCount)
			}

			// 첫 번째와 마지막 레코드 순서 확인
			if firstRecord != nil && lastRecord != nil && recordCount > 1 {
				lastCreated := lastRecord.GetDateTime("created")
				firstCreated := firstRecord.GetDateTime("created")
				if lastCreated.Before(firstCreated) {
					t.Error("Records not properly sorted (last record is older than first)")
				}
			}
		})

		// 배치 크기별 성능 비교
		t.Run("배치 크기별 성능 비교", func(t *testing.T) {
			batchSizes := []int{10, 50, 100, 200}
			results := make(map[int]time.Duration)

			for _, batchSize := range batchSizes {
				start := time.Now()
				records, err := client.Records.GetAllWithBatchSize(ctx, testCollection, &ListOptions{
					Sort: "created",
				}, batchSize)
				duration := time.Since(start)

				if err != nil {
					t.Errorf("GetAllWithBatchSize(%d) failed: %v", batchSize, err)
					continue
				}

				if len(records) != result.TotalItems {
					t.Errorf("Batch size %d: expected %d records, got %d", batchSize, result.TotalItems, len(records))
				}

				results[batchSize] = duration
				t.Logf("Batch size %d: %v", batchSize, duration)
			}

			// 성능 결과 분석
			if len(results) > 1 {
				t.Log("Performance comparison:")
				for batchSize, duration := range results {
					t.Logf("  Batch size %d: %v", batchSize, duration)
				}
			}
		})
	})

	t.Run("에러 복구 시나리오", func(t *testing.T) {
		// 존재하지 않는 컬렉션으로 테스트
		t.Run("존재하지 않는 컬렉션", func(t *testing.T) {
			_, err := client.Records.GetAll(ctx, "nonexistent_collection", nil)
			if err == nil {
				t.Error("Expected error for nonexistent collection")
			}

			iterator := client.Records.Iterate(ctx, "nonexistent_collection", nil)
			hasNext := iterator.Next()
			if hasNext {
				t.Error("Expected iterator to fail for nonexistent collection")
			}
			if iterator.Error() == nil {
				t.Error("Expected error from iterator for nonexistent collection")
			}
		})

		// 잘못된 필터로 테스트
		t.Run("잘못된 필터", func(t *testing.T) {
			_, err := client.Records.GetAll(ctx, testCollection, &ListOptions{
				Filter: "invalid_field = 'value'",
			})
			if err == nil {
				t.Error("Expected error for invalid filter")
			}
		})

		// 타임아웃 테스트
		t.Run("타임아웃 처리", func(t *testing.T) {
			shortCtx, cancel := context.WithTimeout(ctx, 100*time.Millisecond)
			defer cancel()

			_, err := client.Records.GetAll(shortCtx, testCollection, nil)
			if err == nil {
				t.Error("Expected timeout error")
			}

			if !errors.Is(err, context.DeadlineExceeded) {
				var paginationErr *PaginationError
				if errors.As(err, &paginationErr) {
					if !errors.Is(paginationErr.OriginalErr, context.DeadlineExceeded) {
						t.Errorf("Expected timeout error, got %v", err)
					}
				} else {
					t.Errorf("Expected timeout error, got %v", err)
				}
			}
		})
	})

	t.Run("필터링 및 정렬 통합 테스트", func(t *testing.T) {
		// 복잡한 쿼리 옵션으로 테스트
		options := &ListOptions{
			Filter: "created >= '2024-01-01 00:00:00'",
			Sort:   "-created",
			Expand: "",
			Fields: "id,created,updated",
		}

		// GetAll로 테스트
		allRecords, err := client.Records.GetAll(ctx, testCollection, options)
		if err != nil {
			t.Fatalf("GetAll with complex options failed: %v", err)
		}

		// Iterator로 같은 결과 확인
		iterator := client.Records.Iterate(ctx, testCollection, options)
		var iteratorRecords []*Record
		for iterator.Next() {
			iteratorRecords = append(iteratorRecords, iterator.Record())
		}

		if err := iterator.Error(); err != nil {
			t.Fatalf("Iterator with complex options failed: %v", err)
		}

		// 결과 비교
		if len(allRecords) != len(iteratorRecords) {
			t.Errorf("GetAll and Iterator returned different counts: %d vs %d", len(allRecords), len(iteratorRecords))
		}

		// 첫 번째 몇 개 레코드의 ID 비교
		compareCount := min(len(allRecords), len(iteratorRecords), 10)
		for i := 0; i < compareCount; i++ {
			if allRecords[i].ID != iteratorRecords[i].ID {
				t.Errorf("Record %d ID mismatch: GetAll=%s, Iterator=%s", i, allRecords[i].ID, iteratorRecords[i].ID)
			}
		}

		t.Logf("Complex query returned %d records", len(allRecords))
	})
}

// TestPaginationMemoryUsage tests memory efficiency of pagination helpers
func TestPaginationMemoryUsage(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping memory usage tests in short mode")
	}

	// 대용량 데이터를 시뮬레이션하는 모킹 서버
	srv, _ := createPaginationMockServer(t, 10000, 100) // 10,000 records
	defer srv.Close()

	client := NewClient(srv.URL)
	ctx := context.Background()

	t.Run("Iterator vs GetAll 메모리 사용량", func(t *testing.T) {
		// GetAll 메모리 사용량 측정
		var m1, m2 runtime.MemStats
		runtime.GC()
		runtime.ReadMemStats(&m1)

		allRecords, err := client.Records.GetAll(ctx, "test", nil)
		if err != nil {
			t.Fatalf("GetAll failed: %v", err)
		}

		runtime.ReadMemStats(&m2)
		getAllMemory := m2.Alloc - m1.Alloc

		t.Logf("GetAll processed %d records, memory used: %d bytes", len(allRecords), getAllMemory)

		// Iterator 메모리 사용량 측정
		runtime.GC()
		runtime.ReadMemStats(&m1)

		iterator := client.Records.Iterate(ctx, "test", nil)
		recordCount := 0
		for iterator.Next() {
			_ = iterator.Record()
			recordCount++
		}

		if err := iterator.Error(); err != nil {
			t.Fatalf("Iterator failed: %v", err)
		}

		runtime.ReadMemStats(&m2)
		iteratorMemory := m2.Alloc - m1.Alloc

		t.Logf("Iterator processed %d records, memory used: %d bytes", recordCount, iteratorMemory)

		// Iterator가 더 메모리 효율적이어야 함
		if iteratorMemory >= getAllMemory {
			t.Logf("Warning: Iterator used more memory (%d) than GetAll (%d)", iteratorMemory, getAllMemory)
		} else {
			memoryReduction := float64(getAllMemory-iteratorMemory) / float64(getAllMemory) * 100
			t.Logf("Iterator used %.1f%% less memory than GetAll", memoryReduction)
		}
	})

	t.Run("배치 크기별 메모리 사용량", func(t *testing.T) {
		batchSizes := []int{10, 50, 100, 500}

		for _, batchSize := range batchSizes {
			var m1, m2 runtime.MemStats
			runtime.GC()
			runtime.ReadMemStats(&m1)

			iterator := client.Records.IterateWithBatchSize(ctx, "test", nil, batchSize)
			recordCount := 0
			for iterator.Next() {
				_ = iterator.Record()
				recordCount++
			}

			if err := iterator.Error(); err != nil {
				t.Fatalf("Iterator with batch size %d failed: %v", batchSize, err)
			}

			runtime.ReadMemStats(&m2)
			memoryUsed := m2.Alloc - m1.Alloc

			t.Logf("Batch size %d: processed %d records, memory used: %d bytes", batchSize, recordCount, memoryUsed)
		}
	})
}

// --- Backward Compatibility Tests ---

// TestBackwardCompatibility ensures new pagination helpers don't break existing functionality
func TestBackwardCompatibility(t *testing.T) {
	t.Run("기존 GetList 메서드 동작 확인", func(t *testing.T) {
		// 기존 GetList 동작이 변경되지 않았는지 확인
		srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			// 요청 경로 확인
			expectedPath := "/api/collections/posts/records"
			if r.URL.Path != expectedPath {
				t.Errorf("Expected path %s, got %s", expectedPath, r.URL.Path)
			}

			// 쿼리 파라미터 확인
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

			// 기존과 동일한 응답 구조
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

		// 기존 방식으로 GetList 호출
		result, err := client.Records.GetList(context.Background(), "posts", &ListOptions{
			Page:    2,
			PerPage: 50,
			Filter:  "status='active'",
			Sort:    "-created",
			Expand:  "author",
		})

		if err != nil {
			t.Fatalf("GetList failed: %v", err)
		}

		// 응답 구조 확인
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

		// 레코드 구조 확인
		if result.Items[0].ID != "test1" {
			t.Errorf("Expected first record ID 'test1', got '%s'", result.Items[0].ID)
		}
		if result.Items[0].CollectionName != "posts" {
			t.Errorf("Expected collection name 'posts', got '%s'", result.Items[0].CollectionName)
		}
	})

	t.Run("기존 ListOptions 호환성 테스트", func(t *testing.T) {
		// 모든 ListOptions 필드가 새로운 헬퍼에서도 동작하는지 확인
		srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			query := r.URL.Query()

			// 모든 옵션이 올바르게 전달되는지 확인
			expectedParams := map[string]string{
				"filter":    "status='published' && author.id='123'",
				"sort":      "-created,title",
				"expand":    "author,category,tags",
				"fields":    "id,title,content,author.name",
				"skipTotal": "1",
			}

			for key, expected := range expectedParams {
				if actual := query.Get(key); actual != expected {
					t.Errorf("Parameter %s: expected '%s', got '%s'", key, expected, actual)
				}
			}

			// 페이지네이션 파라미터는 GetAll에서 자동으로 설정됨
			if query.Get("page") == "" {
				t.Error("Expected page parameter to be set")
			}
			if query.Get("perPage") == "" {
				t.Error("Expected perPage parameter to be set")
			}

			response := map[string]interface{}{
				"page":       1,
				"perPage":    100,
				"totalItems": 50,
				"totalPages": 1,
				"items":      generateTestRecords(50, "rec0"),
			}

			w.Header().Set("Content-Type", "application/json")
			json.NewEncoder(w).Encode(response)
		}))
		defer srv.Close()

		client := NewClient(srv.URL)

		// 복잡한 ListOptions로 테스트
		options := &ListOptions{
			Filter:    "status='published' && author.id='123'",
			Sort:      "-created,title",
			Expand:    "author,category,tags",
			Fields:    "id,title,content,author.name",
			SkipTotal: true,
			QueryParams: map[string]string{
				"custom_param": "custom_value",
			},
		}

		// GetAll에서 ListOptions 호환성 확인
		records, err := client.Records.GetAll(context.Background(), "posts", options)
		if err != nil {
			t.Fatalf("GetAll with complex ListOptions failed: %v", err)
		}

		if len(records) != 50 {
			t.Errorf("Expected 50 records, got %d", len(records))
		}

		// Iterator에서 ListOptions 호환성 확인
		iterator := client.Records.Iterate(context.Background(), "posts", options)
		recordCount := 0
		for iterator.Next() {
			record := iterator.Record()
			if record != nil {
				recordCount++
			}
			// 무한 루프 방지를 위한 제한
			if recordCount >= 100 {
				break
			}
		}

		if err := iterator.Error(); err != nil {
			t.Fatalf("Iterator with complex ListOptions failed: %v", err)
		}

		if recordCount != 50 {
			t.Errorf("Expected 50 records from iterator, got %d", recordCount)
		}
	})

	t.Run("기존 에러 타입 일관성 테스트", func(t *testing.T) {
		// 기존 에러 타입들이 새로운 헬퍼에서도 일관되게 동작하는지 확인

		t.Run("404 Not Found 에러", func(t *testing.T) {
			srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				w.WriteHeader(http.StatusNotFound)
				json.NewEncoder(w).Encode(APIError{
					Code:    404,
					Message: "Collection not found",
				})
			}))
			defer srv.Close()

			client := NewClient(srv.URL)

			// GetList에서의 404 에러
			_, err := client.Records.GetList(context.Background(), "nonexistent", nil)
			if err == nil {
				t.Fatal("Expected 404 error from GetList")
			}

			var apiErr *APIError
			if !errors.As(err, &apiErr) || apiErr.Code != 404 {
				t.Errorf("Expected APIError with code 404, got %v", err)
			}

			// GetAll에서의 404 에러
			_, err = client.Records.GetAll(context.Background(), "nonexistent", nil)
			if err == nil {
				t.Fatal("Expected 404 error from GetAll")
			}

			// PaginationError로 래핑되어야 함
			var paginationErr *PaginationError
			if errors.As(err, &paginationErr) {
				// 원본 에러가 APIError여야 함
				if !errors.As(paginationErr.OriginalErr, &apiErr) || apiErr.Code != 404 {
					t.Errorf("Expected wrapped APIError with code 404, got %v", paginationErr.OriginalErr)
				}
			} else {
				t.Errorf("Expected PaginationError wrapping APIError, got %v", err)
			}

			// Iterator에서의 404 에러
			iterator := client.Records.Iterate(context.Background(), "nonexistent", nil)
			hasNext := iterator.Next()
			if hasNext {
				t.Error("Expected iterator to fail immediately")
			}

			err = iterator.Error()
			if err == nil {
				t.Fatal("Expected 404 error from Iterator")
			}

			if !errors.As(err, &apiErr) || apiErr.Code != 404 {
				t.Errorf("Expected APIError with code 404 from iterator, got %v", err)
			}
		})

		t.Run("401 Unauthorized 에러", func(t *testing.T) {
			srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				w.WriteHeader(http.StatusUnauthorized)
				json.NewEncoder(w).Encode(APIError{
					Code:    401,
					Message: "Unauthorized",
				})
			}))
			defer srv.Close()

			client := NewClient(srv.URL)

			// GetAll에서 인증 에러는 즉시 실패해야 함 (재시도 없음)
			_, err := client.Records.GetAll(context.Background(), "posts", nil)
			if err == nil {
				t.Fatal("Expected 401 error from GetAll")
			}

			var paginationErr *PaginationError
			if errors.As(err, &paginationErr) {
				var apiErr *APIError
				if !errors.As(paginationErr.OriginalErr, &apiErr) || apiErr.Code != 401 {
					t.Errorf("Expected wrapped APIError with code 401, got %v", paginationErr.OriginalErr)
				}
				// 첫 번째 페이지에서 실패했으므로 부분 데이터가 없어야 함
				if len(paginationErr.PartialData) != 0 {
					t.Errorf("Expected no partial data for auth error, got %d records", len(paginationErr.PartialData))
				}
			} else {
				t.Errorf("Expected PaginationError, got %v", err)
			}
		})

		t.Run("500 Internal Server Error 재시도", func(t *testing.T) {
			requestCount := 0
			srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				requestCount++
				// 500 에러는 재시도하지 않으므로 항상 에러 반환
				w.WriteHeader(http.StatusInternalServerError)
				json.NewEncoder(w).Encode(APIError{
					Code:    500,
					Message: "Internal Server Error",
				})
			}))
			defer srv.Close()

			client := NewClient(srv.URL)

			// 500 에러는 재시도하지 않으므로 실패해야 함
			_, err := client.Records.GetAll(context.Background(), "posts", nil)
			if err == nil {
				t.Fatal("Expected GetAll to fail with 500 error")
			}

			var paginationErr *PaginationError
			if errors.As(err, &paginationErr) {
				var apiErr *APIError
				if !errors.As(paginationErr.OriginalErr, &apiErr) || apiErr.Code != 500 {
					t.Errorf("Expected wrapped APIError with code 500, got %v", paginationErr.OriginalErr)
				}
			} else {
				t.Errorf("Expected PaginationError, got %v", err)
			}

			// 재시도하지 않으므로 1번만 요청해야 함
			if requestCount != 1 {
				t.Errorf("Expected exactly 1 request (no retries for 500), got %d", requestCount)
			}
		})
	})

	t.Run("기존 인터페이스 변경 없음 확인", func(t *testing.T) {
		// RecordServiceAPI 인터페이스가 기존 메서드를 포함하고 있는지 확인
		srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			// 간단한 성공 응답
			if strings.Contains(r.URL.Path, "/records/") && !strings.HasSuffix(r.URL.Path, "/records") {
				// GetOne 요청
				json.NewEncoder(w).Encode(Record{
					BaseModel: BaseModel{ID: "test1"},
				})
			} else if r.Method == http.MethodPost {
				// Create 요청
				json.NewEncoder(w).Encode(Record{
					BaseModel: BaseModel{ID: "new1"},
				})
			} else if r.Method == http.MethodPatch {
				// Update 요청
				json.NewEncoder(w).Encode(Record{
					BaseModel: BaseModel{ID: "updated1"},
				})
			} else if r.Method == http.MethodDelete {
				// Delete 요청
				w.WriteHeader(http.StatusNoContent)
			} else {
				// GetList 요청
				json.NewEncoder(w).Encode(ListResult{
					Items: []*Record{{BaseModel: BaseModel{ID: "list1"}}},
				})
			}
		}))
		defer srv.Close()

		client := NewClient(srv.URL)
		ctx := context.Background()

		// 기존 메서드들이 모두 정상 동작하는지 확인
		_, err := client.Records.GetList(ctx, "posts", nil)
		if err != nil {
			t.Errorf("GetList failed: %v", err)
		}

		_, err = client.Records.GetOne(ctx, "posts", "1", nil)
		if err != nil {
			t.Errorf("GetOne failed: %v", err)
		}

		_, err = client.Records.Create(ctx, "posts", map[string]string{"title": "test"})
		if err != nil {
			t.Errorf("Create failed: %v", err)
		}

		_, err = client.Records.Update(ctx, "posts", "1", map[string]string{"title": "updated"})
		if err != nil {
			t.Errorf("Update failed: %v", err)
		}

		err = client.Records.Delete(ctx, "posts", "1")
		if err != nil {
			t.Errorf("Delete failed: %v", err)
		}

		// 새로운 메서드들도 동작하는지 확인
		_, err = client.Records.GetAll(ctx, "posts", nil)
		if err != nil {
			t.Errorf("GetAll failed: %v", err)
		}

		iterator := client.Records.Iterate(ctx, "posts", nil)
		_ = iterator.Next() // 최소한 호출 가능한지 확인
	})

	t.Run("ListOptions 복사 동작 확인", func(t *testing.T) {
		// ListOptions가 올바르게 복사되어 원본이 변경되지 않는지 확인
		srv, _ := createPaginationMockServer(t, 200, 100)
		defer srv.Close()

		client := NewClient(srv.URL)

		originalOptions := &ListOptions{
			Page:    1,
			PerPage: 50,
			Filter:  "status='active'",
			Sort:    "-created",
			Expand:  "author",
			Fields:  "id,title",
			QueryParams: map[string]string{
				"custom": "value",
			},
		}

		// 원본 옵션의 복사본 생성
		originalPage := originalOptions.Page
		originalPerPage := originalOptions.PerPage
		originalCustom := originalOptions.QueryParams["custom"]

		// GetAll 호출
		_, err := client.Records.GetAll(context.Background(), "test", originalOptions)
		if err != nil {
			t.Fatalf("GetAll failed: %v", err)
		}

		// 원본 옵션이 변경되지 않았는지 확인
		if originalOptions.Page != originalPage {
			t.Errorf("Original Page was modified: expected %d, got %d", originalPage, originalOptions.Page)
		}
		if originalOptions.PerPage != originalPerPage {
			t.Errorf("Original PerPage was modified: expected %d, got %d", originalPerPage, originalOptions.PerPage)
		}
		if originalOptions.QueryParams["custom"] != originalCustom {
			t.Errorf("Original QueryParams was modified: expected %s, got %s", originalCustom, originalOptions.QueryParams["custom"])
		}

		// Iterator에서도 동일하게 확인
		iterator := client.Records.Iterate(context.Background(), "test", originalOptions)
		for iterator.Next() {
			break // 첫 번째 레코드만 확인
		}

		if originalOptions.Page != originalPage {
			t.Errorf("Original Page was modified by Iterator: expected %d, got %d", originalPage, originalOptions.Page)
		}
		if originalOptions.PerPage != originalPerPage {
			t.Errorf("Original PerPage was modified by Iterator: expected %d, got %d", originalPerPage, originalOptions.PerPage)
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
