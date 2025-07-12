package pocketbase

import (
	"context"
	"fmt"
	"net/http"
	"net/url"
	"sync/atomic"
	"time"

	"github.com/golang-jwt/jwt/v5" // JWT 파싱을 위한 라이브러리 추가
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
	path := fmt.Sprintf("/api/collections/%s/auth-with-password", url.PathEscape(a.collection))
	body := map[string]string{"identity": a.identity, "password": a.password}

	var authResponse AuthResponse
	if err := client.send(context.Background(), http.MethodPost, path, body, &authResponse); err != nil {
		return err
	}

	// --- ✨ 수정된 부분: JWT 파싱 로직 ---
	var expiry time.Time
	// 등록된 클레임(RegisteredClaims)을 포함하는 MapClaims를 사용하여 토큰을 파싱합니다.
	// 여기서는 서명 검증은 하지 않고 만료 시간 정보만 추출합니다.
	token, _, err := new(jwt.Parser).ParseUnverified(authResponse.Token, jwt.MapClaims{})
	if err == nil {
		// 'exp' 클레임을 가져옵니다.
		exp, err := token.Claims.GetExpirationTime()
		if err == nil && exp != nil {
			expiry = exp.Time
		}
	}

	// 파싱에 실패하거나 만료 시간이 없는 경우, 안전을 위해 짧은 만료 시간을 설정합니다.
	if expiry.IsZero() {
		// 예: 1분 후 만료로 처리하여 다음 요청 시 다시 갱신하도록 유도
		expiry = time.Now().Add(1 * time.Minute)
	}
	// --- ✨ 수정 끝 ---

	newAuth := &authToken{
		token:    authResponse.Token,
		tokenExp: expiry, // 파싱된 만료 시간으로 설정
	}
	if authResponse.Admin != nil {
		newAuth.model = authResponse.Admin
	} else {
		newAuth.model = authResponse.Record
	}

	a.auth.Store(newAuth)

	return nil
}

func (a *PasswordAuth) Clear() {
	a.auth.Store(nil)
}
