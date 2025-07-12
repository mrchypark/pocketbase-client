package pocketbase

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"net/http"
	"net/url"
	"sync"
	"time"

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

	// Create a new HTTP client instance by copying the existing one,
	// as sse.Client expects a pointer to http.Client.
	sseHTTPClient := *s.Client.HTTPClient
	// Set the timeout to 0 to prevent the streaming connection from timing out.
	sseHTTPClient.Timeout = 0

	sseClient := sse.Client{HTTPClient: &sseHTTPClient}
	conn := sseClient.NewConnection(req)

	var wg sync.WaitGroup
	wg.Add(1)

	// Register event handler
	conn.SubscribeToAll(func(event sse.Event) {
		if event.Type == "PB_CONNECT" {
			var connectEvent struct {
				ClientID string `json:"clientId"`
			}
			if err := json.Unmarshal([]byte(event.Data), &connectEvent); err != nil {
				callback(nil, fmt.Errorf("pocketbase: failed to unmarshal PB_CONNECT event: %w", err))
				cancel()
				return
			}

			if connectEvent.ClientID == "" {
				callback(nil, fmt.Errorf("pocketbase: PB_CONNECT event missing clientId"))
				cancel()
				return
			}

			// Send subscription request in a separate goroutine to avoid blocking the event loop.
			go func() {

				if err := s.sendSubscriptionRequest(context.Background(), path, connectEvent.ClientID, topics); err != nil {

					callback(nil, fmt.Errorf("pocketbase: failed to send subscription request: %w", err))
					cancel()
				} else {

				}
			}()
			return
		}

		if len(event.Data) == 0 { // Ignore empty data, e.g., keep-alive messages
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
		defer wg.Done()
		// Connect() blocks until the connection is closed.
		if err := conn.Connect(); err != nil && !errors.Is(err, context.Canceled) {
			callback(nil, fmt.Errorf("pocketbase: sse subscription failed: %w", err))
		}
	}()

	unsubscribe := func() {
		cancel()
		wg.Wait()
	}

	return unsubscribe, nil
}

// sendSubscriptionRequest sends subscription information using an independent http.Client.
func (s *RealtimeService) sendSubscriptionRequest(ctx context.Context, path, clientID string, topics []string) error {
	bodyMap := map[string]interface{}{
		"clientId":      clientID,
		"subscriptions": topics,
	}
	body, err := json.Marshal(bodyMap)
	if err != nil {
		return fmt.Errorf("failed to marshal subscription body: %w", err)
	}

	endpoint, err := url.JoinPath(s.Client.BaseURL, path)
	if err != nil {
		return fmt.Errorf("invalid subscription path: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, endpoint, bytes.NewReader(body))
	if err != nil {
		return fmt.Errorf("failed to create subscription request: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")
	if s.Client.AuthStore != nil {
		token, err := s.Client.AuthStore.Token(s.Client)
		if err != nil {
			// 토큰 획득/갱신 실패 시 구독 요청도 실패 처리하는 것이 안전합니다.
			return fmt.Errorf("failed to get auth token for subscription: %w", err)
		}
		if token != "" {
			req.Header.Set("Authorization", token)
		}
	}

	// 이 요청만을 위한 일회용, 독립적인 HTTP 클라이언트를 생성합니다.
	httpClient := &http.Client{Timeout: 10 * time.Second}
	resp, err := httpClient.Do(req)
	if err != nil {
		return fmt.Errorf("failed to send subscription request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode >= http.StatusBadRequest {
		return fmt.Errorf("subscription request failed with status %d", resp.StatusCode)
	}

	return nil
}
