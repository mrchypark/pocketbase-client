package pocketbase

import (
	"context"
	"fmt"
	"net/http"
	"net/url"
	"sync/atomic"
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
	client     *Client
	collection string
	identity   string
	password   string
	auth       atomic.Pointer[authToken]

	refreshSingle singleflight.Group
}

type authToken struct {
	token    string
	model    interface{}
	tokenExp time.Time
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
	currentAuth := a.auth.Load()

	// 토큰이 유효하면 즉시 반환 (잠금 없음)
	if currentAuth != nil && time.Now().Before(currentAuth.tokenExp) {
		return currentAuth.token, nil
	}

	// 토큰이 없거나 만료된 경우, singleflight로 한 번만 갱신 실행
	_, err, _ := a.refreshSingle.Do("refresh", func() (interface{}, error) {
		return nil, a.refreshToken(client)
	})
	if err != nil {
		return "", err
	}

	// 갱신된 정보로 다시 로드
	refreshedAuth := a.auth.Load()
	if refreshedAuth == nil {
		return "", fmt.Errorf("authentication failed: token not available after refresh")
	}

	return refreshedAuth.token, nil
}

func (a *PasswordAuth) refreshToken(client *Client) error {
	// 네트워크 요청은 잠금 없이 수행

	path := fmt.Sprintf("/api/collections/%s/auth-with-password", url.PathEscape(a.collection))
	body := map[string]string{"identity": a.identity, "password": a.password}

	var authResponse AuthResponse
	if err := client.send(context.Background(), http.MethodPost, path, body, &authResponse); err != nil {
		return err
	}

	// 갱신된 정보를 담을 새로운 authToken 생성
	newAuth := &authToken{
		token: authResponse.Token,
		// 토큰 만료 시간을 넉넉하게 설정 (실제로는 JWT 파싱해서 설정하는 것이 더 정확)
		tokenExp: time.Now().Add(59 * time.Minute),
	}
	if authResponse.Admin != nil {
		newAuth.model = authResponse.Admin
	} else {
		newAuth.model = authResponse.Record
	}

	// 새로운 인증 정보를 원자적으로 저장
	a.auth.Store(newAuth)

	return nil
}

func (a *PasswordAuth) Clear() {
	a.auth.Store(nil)
}
