package pocketbase

import (
	"context"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/goccy/go-json"
)

func TestCollectionServiceGetList(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/api/collections" {
			t.Fatalf("unexpected path: %s", r.URL.Path)
		}
		if r.URL.Query().Get("page") != "2" {
			t.Fatalf("unexpected query: %s", r.URL.RawQuery)
		}
		w.Header().Set("Content-Type", "application/json")
		fmt.Fprintln(w, `{
                "page": 2,
                "perPage": 30,
                "totalItems": 1,
                "totalPages": 1,
                "items": [{"id":"123","name":"posts"}]
            }`)
	}))
	defer srv.Close()

	c := NewClient(srv.URL)
	res, err := c.Collections.GetList(context.Background(), &ListOptions{Page: 2})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(res.Items) != 1 {
		t.Fatalf("expected 1 item, got %d", len(res.Items))
	}
	if res.Items[0].Name != "posts" {
		t.Fatalf("unexpected collection name: %s", res.Items[0].Name)
	}
}

func TestCollectionServiceGetOne(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/api/collections/posts" {
			t.Fatalf("unexpected path: %s", r.URL.Path)
		}
		_ = json.NewEncoder(w).Encode(Collection{Name: "posts"})
	}))
	defer srv.Close()

	c := NewClient(srv.URL)
	col, err := c.Collections.GetOne(context.Background(), "posts")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if col.Name != "posts" {
		t.Fatalf("unexpected name: %s", col.Name)
	}
}

func TestCollectionServiceCreate(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			t.Fatalf("unexpected method: %s", r.Method)
		}
		var req Collection
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			t.Fatalf("could not decode request: %v", err)
		}
		if req.Name != "posts" {
			t.Fatalf("expected name 'posts' in request, got '%s'", req.Name)
		}
		_ = json.NewEncoder(w).Encode(Collection{Name: "posts"})
	}))
	defer srv.Close()

	c := NewClient(srv.URL)
	col, err := c.Collections.Create(context.Background(), &Collection{Name: "posts"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if col.Name != "posts" {
		t.Fatalf("unexpected name: %s", col.Name)
	}
}

func TestCollectionServiceUpdate(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/api/collections/posts" {
			t.Fatalf("unexpected path: %s", r.URL.Path)
		}
		if r.Method != http.MethodPatch {
			t.Fatalf("unexpected method: %s", r.Method)
		}
		var req Collection
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			t.Fatalf("could not decode request: %v", err)
		}
		if req.Name != "posts" {
			t.Fatalf("expected name 'posts' in request, got '%s'", req.Name)
		}
		_ = json.NewEncoder(w).Encode(Collection{Name: "posts"})
	}))
	defer srv.Close()

	c := NewClient(srv.URL)
	col, err := c.Collections.Update(context.Background(), "posts", &Collection{Name: "posts"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if col.Name != "posts" {
		t.Fatalf("unexpected name: %s", col.Name)
	}
}

func TestCollectionServiceDelete(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/api/collections/posts" {
			t.Fatalf("unexpected path: %s", r.URL.Path)
		}
		if r.Method != http.MethodDelete {
			t.Fatalf("unexpected method: %s", r.Method)
		}
		w.WriteHeader(http.StatusNoContent)
	}))
	defer srv.Close()

	c := NewClient(srv.URL)
	if err := c.Collections.Delete(context.Background(), "posts"); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestCollectionServiceImport(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPut {
			t.Fatalf("unexpected method: %s", r.Method)
		}
		if r.URL.Path != "/api/collections/import" {
			t.Fatalf("unexpected path: %s", r.URL.Path)
		}
		if r.URL.Query().Get("deleteMissing") != "1" {
			t.Fatalf("missing deleteMissing query")
		}
		var v []*Collection
		if err := json.NewDecoder(r.Body).Decode(&v); err != nil {
			t.Fatalf("failed decode: %v", err)
		}
		if len(v) != 1 || v[0].Name != "posts" {
			t.Fatalf("unexpected body: %v", v)
		}
		_ = json.NewEncoder(w).Encode([]*Collection{})
	}))
	defer srv.Close()

	c := NewClient(srv.URL)
	cols := []*Collection{{Name: "posts"}}
	returned, err := c.Collections.Import(context.Background(), cols, true)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if returned == nil {
		t.Fatal("expected collections, got nil")
	}
	if len(returned) != 0 {
		t.Fatalf("expected 0 collections, got %d", len(returned))
	}
}
func TestCollectionServiceGetListQueryParams(t *testing.T) {
	testCases := []struct {
		name      string
		opts      *ListOptions
		wantQuery string
	}{
		{
			name:      "with PerPage",
			opts:      &ListOptions{PerPage: 5},
			wantQuery: "perPage=5",
		},
		{
			name:      "with Page and Sort",
			opts:      &ListOptions{Page: 2, Sort: "-created"},
			wantQuery: "page=2&sort=-created",
		},
		{
			name:      "with Filter",
			opts:      &ListOptions{Filter: "type='base'"},
			wantQuery: "filter=type%3D%27base%27",
		},
		{
			name:      "with nil options",
			opts:      nil,
			wantQuery: "",
		},
		{
			name: "with all options",
			opts: &ListOptions{
				Page:      3,
				PerPage:   10,
				Sort:      "email",
				Filter:    "avatar>0",
				SkipTotal: true,
			},
			wantQuery: "filter=avatar%3E0&page=3&perPage=10&skipTotal=1&sort=email",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				gotQuery := r.URL.RawQuery
				if gotQuery != tc.wantQuery {
					t.Fatalf("unexpected query params:\ngot:  %q\nwant: %q", gotQuery, tc.wantQuery)
				}
				fmt.Fprintln(w, `{"page":1,"perPage":1,"totalItems":0,"totalPages":0,"items":[]}`)
			}))
			defer srv.Close()

			c := NewClient(srv.URL)
			if _, err := c.Collections.GetList(context.Background(), tc.opts); err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
		})
	}
}
