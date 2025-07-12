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
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		time.Sleep(10 * time.Millisecond)
		_ = json.NewEncoder(w).Encode(AuthResponse{
			Token:  "new-token",
			Record: &Record{BaseModel: BaseModel{ID: "user1"}},
		})
	}))
	defer srv.Close()

	c := NewClient(srv.URL)
	// ✅ 수정된 부분: 초기 인증 상태를 그대로 사용합니다.
	// 토큰이 없는 상태에서 Token()을 호출하면 어차피 갱신 로직이 실행됩니다.
	authStrategy := NewPasswordAuth(c, "users", "test", "password")
	c.AuthStore = authStrategy

	var wg sync.WaitGroup
	concurrentRequests := 50

	wg.Add(concurrentRequests)

	for i := 0; i < concurrentRequests; i++ {
		go func() {
			defer wg.Done()
			_, _ = c.AuthStore.Token(c) // 내부적으로 갱신 로직 트리거
		}()
	}

	wg.Wait()
}
