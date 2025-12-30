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

// TestAuthenticateAsAdmin tests the AuthenticateAsAdmin method.
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
	res, err := c.WithAdminPassword(context.Background(), "a", "b")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if res.Token != token {
		t.Fatalf("unexpected token: %s", res.Token)
	}
	tok, _ := c.AuthStore.Token(c)
	if tok != token {
		t.Fatal("token not stored")
	}
	if res.Admin.Email != "admin@example.com" {
		t.Fatalf("unexpected admin email: %s", res.Admin.Email)
	}
}

// TestAuthenticateWithPassword tests the AuthenticateWithPassword method.
func TestAuthenticateWithPassword(t *testing.T) {
	token := "user"
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/api/collections/users/auth-with-password" {
			t.Fatalf("unexpected path: %s", r.URL.Path)
		}
		resp := map[string]any{
			"token": token,
			"record": map[string]any{
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
	res, err := c.WithPassword(context.Background(), "users", "e", "p")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	tok, _ := c.AuthStore.Token(c)
	if tok != token {
		t.Fatal("token not stored")
	}
	u := res.Record.GetString("username")
	if u != "user1" {
		t.Fatalf("unexpected username: %v", u)
	}
}

func TestAuthenticateAsAdminBadCredentials(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusBadRequest)
		_ = json.NewEncoder(w).Encode(map[string]any{"code": 400, "message": "invalid"})
	}))
	defer srv.Close()
	c := NewClient(srv.URL)
	_, err := c.WithAdminPassword(context.Background(), "a", "b")
	if err == nil {
		t.Fatal("expected error")
	}

	// Check for the new Error type
	var pbErr *Error
	if !errors.As(err, &pbErr) {
		t.Fatalf("expected error to be *Error, but it was not (got %T)", err)
	}
	if pbErr.Status != 400 {
		t.Fatalf("unexpected status: %d", pbErr.Status)
	}

	// Check using helper function - for 400 errors, we can check if it's a validation error
	if !IsValidationError(err) {
		// If not validation error, check if it's at least a bad request
		if !pbErr.IsBadRequest() {
			t.Fatalf("expected error to be a bad request error, got status %d", pbErr.Status)
		}
	}
}

type errorRoundTripper struct{}

func (errorRoundTripper) RoundTrip(*http.Request) (*http.Response, error) {
	return nil, errors.New("network down")
}

func TestAuthenticateWithPasswordNetworkError(t *testing.T) {
	c := NewClient("http://example.com")
	c.HTTPClient = &http.Client{Transport: errorRoundTripper{}}
	_, err := c.WithPassword(context.Background(), "users", "e", "p")
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

func TestSendWithWriter_ContextCancellationStopsStreaming(t *testing.T) {
	wroteFirstChunk := make(chan struct{})
	handlerExited := make(chan struct{})

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		defer close(handlerExited)

		flusher, ok := w.(http.Flusher)
		if !ok {
			t.Fatal("no flusher")
		}

		w.Header().Set("Content-Type", "text/plain")
		w.WriteHeader(http.StatusOK)

		for i := 0; ; i++ {
			select {
			case <-r.Context().Done():
				return
			default:
			}

			_, _ = w.Write([]byte("chunk\n"))
			flusher.Flush()

			if i == 0 {
				close(wroteFirstChunk)
			}

			time.Sleep(10 * time.Millisecond)
		}
	}))
	defer srv.Close()

	c := NewClient(srv.URL)
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	var buf bytes.Buffer
	errCh := make(chan error, 1)
	go func() {
		errCh <- c.SendWithOptions(ctx, http.MethodGet, "/", nil, nil, WithResponseWriter(&buf))
	}()

	select {
	case <-wroteFirstChunk:
	case <-time.After(2 * time.Second):
		t.Fatal("timed out waiting for server to write first chunk")
	}

	cancel()

	select {
	case err := <-errCh:
		if err == nil {
			t.Fatal("expected error")
		}
		if !errors.Is(err, context.Canceled) {
			t.Fatalf("expected context cancellation error, got %v", err)
		}
	case <-time.After(2 * time.Second):
		t.Fatal("timed out waiting for client to return after cancellation")
	}

	select {
	case <-handlerExited:
	case <-time.After(2 * time.Second):
		t.Fatal("timed out waiting for server handler to exit after cancellation")
	}
}

// flushWriter is an io.Writer that counts each time Flush is called.
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
	c.AuthStore = NewTokenAuth("tok")
	err := c.Send(context.Background(), http.MethodGet, "/", nil, nil)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if rt.lastAuth != "tok" {
		t.Fatalf("auth header not injected: %s", rt.lastAuth)
	}
}
