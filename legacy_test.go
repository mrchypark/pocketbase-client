package pocketbase

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/goccy/go-json"
)

// TestLegacyAdminAuthRefresh tests the AdminAuthRefresh method of LegacyService.
func TestLegacyAdminAuthRefresh(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/api/admins/auth-refresh" {
			t.Fatalf("unexpected path: %s", r.URL.Path)
		}
		_ = json.NewEncoder(w).Encode(AuthResponse{Token: "t"})
	}))
	defer srv.Close()

	c := NewClient(srv.URL)
	c.Legacy = &LegacyService{Client: c}
	res, err := c.Legacy.AdminAuthRefresh(context.Background())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	tok, _ := c.AuthStore.Token(c)
	if res.Token != "t" || tok != "t" {
		t.Fatalf("token not refreshed: %v", res.Token)
	}
}

// TestLegacyRecordAuthRefresh tests the RecordAuthRefresh method of LegacyService.
func TestLegacyRecordAuthRefresh(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/api/collections/posts/auth-refresh" {
			t.Fatalf("unexpected path: %s", r.URL.Path)
		}
		_ = json.NewEncoder(w).Encode(AuthResponse{Token: "t"})
	}))
	defer srv.Close()

	c := NewClient(srv.URL)
	c.Legacy = &LegacyService{Client: c}
	res, err := c.Legacy.RecordAuthRefresh(context.Background(), "posts")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if res.Token != "t" {
		t.Fatalf("unexpected token: %s", res.Token)
	}
}

// TestLegacyRequestEmailChange tests the RequestEmailChange method of LegacyService.
func TestLegacyRequestEmailChange(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			t.Fatalf("unexpected method: %s", r.Method)
		}
		if r.URL.Path != "/api/collections/users/request-email-change" {
			t.Fatalf("unexpected path: %s", r.URL.Path)
		}
		var body map[string]string
		if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
			t.Fatalf("invalid body: %v", err)
		}
		if body["newEmail"] != "a@b.c" {
			t.Fatalf("unexpected email: %v", body["newEmail"])
		}
		w.WriteHeader(http.StatusNoContent)
	}))
	defer srv.Close()

	c := NewClient(srv.URL)
	c.Legacy = &LegacyService{Client: c}
	if err := c.Legacy.RequestEmailChange(context.Background(), "users", "a@b.c"); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

// TestLegacyConfirmEmailChange tests the ConfirmEmailChange method of LegacyService.
func TestLegacyConfirmEmailChange(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/api/collections/users/confirm-email-change" {
			t.Fatalf("unexpected path: %s", r.URL.Path)
		}
		w.WriteHeader(http.StatusNoContent)
	}))
	defer srv.Close()

	c := NewClient(srv.URL)
	c.Legacy = &LegacyService{Client: c}
	if err := c.Legacy.ConfirmEmailChange(context.Background(), "users", "tok", "pass"); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

// TestLegacyListExternalAuths tests the ListExternalAuths method of LegacyService.
func TestLegacyListExternalAuths(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/api/collections/users/records/1/external-auths" {
			t.Fatalf("unexpected path: %s", r.URL.Path)
		}
		_ = json.NewEncoder(w).Encode([]map[string]any{{"id": "1"}})
	}))
	defer srv.Close()

	c := NewClient(srv.URL)
	c.Legacy = &LegacyService{Client: c}
	res, err := c.Legacy.ListExternalAuths(context.Background(), "users", "1")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(res) != 1 || res[0]["id"] != "1" {
		t.Fatalf("unexpected result: %v", res)
	}
}

func TestLegacyUnlinkExternalAuth(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodDelete {
			t.Fatalf("unexpected method: %s", r.Method)
		}
		if r.URL.Path != "/api/collections/users/records/1/external-auths/google" {
			t.Fatalf("unexpected path: %s", r.URL.Path)
		}
		w.WriteHeader(http.StatusNoContent)
	}))
	defer srv.Close()

	c := NewClient(srv.URL)
	c.Legacy = &LegacyService{Client: c}
	if err := c.Legacy.UnlinkExternalAuth(context.Background(), "users", "1", "google"); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestLegacyAuthWithOAuth2(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/api/collections/users/auth-with-oauth2" {
			t.Fatalf("unexpected path: %s", r.URL.Path)
		}
		_ = json.NewEncoder(w).Encode(AuthResponse{Token: "t"})
	}))
	defer srv.Close()

	c := NewClient(srv.URL)
	c.Legacy = &LegacyService{Client: c}
	res, err := c.Legacy.AuthWithOAuth2(context.Background(), "users", &OAuth2Request{Provider: "g", Code: "c", RedirectURL: "r"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if res.Token != "t" {
		t.Fatalf("unexpected token: %s", res.Token)
	}
}
