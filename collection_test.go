package pocketbase

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"reflect"
	"testing"
)

func TestCollectionModernWireRoundTrip(t *testing.T) {
	fixture := `{"id":"pbc_123","created":"2026-10-07 01:02:03.000Z","updated":"2026-10-07 01:02:03.000Z","name":"members","type":"auth","system":false,"fields":[{"id":"number123","name":"score","type":"number","required":true,"presentable":false,"system":false,"hidden":false,"min":0,"max":100,"onlyInt":true},{"id":"relation123","name":"friends","type":"relation","required":false,"presentable":false,"system":false,"hidden":false,"collectionId":"pbc_456","maxSelect":3,"cascadeDelete":false}],"indexes":["CREATE INDEX idx_score ON members (score)"],"listRule":null,"viewRule":"","createRule":null,"updateRule":"id = @request.auth.id","deleteRule":null,"manageRule":null,"authRule":"verified = true","passwordAuth":{"enabled":true,"identityFields":["email"]},"viewQuery":"SELECT id FROM members"}`
	var col Collection
	if err := json.Unmarshal([]byte(fixture), &col); err != nil {
		t.Fatal(err)
	}
	if col.GetID() != "pbc_123" || col.GetCollectionName() != "members" || col.ListRule != nil || col.ViewRule == nil || *col.ViewRule != "" {
		t.Fatalf("incorrect identity/rules: %+v", col)
	}
	if len(col.Fields) != 2 || col.Fields[0].Options["max"] != float64(100) || col.Fields[1].Options["maxSelect"] != float64(3) {
		t.Fatalf("lost field options: %+v", col.Fields)
	}
	data, err := json.Marshal(col)
	if err != nil {
		t.Fatal(err)
	}
	var want, got map[string]any
	if err := json.Unmarshal([]byte(fixture), &want); err != nil {
		t.Fatal(err)
	}
	if err := json.Unmarshal(data, &got); err != nil {
		t.Fatal(err)
	}
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("round trip changed contract:\ngot %s\nwant %s", data, fixture)
	}
}

func TestCollectionUpdateNullableRules(t *testing.T) {
	public := ""
	for _, rule := range []*string{nil, &public} {
		srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			var body map[string]json.RawMessage
			if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
				t.Error(err)
			}
			want := "null"
			if rule != nil {
				want = `""`
			}
			for _, key := range []string{"listRule", "viewRule", "createRule", "updateRule", "deleteRule"} {
				if string(body[key]) != want {
					t.Errorf("%s: got %s want %s", key, body[key], want)
				}
			}
			if string(body["indexes"]) != "[]" {
				t.Errorf("cannot clear indexes: %s", body["indexes"])
			}
			_, _ = w.Write([]byte(`{"id":"pbc_123","name":"posts"}`))
		}))
		col := &Collection{Name: "posts", ListRule: rule, ViewRule: rule, CreateRule: rule, UpdateRule: rule, DeleteRule: rule, Indexes: []string{}, Options: map[string]any{"listRule": ""}}
		_, err := NewClient(srv.URL).Collections.Update(context.Background(), "posts", col)
		srv.Close()
		if err != nil {
			t.Fatal(err)
		}
	}
}

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

		var body struct {
			Collections   []*Collection `json:"collections"`
			DeleteMissing bool          `json:"deleteMissing"`
		}
		if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
			t.Error(err)
		}
		if len(body.Collections) != 2 || !body.DeleteMissing || r.URL.RawQuery != "" {
			t.Errorf("unexpected import body/query: %+v %s", body, r.URL.RawQuery)
		}
		w.WriteHeader(http.StatusNoContent)
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
	if res != nil {
		t.Fatalf("expected nil after 204, got %v", res)
	}
}
