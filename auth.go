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

// AuthStrategy는 다양한 인증 전략을 위한 인터페이스입니다.
type AuthStrategy interface {
	// Token은 현재 유효한 토큰을 반환합니다. 필요한 경우 내부적으로 토큰 갱신을 수행합니다.
	Token(client *Client) (string, error)
	// Clear는 인증 상태를 초기화합니다.
	Clear()
}

type NilAuth struct{}

func (a *NilAuth) Token(client *Client) (string, error) { return "", nil }
func (a *NilAuth) Clear()                               {}

type TokenAuth struct {
	token string
}

func NewTokenAuth(token string) *TokenAuth {
	return &TokenAuth{token: token}
}

func (a *TokenAuth) Token(client *Client) (string, error) {
	return a.token, nil
}

func (a *TokenAuth) Clear() {
	a.token = ""
}

type PasswordAuth struct {
	client        *Client
	collection    string
	identity      string
	password      string
	mu            sync.RWMutex
	token         string
	model         interface{}
	tokenExp      time.Time
	refreshSingle singleflight.Group
}

func NewPasswordAuth(client *Client, collection, identity, password string) *PasswordAuth {
	return &PasswordAuth{
		client:     client,
		collection: collection,
		identity:   identity,
		password:   password,
	}
}

func (a *PasswordAuth) Token(client *Client) (string, error) {
	a.mu.RLock()
	// 토큰이 없거나 만료되었다면 획득/갱신 로직 수행
	if a.token == "" || time.Now().After(a.tokenExp) {
		a.mu.RUnlock() // Lock을 풀고 갱신 로직 진입
		_, err, _ := a.refreshSingle.Do("refresh", func() (interface{}, error) {
			return nil, a.refreshToken(client)
		})
		if err != nil {
			return "", err
		}
	} else {
		a.mu.RUnlock()
	}

	a.mu.RLock()
	defer a.mu.RUnlock()
	return a.token, nil
}

func (a *PasswordAuth) refreshToken(client *Client) error {
	a.mu.Lock()
	defer a.mu.Unlock()

	var path string
	var body interface{}

	// 기존 토큰/모델이 있으면 refresh, 없으면 새로 인증
	if a.model != nil {
		// 토큰 갱신 로직
		switch m := a.model.(type) {
		case *Admin:
			// 관리자 갱신은 이 경로를 유지합니다. (LegacyService에서 사용)
			path = "/api/admins/auth-refresh"
		case *Record:
			path = fmt.Sprintf("/api/collections/%s/auth-refresh", url.PathEscape(m.CollectionName))
		default:
			// 비정상적인 경우, 다시 비밀번호로 인증합니다.
			path = fmt.Sprintf("/api/collections/%s/auth-with-password", url.PathEscape(a.collection))
			body = map[string]string{
				"identity": a.identity,
				"password": a.password,
			}
		}
	} else {
		// ✨ 수정된 부분: _superusers 특별 취급 로직 제거
		// 이제 항상 /api/collections/{collection}/auth-with-password 경로를 사용합니다.
		path = fmt.Sprintf("/api/collections/%s/auth-with-password", url.PathEscape(a.collection))
		body = map[string]string{
			"identity": a.identity,
			"password": a.password,
		}
	}

	var authResponse AuthResponse
	if err := client.send(context.Background(), http.MethodPost, path, body, &authResponse); err != nil {
		return err
	}

	a.token = authResponse.Token
	a.tokenExp = time.Now().Add(60 * time.Minute) // 갱신 주기 설정
	if authResponse.Admin != nil {
		a.model = authResponse.Admin
	} else {
		a.model = authResponse.Record
	}

	return nil
}

func (a *PasswordAuth) Clear() {
	a.mu.Lock()
	defer a.mu.Unlock()
	a.token = ""
	a.model = nil
	a.identity = ""
	a.password = ""
	a.tokenExp = time.Time{}
}
