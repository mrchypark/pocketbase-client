package pocketbase

import (
	"context"
	"fmt"
	"net/http"
	"net/url"
	"sync"
	"time"

	"golang.org/x/sync/singleflight"
)

// AuthStore manages authentication state and token refreshing.
type AuthStore struct {
	client        *Client
	refreshSingle singleflight.Group
	mu            sync.RWMutex
	token         string
	model         interface{}
	tokenExp      time.Time
}

func newAuthStore(c *Client) *AuthStore {
	return &AuthStore{client: c}
}

// Token returns a valid auth token, refreshing it when expired.
func (a *AuthStore) Token() (string, error) {
	a.mu.RLock()
	token := a.token
	exp := a.tokenExp
	a.mu.RUnlock()

	if token == "" {
		return "", nil
	}

	if time.Now().Before(exp) {
		return token, nil
	}

	if _, err, _ := a.refreshSingle.Do("refresh", func() (interface{}, error) {
		return nil, a.refreshToken()
	}); err != nil {
		return "", err
	}

	a.mu.RLock()
	defer a.mu.RUnlock()
	return a.token, nil
}

// Set updates the store with a new token and model.
func (a *AuthStore) Set(token string, model interface{}) {
	a.mu.Lock()
	defer a.mu.Unlock()
	a.token = token
	a.model = model
	a.tokenExp = time.Now().Add(50 * time.Minute)
}

// Clear removes any stored authentication state.
func (a *AuthStore) Clear() {
	a.mu.Lock()
	defer a.mu.Unlock()
	a.token = ""
	a.model = nil
	a.tokenExp = time.Time{}
}

func (a *AuthStore) refreshToken() error {
	a.mu.RLock()
	m := a.model
	a.mu.RUnlock()

	if m == nil {
		return fmt.Errorf("pocketbase: no auth data to refresh")
	}

	switch v := m.(type) {
	case *Admin:
		var res AuthResponse
		if err := a.client.send(context.Background(), http.MethodPost, "/api/admins/auth-refresh", nil, &res); err != nil {
			return err
		}
		a.Set(res.Token, res.Admin)
	case *Record:
		collection := v.CollectionName
		if collection == "" {
			return fmt.Errorf("pocketbase: record collection name missing")
		}
		path := fmt.Sprintf("/api/collections/%s/auth-refresh", url.PathEscape(collection))
		var res AuthResponse
		if err := a.client.send(context.Background(), http.MethodPost, path, nil, &res); err != nil {
			return err
		}
		a.Set(res.Token, res.Record)
	default:
		return fmt.Errorf("pocketbase: unsupported model type %T", m)
	}
	return nil
}
