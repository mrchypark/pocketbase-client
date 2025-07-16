package pocketbase

import (
	"context"
	"io"
	"net/http"
	"net/http/httptest"
	"sync"
	"testing"
	"time"
)

// TestRealtimeSubscribeRace tests data races during concurrent subscribe/unsubscribe operations.
func TestRealtimeSubscribeRace(t *testing.T) {
	// Test server that handles real-time connections
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/event-stream")
		flusher, ok := w.(http.Flusher)
		if !ok {
			http.Error(w, "streaming unsupported", http.StatusInternalServerError)
			return
		}

		// Wait until client sends subscription request
		if r.Method == http.MethodPost {
			w.WriteHeader(http.StatusNoContent)
			return
		}

		// Send connection success event
		_, _ = io.WriteString(w, "event: PB_CONNECT\ndata: {\"clientId\":\"test-client-id\"}\n\n")
		flusher.Flush()

		// Maintain connection until context is cancelled
		<-r.Context().Done()
	}))
	defer srv.Close()

	c := NewClient(srv.URL)
	var wg sync.WaitGroup
	concurrentSubs := 20 // Number of concurrent subscribe/unsubscribe operations

	wg.Add(concurrentSubs)
	for i := 0; i < concurrentSubs; i++ {
		go func() {
			defer wg.Done()
			ctx, cancel := context.WithTimeout(context.Background(), 100*time.Millisecond)
			defer cancel()

			// Attempt to subscribe and immediately unsubscribe if successful
			unsubscribe, err := c.Realtime.Subscribe(ctx, []string{"test"}, func(e *RealtimeEvent, err error) {
				// The callback itself is not important
			})
			if err == nil && unsubscribe != nil {
				unsubscribe()
			}
		}()
	}

	wg.Wait()
}
