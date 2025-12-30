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

// AuthStrategyWithContext is an optional extension interface for AuthStrategy implementations
// that want to respect request cancellation and deadlines during token acquisition/refresh.
//
// If the configured strategy implements this interface, the client will prefer TokenWithContext
// over Token when injecting Authorization headers.
type AuthStrategyWithContext interface {
	AuthStrategy
	TokenWithContext(ctx context.Context, client *Client) (string, error)
}

type NilAuth struct{}

func (a *NilAuth) Token(client *Client) (string, error) { return "", nil }
func (a *NilAuth) Clear()                               {}

type TokenAuth struct {
	token atomic.Value // stores string
}

func NewTokenAuth(token string) *TokenAuth {
	a := &TokenAuth{}
	a.token.Store(token)
	return a
}

func (a *TokenAuth) Token(client *Client) (string, error) {
	if v := a.token.Load(); v != nil {
		return v.(string), nil
	}
	return "", nil
}

func (a *TokenAuth) TokenWithContext(ctx context.Context, client *Client) (string, error) {
	return a.Token(client)
}

func (a *TokenAuth) Clear() {
	a.token.Store("")
}

type PasswordAuth struct {
	client     *Client
	collection string
	identity   string
	password   string
	auth       atomic.Pointer[authToken]

	refreshSingle singleflight.Group
}

const tokenExpiryLeeway = 30 * time.Second

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
	return a.TokenWithContext(context.Background(), client)
}

func (a *PasswordAuth) TokenWithContext(ctx context.Context, client *Client) (string, error) {
	currentAuth := a.auth.Load()

	// Return immediately if token is valid (no lock)
	if currentAuth != nil && time.Now().Add(tokenExpiryLeeway).Before(currentAuth.tokenExp) {
		return currentAuth.token, nil
	}

	// If token is missing or expired, execute refresh only once with singleflight
	_, err, _ := a.refreshSingle.Do("refresh", func() (any, error) {
		return nil, a.refreshToken(ctx, client)
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

func (a *PasswordAuth) refreshToken(ctx context.Context, client *Client) error {
	path := fmt.Sprintf("/api/collections/%s/auth-with-password", url.PathEscape(a.collection))
	body := map[string]string{"identity": a.identity, "password": a.password}

	var authResponse AuthResponse
	if err := client.send(ctx, http.MethodPost, path, body, &authResponse); err != nil {
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
