package pocketbase

import (
	"context"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strconv"
	"strings"
	"testing"
)

func TestRecordService_GetAll(t *testing.T) {
	tests := []struct {
		name          string
		totalRecords  int
		expectedCalls int
		expectedItems int
		opts          *ListOptions
		expectError   bool
	}{
		{
			name:          "single page - less than 500 records",
			totalRecords:  100,
			expectedCalls: 1,
			expectedItems: 100,
			opts:          nil,
		},
		{
			name:          "single page - exactly 500 records",
			totalRecords:  500,
			expectedCalls: 2, // First page returns 500, second page returns 0
			expectedItems: 500,
			opts:          nil,
		},
		{
			name:          "multiple pages - 750 records",
			totalRecords:  750,
			expectedCalls: 2, // First page 500, second page 250
			expectedItems: 750,
			opts:          nil,
		},
		{
			name:          "multiple pages - 1500 records",
			totalRecords:  1500,
			expectedCalls: 4, // 500 + 500 + 500 + 0 (to check if there are more)
			expectedItems: 1500,
			opts:          nil,
		},
		{
			name:          "with filter options",
			totalRecords:  300,
			expectedCalls: 1,
			expectedItems: 300,
			opts: &ListOptions{
				Filter: "status = 'active'",
				Sort:   "-created",
			},
		},
		{
			name:          "empty collection",
			totalRecords:  0,
			expectedCalls: 1,
			expectedItems: 0,
			opts:          nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			callCount := 0
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				callCount++

				// Verify request parameters
				if r.Method != http.MethodGet {
					t.Errorf("Expected GET request, got %s", r.Method)
				}

				// Parse query parameters
				query := r.URL.Query()
				page, _ := strconv.Atoi(query.Get("page"))
				perPage, _ := strconv.Atoi(query.Get("perPage"))
				skipTotal := query.Get("skipTotal")

				// Verify pagination parameters
				if perPage != 500 {
					t.Errorf("Expected perPage=500, got %d", perPage)
				}
				if skipTotal != "1" {
					t.Errorf("Expected skipTotal=1, got %s", skipTotal)
				}

				// Verify filter and sort options are preserved
				if tt.opts != nil {
					if tt.opts.Filter != "" && query.Get("filter") != tt.opts.Filter {
						t.Errorf("Expected filter=%s, got %s", tt.opts.Filter, query.Get("filter"))
					}
					if tt.opts.Sort != "" && query.Get("sort") != tt.opts.Sort {
						t.Errorf("Expected sort=%s, got %s", tt.opts.Sort, query.Get("sort"))
					}
				}

				// Calculate items for this page
				startIndex := (page - 1) * 500
				endIndex := startIndex + 500
				if endIndex > tt.totalRecords {
					endIndex = tt.totalRecords
				}

				itemsThisPage := endIndex - startIndex
				if itemsThisPage < 0 {
					itemsThisPage = 0
				}

				// Generate response
				var items []string
				for i := 0; i < itemsThisPage; i++ {
					recordIndex := startIndex + i
					items = append(items, fmt.Sprintf(`{
						"id": "record_%d",
						"collectionId": "test_collection",
						"collectionName": "test",
						"name": "Record %d"
					}`, recordIndex, recordIndex))
				}

				response := fmt.Sprintf(`{
					"page": %d,
					"perPage": 500,
					"totalItems": %d,
					"totalPages": 1,
					"items": [%s]
				}`, page, itemsThisPage, strings.Join(items, ","))

				w.Header().Set("Content-Type", "application/json")
				w.WriteHeader(http.StatusOK)
				w.Write([]byte(response))
			}))
			defer server.Close()

			client := NewClient(server.URL)
			service := NewRecordService[Record](client, "test")

			// Execute GetAll
			result, err := service.GetAll(context.Background(), tt.opts)

			// Verify error expectation
			if tt.expectError {
				if err == nil {
					t.Error("Expected error, but got none")
				}
				return
			}

			if err != nil {
				t.Fatalf("Unexpected error: %v", err)
			}

			// Verify call count
			if callCount != tt.expectedCalls {
				t.Errorf("Expected %d API calls, got %d", tt.expectedCalls, callCount)
			}

			// Verify result structure
			if result == nil {
				t.Fatal("Result is nil")
			}

			// Verify pagination metadata
			if result.Page != 1 {
				t.Errorf("Expected Page=1, got %d", result.Page)
			}
			if result.TotalPages != 1 {
				t.Errorf("Expected TotalPages=1, got %d", result.TotalPages)
			}

			// For single page results (< 500 items), PerPage remains 500
			// For multi-page results, PerPage equals total items
			expectedPerPage := tt.expectedItems
			if tt.expectedItems < 500 {
				expectedPerPage = 500 // Single page case keeps original PerPage
			}
			if result.PerPage != expectedPerPage {
				t.Errorf("Expected PerPage=%d, got %d", expectedPerPage, result.PerPage)
			}

			if result.TotalItems != tt.expectedItems {
				t.Errorf("Expected TotalItems=%d, got %d", tt.expectedItems, result.TotalItems)
			}

			// Verify items count
			if len(result.Items) != tt.expectedItems {
				t.Errorf("Expected %d items, got %d", tt.expectedItems, len(result.Items))
			}

			// Verify items are in correct order
			for i, item := range result.Items {
				expectedID := fmt.Sprintf("record_%d", i)
				if item.ID != expectedID {
					t.Errorf("Expected item[%d].ID=%s, got %s", i, expectedID, item.ID)
				}
			}
		})
	}
}

