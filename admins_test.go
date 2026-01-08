package pocketbase

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/goccy/go-json"
)

// TestAdminServiceGetList tests the GetList method of AdminService.
func TestAdminServiceGetList(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/api/admins" {
			t.Fatalf("unexpected path: %s", r.URL.Path)
		}
		if r.URL.Query().Get("page") != "2" {
			t.Fatalf("unexpected query: %s", r.URL.RawQuery)
		}
		_ = json.NewEncoder(w).Encode(ListResult{})
	}))
	defer srv.Close()

	c := NewClient(srv.URL)
	if _, err := c.Admins.GetList(context.Background(), &ListOptions{Page: 2}); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

// TestAdminServiceGetOne tests the GetOne method of AdminService.
func TestAdminServiceGetOne(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/api/admins/1" {
			t.Fatalf("unexpected path: %s", r.URL.Path)
		}
		_ = json.NewEncoder(w).Encode(Admin{ID: "1"})
	}))
	defer srv.Close()

	c := NewClient(srv.URL)
	adm, err := c.Admins.GetOne(context.Background(), "1")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if adm.ID != "1" {
		t.Fatalf("unexpected id: %s", adm.ID)
	}
}

func TestAdminServiceCreate(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			t.Fatalf("unexpected method: %s", r.Method)
		}
		_ = json.NewEncoder(w).Encode(Admin{ID: "2"})
	}))
	defer srv.Close()

	c := NewClient(srv.URL)
	adm, err := c.Admins.Create(context.Background(), map[string]string{"email": "a"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if adm.ID != "2" {
		t.Fatalf("unexpected id: %s", adm.ID)
	}
}

func TestAdminServiceUpdate(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/api/admins/1" {
			t.Fatalf("unexpected path: %s", r.URL.Path)
		}
		if r.Method != http.MethodPatch {
			t.Fatalf("unexpected method: %s", r.Method)
		}
		_ = json.NewEncoder(w).Encode(Admin{ID: "1"})
	}))
	defer srv.Close()

	c := NewClient(srv.URL)
	adm, err := c.Admins.Update(context.Background(), "1", map[string]string{"email": "b"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if adm.ID != "1" {
		t.Fatalf("unexpected id: %s", adm.ID)
	}
}

func TestAdminServiceDelete(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/api/admins/1" {
			t.Fatalf("unexpected path: %s", r.URL.Path)
		}
		if r.Method != http.MethodDelete {
			t.Fatalf("unexpected method: %s", r.Method)
		}
		w.WriteHeader(http.StatusNoContent)
	}))
	defer srv.Close()

	c := NewClient(srv.URL)
	if err := c.Admins.Delete(context.Background(), "1"); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}
func TestAdminServiceGetListQueryParams(t *testing.T) {
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
			opts:      &ListOptions{Filter: "email='test@example.com'"},
			wantQuery: "filter=email%3D%27test%40example.com%27",
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
				_ = json.NewEncoder(w).Encode(ListResult{})
			}))
			defer srv.Close()

			c := NewClient(srv.URL)
			if _, err := c.Admins.GetList(context.Background(), tc.opts); err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
		})
	}
}
