package pocketbase

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/goccy/go-json"
)

// TestLogServiceGetRequestsList tests the GetRequestsList method of LogService.
func TestLogServiceGetRequestsList(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/api/logs/requests" {
			t.Fatalf("unexpected path: %s", r.URL.Path)
		}
		if err := json.NewEncoder(w).Encode(ListResult{}); err != nil {
			t.Fatalf("encode error: %v", err)
		}
	}))
	defer srv.Close()

	c := NewClient(srv.URL)
	if _, err := c.Logs.GetRequestsList(context.Background(), nil); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestLogServiceGetRequest(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/api/logs/requests/1" {
			t.Fatalf("unexpected path: %s", r.URL.Path)
		}
		if err := json.NewEncoder(w).Encode(map[string]interface{}{"id": "1"}); err != nil {
			t.Fatalf("encode error: %v", err)
		}
	}))
	defer srv.Close()

	c := NewClient(srv.URL)
	res, err := c.Logs.GetRequest(context.Background(), "1")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if res["id"] != "1" {
		t.Fatalf("unexpected response: %v", res)
	}
}

func TestLogServiceGetStats(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			t.Fatalf("unexpected method: %s", r.Method)
		}
		if r.URL.Path != "/api/logs/stats" {
			t.Fatalf("unexpected path: %s", r.URL.Path)
		}
		if err := json.NewEncoder(w).Encode(LogStats{Total: 1, Items: []LogStatItem{{Time: "t", Count: 1}}}); err != nil {
			t.Fatalf("encode error: %v", err)
		}
	}))
	defer srv.Close()

	c := NewClient(srv.URL)
	res, err := c.Logs.GetStats(context.Background())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if res.Total != 1 || len(res.Items) != 1 || res.Items[0].Count != 1 {
		t.Fatalf("unexpected result: %+v", res)
	}
}
func TestLogServiceGetRequestsListQueryParams(t *testing.T) {
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
			opts:      &ListOptions{Filter: "status=200"},
			wantQuery: "filter=status%3D200",
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
			if _, err := c.Logs.GetRequestsList(context.Background(), tc.opts); err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
		})
	}
}
