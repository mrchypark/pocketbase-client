package pocketbase

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"net/url"
	"sync"

	"github.com/goccy/go-json"
	"github.com/tmaxmax/go-sse"
)

// RealtimeServiceAPI defines the real-time subscription functionality.
type RealtimeServiceAPI interface {
	Subscribe(ctx context.Context, topics []string, callback RealtimeCallback) (UnsubscribeFunc, error)
}

// RealtimeCallback is the type for callback functions that handle real-time events.
type RealtimeCallback func(event *RealtimeEvent, err error)

// UnsubscribeFunc is a function that unsubscribes from a real-time topic.
type UnsubscribeFunc func()

// RealtimeService handles the real-time subscription API.
type RealtimeService struct {
	Client *Client
}

var _ RealtimeServiceAPI = (*RealtimeService)(nil)

// Subscribe subscribes to specific topics and executes a callback when an event occurs.
// This implementation ensures that the subscription is confirmed before returning.
func (s *RealtimeService) Subscribe(ctx context.Context, topics []string, callback RealtimeCallback) (UnsubscribeFunc, error) {
	path := "/api/realtime"
	endpoint, err := url.JoinPath(s.Client.BaseURL, path)
	if err != nil {
		return nil, fmt.Errorf("pocketbase: invalid realtime path: %w", err)
	}

	subCtx, cancel := context.WithCancel(ctx)
	req, err := http.NewRequestWithContext(subCtx, http.MethodGet, endpoint, nil)
	if err != nil {
		cancel()
		return nil, fmt.Errorf("pocketbase: failed to create sse request: %w", err)
	}

	sseHTTPClient := *s.Client.HTTPClient
	sseHTTPClient.Timeout = 0 // Disable timeout for streaming

	sseClient := sse.Client{HTTPClient: &sseHTTPClient}
	conn := sseClient.NewConnection(req)

	var wg sync.WaitGroup
	wg.Add(1) // Wait for the initial connection and subscription

	connectErrChan := make(chan error, 1)

	// Register event handler
	conn.SubscribeToAll(func(event sse.Event) {
		// --- Initial Connection Handling ---
		if event.Type == "PB_CONNECT" {
			var connectEvent struct {
				ClientID string `json:"clientId"`
			}
			if err := json.Unmarshal([]byte(event.Data), &connectEvent); err != nil {
				connectErrChan <- fmt.Errorf("pocketbase: failed to unmarshal PB_CONNECT event: %w", err)
				wg.Done()
				return
			}
			if connectEvent.ClientID == "" {
				connectErrChan <- fmt.Errorf("pocketbase: PB_CONNECT event missing clientId")
				wg.Done()
				return
			}

			// Send subscription request using the main client's send method
			body := map[string]interface{}{"clientId": connectEvent.ClientID, "subscriptions": topics}
			if err := s.Client.send(subCtx, http.MethodPost, path, body, nil); err != nil {
				connectErrChan <- fmt.Errorf("pocketbase: failed to send subscription request: %w", err)
			} else {
				connectErrChan <- nil // Success
			}
			wg.Done() // Signal that the subscription attempt is complete
			return
		}

		// --- Regular Event Handling ---
		if len(event.Data) == 0 { // Ignore empty data (e.g., keep-alive messages)
			return
		}

		var rtEvent RealtimeEvent
		if err := json.Unmarshal([]byte(event.Data), &rtEvent); err != nil {
			callback(nil, fmt.Errorf("pocketbase: failed to unmarshal realtime event: %w. Raw data: %s", err, string(event.Data)))
			return
		}
		callback(&rtEvent, nil)
	})

	// Start connection in a separate goroutine.
	go func() {
		// Connect() blocks until the connection is closed.
		if err := conn.Connect(); err != nil && !errors.Is(err, context.Canceled) {
			callback(nil, fmt.Errorf("pocketbase: sse subscription failed: %w", err))
		}
	}()

	// Wait for the subscription to be confirmed or fail
	wg.Wait()
	close(connectErrChan)
	if err := <-connectErrChan; err != nil {
		cancel() // Clean up context on failure
		return nil, err
	}

	// Unsubscribe function to be returned to the caller
	unsubscribe := func() {
		cancel()
	}

	return unsubscribe, nil
}
