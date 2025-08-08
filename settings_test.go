package pocketbase

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/goccy/go-json"
)

// TestSettingServiceGetAll tests the GetAll method of SettingService.
func TestSettingServiceGetAll(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/api/settings" {
			t.Fatalf("unexpected path: %s", r.URL.Path)
		}
		if err := json.NewEncoder(w).Encode(map[string]any{"appName": "pb"}); err != nil {
			t.Fatalf("encode error: %v", err)
		}
	}))
	defer srv.Close()

	c := NewClient(srv.URL)
	m, err := c.Settings.GetAll(context.Background())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if m["appName"] != "pb" {
		t.Fatalf("unexpected value: %v", m)
	}
}

func TestSettingServiceUpdate(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPatch {
			t.Fatalf("unexpected method: %s", r.Method)
		}
		if err := json.NewEncoder(w).Encode(map[string]any{"appName": "pb"}); err != nil {
			t.Fatalf("encode error: %v", err)
		}
	}))
	defer srv.Close()

	c := NewClient(srv.URL)
	m, err := c.Settings.Update(context.Background(), map[string]any{"appName": "pb"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if m["appName"] != "pb" {
		t.Fatalf("unexpected value: %v", m)
	}
}

func TestSettingServiceTestS3(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			t.Fatalf("unexpected method: %s", r.Method)
		}
		if r.URL.Path != "/api/settings/test/s3" {
			t.Fatalf("unexpected path: %s", r.URL.Path)
		}
		if err := json.NewEncoder(w).Encode(map[string]any{"ok": true}); err != nil {
			t.Fatalf("encode error: %v", err)
		}
	}))
	defer srv.Close()

	c := NewClient(srv.URL)
	res, err := c.Settings.TestS3(context.Background())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	okVal, ok := res["ok"].(bool)
	if !ok || !okVal {
		t.Fatalf("unexpected response: %v", res)
	}
}

func TestSettingServiceTestEmail(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			t.Fatalf("unexpected method: %s", r.Method)
		}
		if r.URL.Path != "/api/settings/test/email" {
			t.Fatalf("unexpected path: %s", r.URL.Path)
		}
		var body map[string]any
		if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
			t.Fatalf("invalid body: %v", err)
		}
		if body["email"] != "a@b.com" {
			t.Fatalf("unexpected email: %v", body)
		}
		if err := json.NewEncoder(w).Encode(map[string]any{"ok": true}); err != nil {
			t.Fatalf("encode error: %v", err)
		}
	}))
	defer srv.Close()

	c := NewClient(srv.URL)
	res, err := c.Settings.TestEmail(context.Background(), "a@b.com")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	okVal, ok := res["ok"].(bool)
	if !ok || !okVal {
		t.Fatalf("unexpected response: %v", res)
	}
}

func TestSettingServiceGenerateAppleClientSecret(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			t.Fatalf("unexpected method: %s", r.Method)
		}
		if r.URL.Path != "/api/settings/apple/generate-client-secret" {
			t.Fatalf("unexpected path: %s", r.URL.Path)
		}
		var body map[string]any
		if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
			t.Fatalf("invalid body: %v", err)
		}
		if body["clientId"] != "id" {
			t.Fatalf("unexpected body: %v", body)
		}
		if err := json.NewEncoder(w).Encode(map[string]any{"secret": "s"}); err != nil {
			t.Fatalf("encode error: %v", err)
		}
	}))
	defer srv.Close()

	c := NewClient(srv.URL)
	res, err := c.Settings.GenerateAppleClientSecret(context.Background(), map[string]any{"clientId": "id"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	secret, ok := res["secret"].(string)
	if !ok || secret != "s" {
		t.Fatalf("unexpected response: %v", res)
	}
}
