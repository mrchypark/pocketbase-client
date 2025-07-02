package pocketbase

import (
	"bytes"
	"context"
	"errors"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/goccy/go-json"
)

func TestAuthenticateAsAdmin(t *testing.T) {
	token := "abc"
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/api/collections/_superusers/auth-with-password" {
			t.Fatalf("unexpected path: %s", r.URL.Path)
		}
		_ = json.NewEncoder(w).Encode(AuthResponse{
			Token: token,
			Admin: &Admin{Email: "admin@example.com"},
		})
	}))
	defer srv.Close()
	c := NewClient(srv.URL)
	res, err := c.AuthenticateAsAdmin(context.Background(), "a", "b")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if res.Token != token {
		t.Fatalf("unexpected token: %s", res.Token)
	}
	tok, _ := c.AuthStore.Token()
	if tok != token {
		t.Fatal("token not stored")
	}
	if res.Admin.Email != "admin@example.com" {
		t.Fatalf("unexpected admin email: %s", res.Admin.Email)
	}
}

func TestAuthenticateWithPassword(t *testing.T) {
	token := "user"
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/api/collections/users/auth-with-password" {
			t.Fatalf("unexpected path: %s", r.URL.Path)
		}
		resp := map[string]interface{}{
			"token": token,
			"record": map[string]interface{}{
				"id":             "1",
				"collectionId":   "users",
				"collectionName": "users",
				"username":       "user1",
			},
		}
		_ = json.NewEncoder(w).Encode(resp)
	}))
	defer srv.Close()
	c := NewClient(srv.URL)
	res, err := c.AuthenticateWithPassword(context.Background(), "users", "e", "p")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	tok, _ := c.AuthStore.Token()
	if tok != token {
		t.Fatal("token not stored")
	}
	if res.Record.Data["username"] != "user1" {
		t.Fatalf("unexpected username: %v", res.Record.Data["username"])
	}
}

func TestAuthenticateAsAdminBadCredentials(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusBadRequest)
		_ = json.NewEncoder(w).Encode(APIError{Code: 400, Message: "invalid"})
	}))
	defer srv.Close()
	c := NewClient(srv.URL)
	_, err := c.AuthenticateAsAdmin(context.Background(), "a", "b")
	if err == nil {
		t.Fatal("expected error")
	}
	apiErr, ok := err.(*APIError)
	if !ok {
		t.Fatalf("unexpected error type: %T", err)
	}
	if apiErr.Code != 400 {
		t.Fatalf("unexpected code: %d", apiErr.Code)
	}
}

type errorRoundTripper struct{}

func (errorRoundTripper) RoundTrip(*http.Request) (*http.Response, error) {
	return nil, errors.New("network down")
}

func TestAuthenticateWithPasswordNetworkError(t *testing.T) {
	c := NewClient("http://example.com")
	c.HTTPClient = &http.Client{Transport: errorRoundTripper{}}
	_, err := c.AuthenticateWithPassword(context.Background(), "users", "e", "p")
	if err == nil {
		t.Fatal("expected error")
	}
}

func TestSendStreamNetworkError(t *testing.T) {
	c := NewClient("http://example.com")
	c.HTTPClient = &http.Client{Transport: errorRoundTripper{}}
	rc, err := c.sendStream(context.Background(), http.MethodGet, "/", nil, "")
	if err == nil {
		t.Fatal("expected error")
	}
	if rc != nil {
		t.Fatal("expected nil reader")
	}
}

func TestNewClientWithHTTPClient(t *testing.T) {
	hc := &http.Client{Timeout: time.Second}
	c := NewClient("http://example.com", WithHTTPClient(hc))
	if c.HTTPClient != hc {
		t.Fatal("custom http client not set")
	}
}

func TestSendWithWriter(t *testing.T) {
	const chunk = "data-"
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		flusher, ok := w.(http.Flusher)
		if !ok {
			t.Fatal("no flusher")
		}
		for i := 0; i < 3; i++ {
			_, _ = w.Write([]byte(chunk))
			flusher.Flush()
		}
	}))
	defer srv.Close()

	c := NewClient(srv.URL)
	var buf bytes.Buffer
	err := c.SendWithOptions(context.Background(), http.MethodGet, "/", nil, nil, WithResponseWriter(&buf))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if buf.String() != strings.Repeat(chunk, 3) {
		t.Fatalf("unexpected buffer: %s", buf.String())
	}
}

// flushWriter는 Flush가 호출될 때마다 카운트하는 io.Writer입니다.
type flushWriter struct {
	bytes.Buffer
	flushCount int
}

func (fw *flushWriter) Flush() { fw.flushCount++ }

func TestCopyWithFlush(t *testing.T) {
	src := io.MultiReader(strings.NewReader("a"), strings.NewReader("b"), strings.NewReader("c"))
	fw := &flushWriter{}
	n, err := copyWithFlush(fw, src)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if n != 3 {
		t.Fatalf("unexpected written: %d", n)
	}
	if fw.String() != "abc" {
		t.Fatalf("unexpected buffer: %s", fw.String())
	}
	if fw.flushCount != 3 {
		t.Fatalf("unexpected flush count: %d", fw.flushCount)
	}
}

func TestSendWithWriterAndResponseDataError(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_ = json.NewEncoder(w).Encode(map[string]string{"msg": "ok"})
	}))
	defer srv.Close()

	c := NewClient(srv.URL)
	var resp map[string]string
	var buf bytes.Buffer
	err := c.SendWithOptions(context.Background(), http.MethodGet, "/", nil, &resp, WithResponseWriter(&buf))
	if err == nil {
		t.Fatal("expected error")
	}
}

func TestNewClientUsesAuthInjector(t *testing.T) {
	c := NewClient("http://example.com")
	if _, ok := c.HTTPClient.Transport.(*authInjector); !ok {
		t.Fatalf("unexpected transport type: %T", c.HTTPClient.Transport)
	}
}

type recordingRoundTripper struct{ lastAuth string }

func (rt *recordingRoundTripper) RoundTrip(req *http.Request) (*http.Response, error) {
	rt.lastAuth = req.Header.Get("Authorization")
	return &http.Response{
		StatusCode: http.StatusNoContent,
		Body:       io.NopCloser(strings.NewReader("")),
		Header:     make(http.Header),
	}, nil
}

func TestNewClientWithHTTPClientTransportWrapped(t *testing.T) {
	rt := &recordingRoundTripper{}
	hc := &http.Client{Transport: rt}
	c := NewClient("http://example.com", WithHTTPClient(hc))
	c.AuthStore.Set("tok", &Admin{})
	err := c.Send(context.Background(), http.MethodGet, "/", nil, nil)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if rt.lastAuth != "tok" {
		t.Fatalf("auth header not injected: %s", rt.lastAuth)
	}
}
