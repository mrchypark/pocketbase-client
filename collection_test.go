package pocketbase

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestCollectionService_GetList(t *testing.T) {
	// Mock server
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/api/collections" {
			t.Fatalf("unexpected path: %s", r.URL.Path)
		}
		if r.Method != http.MethodGet {
			t.Fatalf("unexpected method: %s", r.Method)
		}

		resp := CollectionListResult{
			Page:       1,
			PerPage:    10,
			TotalItems: 1,
			TotalPages: 1,
			Items: []*Collection{
				{
					Name: "test_collection",
					Type: "base",
				},
			},
		}
		json.NewEncoder(w).Encode(resp)
	}))
	defer srv.Close()

	// Client and service setup
	c := NewClient(srv.URL)
	s := &CollectionService{Client: c}

	// Call the method under test
	res, err := s.GetList(context.Background(), nil)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	// Assertions
	if res.TotalItems != 1 {
		t.Fatalf("expected 1 total item, got %d", res.TotalItems)
	}
	if len(res.Items) != 1 {
				t.Fatalf("expected 1 item, got %d", len(res.Items))
	}
	if res.Items[0].Name != "test_collection" {
		t.Fatalf("expected collection name 'test_collection', got %s", res.Items[0].Name)
	}
}

func TestCollectionService_GetOne(t *testing.T) {
	// Mock server
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/api/collections/test_id" {
			t.Fatalf("unexpected path: %s", r.URL.Path)
		}
		if r.Method != http.MethodGet {
			t.Fatalf("unexpected method: %s", r.Method)
		}

		resp := Collection{
			Name: "test_collection_one",
			Type: "base",
		}
		json.NewEncoder(w).Encode(resp)
	}))
	defer srv.Close()

	// Client and service setup
	c := NewClient(srv.URL)
	s := &CollectionService{Client: c}

	// Call the method under test
	res, err := s.GetOne(context.Background(), "test_id")
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	// Assertions
	if res.Name != "test_collection_one" {
		t.Fatalf("expected collection name 'test_collection_one', got %s", res.Name)
	}
}

func TestCollectionService_Create(t *testing.T) {
	// Mock server
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/api/collections" {
			t.Fatalf("unexpected path: %s", r.URL.Path)
		}
		if r.Method != http.MethodPost {
			t.Fatalf("unexpected method: %s", r.Method)
		}

		var reqCol Collection
		json.NewDecoder(r.Body).Decode(&reqCol)

		resp := reqCol
		resp.Name = "created_" + reqCol.Name // Simulate server-side modification
		json.NewEncoder(w).Encode(resp)
	}))
	defer srv.Close()

	// Client and service setup
	c := NewClient(srv.URL)
	s := &CollectionService{Client: c}

	// Call the method under test
	newCol := &Collection{Name: "new_collection", Type: "base"}
	res, err := s.Create(context.Background(), newCol)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	// Assertions
	if res.Name != "created_new_collection" {
		t.Fatalf("expected created collection name 'created_new_collection', got %s", res.Name)
	}
}

func TestCollectionService_Update(t *testing.T) {
	// Mock server
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/api/collections/test_id" {
			t.Fatalf("unexpected path: %s", r.URL.Path)
		}
		if r.Method != http.MethodPatch {
			t.Fatalf("unexpected method: %s", r.Method)
		}

		var reqCol Collection
		json.NewDecoder(r.Body).Decode(&reqCol)

		resp := reqCol
		resp.Name = "updated_" + reqCol.Name // Simulate server-side modification
		json.NewEncoder(w).Encode(resp)
	}))
	defer srv.Close()

	// Client and service setup
	c := NewClient(srv.URL)
	s := &CollectionService{Client: c}

	// Call the method under test
	updatedCol := &Collection{Name: "existing_collection", Type: "base"}
	res, err := s.Update(context.Background(), "test_id", updatedCol)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	// Assertions
	if res.Name != "updated_existing_collection" {
		t.Fatalf("expected updated collection name 'updated_existing_collection', got %s", res.Name)
	}
}

func TestCollectionService_Delete(t *testing.T) {
	// Mock server
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/api/collections/test_id" {
			t.Fatalf("unexpected path: %s", r.URL.Path)
		}
		if r.Method != http.MethodDelete {
			t.Fatalf("unexpected method: %s", r.Method)
		}
		w.WriteHeader(http.StatusNoContent)
	}))
	defer srv.Close()

	// Client and service setup
	c := NewClient(srv.URL)
	s := &CollectionService{Client: c}

	// Call the method under test
	err := s.Delete(context.Background(), "test_id")
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
}

func TestCollectionService_Import(t *testing.T) {
	// Mock server
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/api/collections/import" && r.URL.Path != "/api/collections/import?deleteMissing=1" {
			t.Fatalf("unexpected path: %s", r.URL.Path)
		}
		if r.Method != http.MethodPut {
			t.Fatalf("unexpected method: %s", r.Method)
		}

		var reqCols []*Collection
		json.NewDecoder(r.Body).Decode(&reqCols)

		resp := make([]*Collection, len(reqCols))
		for i, col := range reqCols {
			resp[i] = &Collection{Name: "imported_" + col.Name, Type: col.Type}
		}
		json.NewEncoder(w).Encode(resp)
	}))
	defer srv.Close()

	// Client and service setup
	c := NewClient(srv.URL)
	s := &CollectionService{Client: c}

	// Call the method under test
	colsToImport := []*Collection{
		{Name: "col1", Type: "base"},
		{Name: "col2", Type: "auth"},
	}
	res, err := s.Import(context.Background(), colsToImport, true)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	// Assertions
	if len(res) != 2 {
		t.Fatalf("expected 2 imported collections, got %d", len(res))
	}
	if res[0].Name != "imported_col1" || res[1].Name != "imported_col2" {
		t.Fatalf("unexpected imported collection names: %v", res)
	}
}
