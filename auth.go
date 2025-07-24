package pocketbase

import (
	"context"
	"fmt"
	"net/http"
	"net/url"
	"sync/atomic"
	"time"

	"github.com/golang-jwt/jwt/v5" // Library for JWT parsing
	"golang.org/x/sync/singleflight"
)

// AuthStrategy is an interface for various authentication strategies.
type AuthStrategy interface {
	// Token returns the currently valid token. Performs token refresh internally if needed.
	Token(client *Client) (string, error)
	// Clear initializes the authentication state.
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
	model    any
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

	// Return immediately if token is valid (no lock)
	if currentAuth != nil && time.Now().Before(currentAuth.tokenExp) {
		return currentAuth.token, nil
	}

	// If token is missing or expired, execute refresh only once with singleflight
	_, err, _ := a.refreshSingle.Do("refresh", func() (any, error) {
		return nil, a.refreshToken(client)
	})
	if err != nil {
		return "", err
	}

	// Reload with refreshed information
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

	// --- ✨ Modified part: JWT parsing logic ---
	var expiry time.Time
	// Parse token using MapClaims that includes registered claims (RegisteredClaims).
	// Here we don't verify signature, only extract expiration time information.
	token, _, err := new(jwt.Parser).ParseUnverified(authResponse.Token, jwt.MapClaims{})
	if err == nil {
		// Get the 'exp' claim.
		exp, err := token.Claims.GetExpirationTime()
		if err == nil && exp != nil {
			expiry = exp.Time
		}
	}

	// If parsing fails or there's no expiration time, set a short expiration time for safety.
	if expiry.IsZero() {
		// Example: Set to expire after 1 minute to trigger refresh on next request
		expiry = time.Now().Add(1 * time.Minute)
	}
	// --- ✨ End of modification ---

	newAuth := &authToken{
		token:    authResponse.Token,
		tokenExp: expiry, // Set to parsed expiration time
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