func TestRecordService_GetAll_ErrorHandling(t *testing.T) {
	tests := []struct {
		name        string
		serverError int
		expectError bool
	}{
		{
			name:        "server error on first page",
			serverError: http.StatusInternalServerError,
			expectError: true,
		},
		{
			name:        "server error on second page",
			serverError: http.StatusBadRequest,
			expectError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			callCount := 0
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				callCount++

				// Return error on specific call
				if (tt.name == "server error on first page" && callCount == 1) ||
					(tt.name == "server error on second page" && callCount == 2) {
					w.WriteHeader(tt.serverError)
					w.Write([]byte(`{"message": "Server error"}`))
					return
				}

				// Return successful response for other calls
				response := `{
					"page": 1,
					"perPage": 500,
					"totalItems": 500,
					"totalPages": 1,
					"items": []
				}`
				for i := 0; i < 500; i++ {
					if i == 0 {
						response = strings.Replace(response, `"items": []`, fmt.Sprintf(`"items": [{"id": "record_%d", "collectionId": "test", "collectionName": "test"}`, i), 1)
					} else {
						response = strings.Replace(response, `]}`, fmt.Sprintf(`, {"id": "record_%d", "collectionId": "test", "collectionName": "test"}]}`, i), 1)
					}
				}

				w.Header().Set("Content-Type", "application/json")
				w.WriteHeader(http.StatusOK)
				w.Write([]byte(response))
			}))
			defer server.Close()

			client := NewClient(server.URL)
			service := NewRecordService[Record](client, "test")

			// Execute GetAll
			result, err := service.GetAll(context.Background(), nil)

			// Verify error expectation
			if tt.expectError {
				if err == nil {
					t.Error("Expected error, but got none")
				}
				if result != nil {
					t.Error("Expected nil result on error, but got result")
				}
			} else {
				if err != nil {
					t.Errorf("Unexpected error: %v", err)
				}
				if result == nil {
					t.Error("Expected result, but got nil")
				}
			}
		})
	}
}

func TestRecordService_GetAll_ContextCancellation(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Simulate slow response
		select {
		case <-r.Context().Done():
			return
		default:
			response := `{
				"page": 1,
				"perPage": 500,
				"totalItems": 1000,
				"totalPages": 2,
				"items": []
			}`
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusOK)
			w.Write([]byte(response))
		}
	}))
	defer server.Close()

	client := NewClient(server.URL)
	service := NewRecordService[Record](client, "test")

	// Create cancelled context
	ctx, cancel := context.WithCancel(context.Background())
	cancel() // Cancel immediately

	// Execute GetAll with cancelled context
	result, err := service.GetAll(ctx, nil)

	// Verify context cancellation is handled
	if err == nil {
		t.Error("Expected error due to context cancellation, but got none")
	}
	if result != nil {
		t.Error("Expected nil result on context cancellation, but got result")
	}
}

func TestRecordService_GetAll_OptionsPreservation(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		query := r.URL.Query()

		// Verify that original options are preserved
		expectedParams := map[string]string{
			"filter": "status = 'active'",
			"sort":   "-created,name",
			"expand": "author,category",
			"fields": "id,name,status",
		}

		for param, expected := range expectedParams {
			if query.Get(param) != expected {
				t.Errorf("Expected %s=%s, got %s", param, expected, query.Get(param))
			}
		}

		// Verify that pagination params are overridden
		if query.Get("page") == "" {
			t.Error("Expected page parameter to be set")
		}
		if query.Get("perPage") != "500" {
			t.Errorf("Expected perPage=500, got %s", query.Get("perPage"))
		}
		if query.Get("skipTotal") != "1" {
			t.Errorf("Expected skipTotal=1, got %s", query.Get("skipTotal"))
		}

		response := `{
			"page": 1,
			"perPage": 500,
			"totalItems": 10,
			"totalPages": 1,
			"items": [
				{"id": "1", "collectionId": "test", "collectionName": "test", "name": "Test 1"},
				{"id": "2", "collectionId": "test", "collectionName": "test", "name": "Test 2"}
			]
		}`

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(response))
	}))
	defer server.Close()

	client := NewClient(server.URL)
	service := NewRecordService[Record](client, "test")

	// Test with comprehensive options
	opts := &ListOptions{
		Filter:  "status = 'active'",
		Sort:    "-created,name",
		Expand:  "author,category",
		Fields:  "id,name,status",
		Page:    5,   // Should be overridden
		PerPage: 100, // Should be overridden
	}

	result, err := service.GetAll(context.Background(), opts)

	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}

	if result == nil {
		t.Fatal("Result is nil")
	}

	// Verify that the original options object is not modified
	if opts.Page != 5 {
		t.Errorf("Original options.Page was modified: expected 5, got %d", opts.Page)
	}
	if opts.PerPage != 100 {
		t.Errorf("Original options.PerPage was modified: expected 100, got %d", opts.PerPage)
	}
}

// Benchmark test for GetAll performance
func BenchmarkRecordService_GetAll(b *testing.B) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		response := `{
			"page": 1,
			"perPage": 500,
			"totalItems": 100,
			"totalPages": 1,
			"items": [`

		// Generate 100 test records
		for i := 0; i < 100; i++ {
			if i > 0 {
				response += ","
			}
			response += fmt.Sprintf(`{
				"id": "record_%d",
				"collectionId": "test_collection",
				"collectionName": "test",
				"name": "Record %d"
			}`, i, i)
		}

		response += `]}`

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(response))
	}))
	defer server.Close()

	client := NewClient(server.URL)
	service := NewRecordService[Record](client, "test")

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, err := service.GetAll(context.Background(), nil)
		if err != nil {
			b.Fatalf("Unexpected error: %v", err)
		}
	}
}
