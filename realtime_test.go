package pocketbase

import (
	"bytes"
	"context"
	"io"
	"net/http"
	"net/http/httptest"
	"sync"
	"testing"
	"time"
)

func TestRealtimeServiceSubscribe(t *testing.T) {
	var mu sync.Mutex
	var postBody []byte
	postReceived := make(chan struct{})

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.Method {
		case http.MethodGet:
			if r.URL.Path != "/api/realtime" {
				t.Fatalf("unexpected path: %s", r.URL.Path)
			}
			w.Header().Set("Content-Type", "text/event-stream")
			flusher, ok := w.(http.Flusher)
			if !ok {
				t.Fatal("no flusher")
			}
			_, _ = io.WriteString(w, "event: PB_CONNECT\ndata: {\"clientId\":\"test-client-id\"}\n\n")
            flusher.Flush()
            <-postReceived
            _, _ = io.WriteString(w, "data: {\"action\":\"update\"}\n\n")
            flusher.Flush()
		case http.MethodPost:
			if r.URL.Path != "/api/realtime" {
				t.Fatalf("unexpected path: %s", r.URL.Path)
			}
			data, err := io.ReadAll(r.Body)
			if err != nil {
				t.Fatalf("read body: %v", err)
			}
			mu.Lock()
			postBody = data
			mu.Unlock()
			w.WriteHeader(http.StatusNoContent)
			close(postReceived)
		default:
			t.Fatalf("unexpected method: %s", r.Method)
		}
	}))
	defer srv.Close()

	c := NewClient(srv.URL)
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()

	done := make(chan *RealtimeEvent, 1)
	unsub, err := c.Realtime.Subscribe(ctx, []string{"posts"}, func(ev *RealtimeEvent, err error) {
		if err != nil {
			t.Fatalf("callback error: %v", err)
		}
		done <- ev
	})
	if err != nil {
		t.Fatalf("subscribe err: %v", err)
	}

	ev := <-done
	unsub()
	if ev == nil || ev.Action != "update" {
		t.Fatalf("unexpected event: %+v", ev)
	}

	mu.Lock()
	body := append([]byte(nil), postBody...)
	mu.Unlock()
	if !bytes.Contains(body, []byte("posts")) {
		t.Fatalf("subscription body missing: %s", string(body))
	}
}
