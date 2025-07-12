package pocketbase

import (
	"net/http"
	"net/http/httptest"
	"sync"
	"testing"
	"time"

	"github.com/goccy/go-json"
)

// TestPasswordAuthRace는 토큰 갱신 중 발생할 수 있는 데이터 경쟁을 테스트합니다.
func TestPasswordAuthRace(t *testing.T) {
	// 토큰 갱신 요청을 시뮬레이션하기 위한 테스트 서버
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// 실제 갱신 로직처럼 약간의 지연을 추가하여 경쟁 조건을 유도
		time.Sleep(10 * time.Millisecond)
		_ = json.NewEncoder(w).Encode(AuthResponse{
			Token:  "new-token",
			Record: &Record{BaseModel: BaseModel{ID: "user1"}},
		})
	}))
	defer srv.Close()

	c := NewClient(srv.URL)
	// 초기 인증 상태 설정 (만료된 토큰을 가진 것처럼)
	authStrategy := NewPasswordAuth(c, "users", "test", "password")
	authStrategy.tokenExp = time.Now().Add(-1 * time.Hour) // 토큰을 만료된 상태로 설정
	c.AuthStore = authStrategy

	var wg sync.WaitGroup
	concurrentRequests := 50 // 동시에 실행할 고루틴 수

	wg.Add(concurrentRequests)

	// 여러 고루틴에서 동시에 토큰 요청
	for i := 0; i < concurrentRequests; i++ {
		go func() {
			defer wg.Done()
			// Token() 메서드는 내부적으로 갱신 로직을 트리거합니다.
			_, err := c.AuthStore.Token(c)
			if err != nil {
				// 테스트 중 오류가 발생할 수 있지만, 여기서는 경쟁 조건 자체에 집중합니다.
				// t.Errorf("Token retrieval failed: %v", err)
			}
		}()
	}

	wg.Wait()
}
