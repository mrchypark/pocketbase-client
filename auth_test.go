package pocketbase

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

func TestNilAuth(t *testing.T) {
	auth := &NilAuth{}

	token, err := auth.Token(nil)
	if err != nil {
		t.Fatalf("NilAuth.Token() should not return an error, but got: %v", err)
	}
	if token != "" {
		t.Fatalf("NilAuth.Token() should return an empty string, but got: %s", token)
	}

	// Clear should be a no-op and not panic
	auth.Clear()
}

func TestTokenAuth(t *testing.T) {
	const testToken = "test-token"
	auth := NewTokenAuth(testToken)

	token, err := auth.Token(nil)
	if err != nil {
		t.Fatalf("TokenAuth.Token() should not return an error, but got: %v", err)
	}
	if token != testToken {
		t.Fatalf("Expected token '%s', but got: %s", testToken, token)
	}
}

func TestTokenAuth_Clear(t *testing.T) {
	auth := NewTokenAuth("test-token")
	if auth.token != "test-token" {
		t.Fatalf("Expected initial token to be 'test-token', got '%s'", auth.token)
	}

	auth.Clear()

	if auth.token != "" {
		t.Errorf("Expected token to be empty after Clear, got '%s'", auth.token)
	}
}

func TestPasswordAuth_Clear(t *testing.T) {
	// Create a valid JWT token that expires in 1 hour
	expirationTime := time.Now().Add(1 * time.Hour)
	claims := &jwt.RegisteredClaims{
		ExpiresAt: jwt.NewNumericDate(expirationTime),
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	tokenString, err := token.SignedString([]byte("secret"))
	if err != nil {
		t.Fatalf("Failed to create test token: %v", err)
	}

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/api/collections/users/auth-with-password" {
			w.WriteHeader(http.StatusOK)
			fmt.Fprintf(w, `{"token":"%s","record":{"id":"recd1"}}`, tokenString)
		}
	}))
	defer server.Close()

	client := NewClient(server.URL)
	auth := NewPasswordAuth(client, "users", "testuser", "testpass")

	// Authenticate to populate the auth state
	_, err = auth.Token(client)
	if err != nil {
		t.Fatalf("Authentication failed: %v", err)
	}
	if auth.auth.Load() == nil {
		t.Fatal("Auth state should not be nil after successful authentication")
	}

	// Clear the auth state
	auth.Clear()

	// Verify that the auth state is now nil
	if auth.auth.Load() != nil {
		t.Fatal("Auth state should be nil after Clear is called")
	}
}

func TestPasswordAuth_Token_RefreshFailure(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/api/collections/users/auth-with-password" {
			w.WriteHeader(http.StatusInternalServerError)
			w.Write([]byte(`{"code":500,"message":"Internal server error"}`))
		}
	}))
	defer server.Close()

	client := NewClient(server.URL)
	auth := NewPasswordAuth(client, "users", "testuser", "testpass")

	// Initial token fetch should fail
	_, err := auth.Token(client)
	if err == nil {
		t.Fatal("Expected an error on token refresh failure, but got nil")
	}

	// Verify that the auth state is still nil
	if auth.auth.Load() != nil {
		t.Fatal("auth.auth should be nil after a failed refresh")
	}
}