package pocketbase

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/goccy/go-json"
)

type testModel struct {
	ID   string `json:"id"`
	Name string `json:"name"`
}

func (m *testModel) ToMap() map[string]any {
	return map[string]any{
		"name": m.Name,
	}
}

func TestTypedRecordService_CRUD(t *testing.T) {
	t.Parallel()

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")

		switch {
		case r.Method == http.MethodGet && r.URL.Path == "/api/collections/tests/records/rec1":
			if got := r.URL.Query().Get("expand"); got != "rel" {
				t.Fatalf("expand got %q, want %q", got, "rel")
			}
			_ = json.NewEncoder(w).Encode(map[string]any{
				"id":   "rec1",
				"name": "one",
			})
			return

		case r.Method == http.MethodGet && r.URL.Path == "/api/collections/tests/records":
			if got := r.URL.Query().Get("page"); got != "2" {
				t.Fatalf("page got %q, want %q", got, "2")
			}
			if got := r.URL.Query().Get("perPage"); got != "10" {
				t.Fatalf("perPage got %q, want %q", got, "10")
			}
			_ = json.NewEncoder(w).Encode(map[string]any{
				"page":       2,
				"perPage":    10,
				"totalItems": 1,
				"totalPages": 1,
				"items": []map[string]any{
					{"id": "rec1", "name": "one"},
				},
			})
			return

		case r.Method == http.MethodPost && r.URL.Path == "/api/collections/tests/records":
			var body map[string]any
			_ = json.NewDecoder(r.Body).Decode(&body)
			if got := body["name"]; got != "created" {
				t.Fatalf("create body.name got %#v, want %#v", got, "created")
			}
			_ = json.NewEncoder(w).Encode(map[string]any{
				"id":   "rec2",
				"name": "created",
			})
			return

		case r.Method == http.MethodPatch && r.URL.Path == "/api/collections/tests/records/rec2":
			var body map[string]any
			_ = json.NewDecoder(r.Body).Decode(&body)
			if got := body["name"]; got != "updated" {
				t.Fatalf("update body.name got %#v, want %#v", got, "updated")
			}
			_ = json.NewEncoder(w).Encode(map[string]any{
				"id":   "rec2",
				"name": "updated",
			})
			return

		case r.Method == http.MethodDelete && r.URL.Path == "/api/collections/tests/records/rec2":
			w.WriteHeader(http.StatusNoContent)
			return

		default:
			t.Fatalf("unexpected request: %s %s", r.Method, r.URL.String())
		}
	}))
	t.Cleanup(srv.Close)

	client := NewClient(srv.URL)
	svc := NewTypedRecordService[testModel](client, "tests")

	ctx := context.Background()

	gotOne, err := svc.GetOne(ctx, "rec1", &GetOneOptions{Expand: "rel"})
	if err != nil {
		t.Fatalf("GetOne error: %v", err)
	}
	if gotOne.ID != "rec1" || gotOne.Name != "one" {
		t.Fatalf("GetOne got %#v", gotOne)
	}

	gotList, err := svc.GetList(ctx, &ListOptions{Page: 2, PerPage: 10})
	if err != nil {
		t.Fatalf("GetList error: %v", err)
	}
	if gotList.Page != 2 || gotList.PerPage != 10 || len(gotList.Items) != 1 {
		t.Fatalf("GetList got %#v", gotList)
	}
	if gotList.Items[0].ID != "rec1" {
		t.Fatalf("GetList item got %#v", gotList.Items[0])
	}

	created, err := svc.Create(ctx, &testModel{Name: "created"})
	if err != nil {
		t.Fatalf("Create error: %v", err)
	}
	if created.ID != "rec2" || created.Name != "created" {
		t.Fatalf("Create got %#v", created)
	}

	updated, err := svc.Update(ctx, "rec2", &testModel{Name: "updated"})
	if err != nil {
		t.Fatalf("Update error: %v", err)
	}
	if updated.ID != "rec2" || updated.Name != "updated" {
		t.Fatalf("Update got %#v", updated)
	}

	// Delete is provided by the embedded RecordService.
	if err := svc.Delete(ctx, "tests", "rec2"); err != nil {
		t.Fatalf("Delete error: %v", err)
	}
}
