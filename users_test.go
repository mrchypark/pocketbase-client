package pocketbase

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/goccy/go-json"
)

func TestUserServiceRequestPasswordReset(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			t.Fatalf("unexpected method: %s", r.Method)
		}
		if r.URL.Path != "/api/collections/users/request-password-reset" {
			t.Fatalf("unexpected path: %s", r.URL.Path)
		}
		w.WriteHeader(http.StatusNoContent)
	}))
	defer srv.Close()

	c := NewClient(srv.URL)
	if err := c.Users.RequestPasswordReset(context.Background(), "users", "a@example.com"); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestUserServiceConfirmPasswordReset(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/api/collections/users/confirm-password-reset" {
			t.Fatalf("unexpected path: %s", r.URL.Path)
		}
		w.WriteHeader(http.StatusNoContent)
	}))
	defer srv.Close()

	c := NewClient(srv.URL)
	err := c.Users.ConfirmPasswordReset(context.Background(), "users", "tok", "p", "p")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestUserServiceRequestVerification(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/api/collections/users/request-verification" {
			t.Fatalf("unexpected path: %s", r.URL.Path)
		}
		w.WriteHeader(http.StatusNoContent)
	}))
	defer srv.Close()

	c := NewClient(srv.URL)
	if err := c.Users.RequestVerification(context.Background(), "users", "e@example.com"); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestUserServiceConfirmVerification(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/api/collections/users/confirm-verification" {
			t.Fatalf("unexpected path: %s", r.URL.Path)
		}
		w.WriteHeader(http.StatusNoContent)
	}))
	defer srv.Close()

	c := NewClient(srv.URL)
	if err := c.Users.ConfirmVerification(context.Background(), "users", "tok"); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestUserServiceGetOAuth2Providers(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/api/collections/users/auth-methods" {
			t.Fatalf("unexpected path: %s", r.URL.Path)
		}
		_ = json.NewEncoder(w).Encode(map[string]interface{}{"google": true})
	}))
	defer srv.Close()

	c := NewClient(srv.URL)
	p, err := c.Users.GetOAuth2Providers(context.Background(), "users")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !p["google"].(bool) {
		t.Fatalf("unexpected providers: %v", p)
	}
}

func TestUserServiceAuthWithOAuth2(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/api/collections/users/auth-with-oauth2" {
			t.Fatalf("unexpected path: %s", r.URL.Path)
		}
		_ = json.NewEncoder(w).Encode(AuthResponse{Token: "tok"})
	}))
	defer srv.Close()

	c := NewClient(srv.URL)
	_, err := c.Users.AuthWithOAuth2(context.Background(), "users", "google", "code", "ver", "url", nil)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestUserServiceAuthRefresh(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/api/collections/users/auth-refresh" {
			t.Fatalf("unexpected path: %s", r.URL.Path)
		}
		if r.Header.Get("Authorization") != "tok" {
			t.Fatalf("missing auth header: %s", r.Header.Get("Authorization"))
		}
		_ = json.NewEncoder(w).Encode(AuthResponse{Token: "new"})
	}))
	defer srv.Close()

	c := NewClient(srv.URL)
	c.AuthStore.Set("tok", &Record{CollectionName: "users"})
	_, err := c.Users.AuthRefresh(context.Background(), "users")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestUserServiceRequestOTP(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/api/collections/users/request-otp" {
			t.Fatalf("unexpected path: %s", r.URL.Path)
		}
		_ = json.NewEncoder(w).Encode(map[string]string{"otpId": "id"})
	}))
	defer srv.Close()

	c := NewClient(srv.URL)
	if _, err := c.Users.RequestOTP(context.Background(), "users", "e@x.com"); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestUserServiceAuthWithOTP(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/api/collections/users/auth-with-otp" {
			t.Fatalf("unexpected path: %s", r.URL.Path)
		}
		_ = json.NewEncoder(w).Encode(AuthResponse{Token: "tok"})
	}))
	defer srv.Close()

	c := NewClient(srv.URL)
	_, err := c.Users.AuthWithOTP(context.Background(), "users", "id", "pass")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestUserServiceRequestEmailChange(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/api/collections/users/request-email-change" {
			t.Fatalf("unexpected path: %s", r.URL.Path)
		}
		w.WriteHeader(http.StatusNoContent)
	}))
	defer srv.Close()

	c := NewClient(srv.URL)
	c.AuthStore.Set("tok", &Record{CollectionName: "users"})
	if err := c.Users.RequestEmailChange(context.Background(), "users", "a@b.com"); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestUserServiceConfirmEmailChange(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/api/collections/users/confirm-email-change" {
			t.Fatalf("unexpected path: %s", r.URL.Path)
		}
		w.WriteHeader(http.StatusNoContent)
	}))
	defer srv.Close()

	c := NewClient(srv.URL)
	if err := c.Users.ConfirmEmailChange(context.Background(), "users", "tok", "p"); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestUserServiceImpersonate(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/api/collections/users/impersonate/abc" {
			t.Fatalf("unexpected path: %s", r.URL.Path)
		}
		_ = json.NewEncoder(w).Encode(AuthResponse{Token: "imp"})
	}))
	defer srv.Close()

	c := NewClient(srv.URL)
	c.AuthStore.Set("tok", &Record{CollectionName: "users"})
	imp, err := c.Users.Impersonate(context.Background(), "users", "abc", 0)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	tok, err := imp.AuthStore.Token()
	if err != nil || tok != "imp" {
		t.Fatalf("unexpected token: %s", tok)
	}
}
