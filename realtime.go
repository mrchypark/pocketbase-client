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

	// sse.Client는 http.Client의 포인터를 받으므로, 복사하여 새로운 인스턴스를 만듭니다.
	sseHTTPClient := *s.Client.HTTPClient
	// 스트리밍 연결이 타임아웃으로 끊기지 않도록 설정합니다.
	sseHTTPClient.Timeout = 0

	sseClient := sse.Client{HTTPClient: &sseHTTPClient}
	conn := sseClient.NewConnection(req)

	var wg sync.WaitGroup
	wg.Add(1)

	// 이벤트 핸들러 등록
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

			// 이벤트 루프를 막지 않도록 별도 고루틴에서 구독 요청 전송
			go func() {

				if err := s.sendSubscriptionRequest(subCtx, path, connectEvent.ClientID, topics); err != nil {

					callback(nil, fmt.Errorf("pocketbase: failed to send subscription request: %w", err))
					cancel()
				} else {

				}
			}()
			return
		}

		if len(event.Data) == 0 { // Keep-alive 등 빈 데이터 무시
			return
		}

		var rtEvent RealtimeEvent
		if err := json.Unmarshal([]byte(event.Data), &rtEvent); err != nil {
			callback(nil, fmt.Errorf("pocketbase: failed to unmarshal realtime event: %w. Raw data: %s", err, string(event.Data)))
			return
		}
		callback(&rtEvent, nil)
	})

	// 별도 고루틴에서 연결 시작
	go func() {
		defer wg.Done()
		// Connect()는 연결이 끊길 때까지 블로킹됩니다.
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

// sendSubscriptionRequest는 독립된 http.Client를 사용하여 구독 정보를 전송합니다.
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
	if token := s.Client.AuthStore.token; token != "" {
		req.Header.Set("Authorization", token)
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
