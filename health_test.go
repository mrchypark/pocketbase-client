package pocketbase

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/goccy/go-json"
)

func TestClientHealthCheck(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/api/health" {
			t.Fatalf("unexpected path: %s", r.URL.Path)
		}
		_ = json.NewEncoder(w).Encode(map[string]interface{}{"code": "200"})
	}))
	defer srv.Close()

	c := NewClient(srv.URL)
	m, err := c.HealthCheck(context.Background())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if m["code"] != "200" {
		t.Fatalf("unexpected response: %v", m)
	}
}
