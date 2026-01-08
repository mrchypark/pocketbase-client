package pocketbase

import (
	"net/http"
	"net/http/httptest"
	"sync"
	"testing"
	"time"

	"github.com/goccy/go-json"
)

// TestPasswordAuthRace tests for data races that can occur during token refresh.
func TestPasswordAuthRace(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		time.Sleep(10 * time.Millisecond)
		_ = json.NewEncoder(w).Encode(AuthResponse{
			Token:  "new-token",
			Record: &Record{ID: "user1"},
		})
	}))
	defer srv.Close()

	c := NewClient(srv.URL)
	// ✅ Modified part: Use the initial authentication state as is.
	// When Token() is called without a token, the refresh logic will be executed anyway.
	authStrategy := NewPasswordAuth(c, "users", "test", "password")
	c.AuthStore = authStrategy

	var wg sync.WaitGroup
	concurrentRequests := 50

	wg.Add(concurrentRequests)

	for i := 0; i < concurrentRequests; i++ {
		go func() {
			defer wg.Done()
			_, _ = c.AuthStore.Token(c) // Internally triggers refresh logic
		}()
	}

	wg.Wait()
}
