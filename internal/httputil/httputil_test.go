package httputil

import (
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestDoSuccess(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		_, _ = io.WriteString(w, "ok")
	}))
	defer srv.Close()
	req, err := JSONRequest(http.MethodGet, srv.URL, nil)
	if err != nil {
		t.Fatal(err)
	}
	body, err := Do(&http.Client{}, req)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if string(body) != "ok" {
		t.Fatalf("unexpected body: %s", string(body))
	}
}

func TestDoErrorStatus(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusBadRequest)
		_, _ = io.WriteString(w, "bad")
	}))
	defer srv.Close()
	req, err := JSONRequest(http.MethodGet, srv.URL, nil)
	if err != nil {
		t.Fatal(err)
	}
	if _, err := Do(&http.Client{}, req); err == nil {
		t.Fatal("expected error")
	}
}

func TestJSONRequest(t *testing.T) {
	req, err := JSONRequest(http.MethodPost, "http://example.com", map[string]string{"a": "b"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if req.Header.Get("Content-Type") != "application/json" {
		t.Fatal("expected content-type")
	}
}

func BenchmarkJSONRequest(b *testing.B) {
	for i := 0; i < b.N; i++ {
		req, err := JSONRequest(http.MethodPost, "http://example.com", map[string]string{"a": "b"})
		if err != nil || req == nil {
			b.Fatal("request error")
		}
	}
}

func TestDoServerError(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
		_, _ = io.WriteString(w, "fail")
	}))
	defer srv.Close()
	req, err := JSONRequest(http.MethodGet, srv.URL, nil)
	if err != nil {
		t.Fatal(err)
	}
	_, err = Do(&http.Client{}, req)
	if err == nil {
		t.Fatal("expected error")
	}
	if !strings.Contains(err.Error(), "http error 500") {
		t.Fatalf("unexpected error: %v", err)
	}
}
