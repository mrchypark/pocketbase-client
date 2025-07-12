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

// TestRealtimeSubscribeRace는 동시 구독/구독 취소 시의 데이터 경쟁을 테스트합니다.
func TestRealtimeSubscribeRace(t *testing.T) {
	// 실시간 연결을 처리하는 테스트 서버
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/event-stream")
		flusher, ok := w.(http.Flusher)
		if !ok {
			http.Error(w, "streaming unsupported", http.StatusInternalServerError)
			return
		}

		// 클라이언트가 구독 요청을 보낼 때까지 대기
		if r.Method == http.MethodPost {
			w.WriteHeader(http.StatusNoContent)
			return
		}

		// 연결 성공 이벤트 전송
		_, _ = io.WriteString(w, "event: PB_CONNECT\ndata: {\"clientId\":\"test-client-id\"}\n\n")
		flusher.Flush()

		// 컨텍스트가 취소될 때까지 연결 유지
		<-r.Context().Done()
	}))
	defer srv.Close()

	c := NewClient(srv.URL)
	var wg sync.WaitGroup
	concurrentSubs := 20 // 동시 구독/취소 수

	wg.Add(concurrentSubs)
	for i := 0; i < concurrentSubs; i++ {
		go func() {
			defer wg.Done()
			ctx, cancel := context.WithTimeout(context.Background(), 100*time.Millisecond)
			defer cancel()

			// 구독을 시도하고 성공하면 즉시 구독 취소
			unsubscribe, err := c.Realtime.Subscribe(ctx, []string{"test"}, func(e *RealtimeEvent, err error) {
				// 콜백 자체는 중요하지 않음
			})
			if err == nil && unsubscribe != nil {
				unsubscribe()
			}
		}()
	}

	wg.Wait()
}
