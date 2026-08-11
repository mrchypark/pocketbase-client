package pocketbase

import (
	"context"
	"fmt"
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

	// Delete is now a method on TypedRecordService.
	if err := svc.Delete(ctx, "rec2"); err != nil {
		t.Fatalf("Delete error: %v", err)
	}
}

func TestTypedRecordService_GetAll_SkipTotal(t *testing.T) {
	t.Parallel()

	const perPage = 100
	var requests int
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		requests++

		if got := r.URL.Query().Get("skipTotal"); got != "1" {
			t.Fatalf("skipTotal got %q, want %q", got, "1")
		}

		page := 1
		_ = json.Unmarshal([]byte(r.URL.Query().Get("page")), &page)

		const total = 30
		start := (page - 1) * perPage
		if start > total {
			start = total
		}
		items := make([]map[string]any, 0, total-start)
		for i := start; i < total; i++ {
			items = append(items, map[string]any{"id": fmt.Sprintf("rec%d", i+1), "name": "x"})
		}

		// With skipTotal, PocketBase reports totalPages as -1.
		_ = json.NewEncoder(w).Encode(map[string]any{
			"page":       page,
			"perPage":    perPage,
			"totalItems": total,
			"totalPages": -1,
			"items":      items,
		})
	}))
	t.Cleanup(srv.Close)

	client := NewClient(srv.URL)
	svc := NewTypedRecordService[testModel](client, "tests")

	all, err := svc.GetAll(context.Background(), &ListOptions{SkipTotal: true, PerPage: perPage})
	if err != nil {
		t.Fatalf("GetAll error: %v", err)
	}
	if len(all) != 30 {
		t.Errorf("GetAll returned %d records, want 30", len(all))
	}
	// With short-page guard, we stop after page 1 because 30 items < perPage (100).
	// No need to make a second request to discover there are no more items.
	if requests != 1 {
		t.Errorf("expected 1 request (page 1 only), got %d", requests)
	}
}

func TestTypedRecordService_NilBody(t *testing.T) {
	t.Parallel()

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		t.Fatalf("unexpected request: %s %s", r.Method, r.URL.String())
	}))
	t.Cleanup(srv.Close)

	client := NewClient(srv.URL)
	svc := NewTypedRecordService[testModel](client, "tests")
	ctx := context.Background()

	if _, err := svc.Create(ctx, nil); err == nil {
		t.Error("Create(nil) expected an error, got nil")
	}
	if _, err := svc.Update(ctx, "rec2", nil); err == nil {
		t.Error("Update(nil) expected an error, got nil")
	}
}
