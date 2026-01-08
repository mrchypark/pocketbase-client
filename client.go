package pocketbase

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
	"sync"

	"github.com/goccy/go-json"
)

// Client interacts with the PocketBase API.
type Client struct {
	mu sync.RWMutex

	BaseURL    string       // Base URL of the PocketBase server
	HTTPClient *http.Client // HTTP client used for requests

	AuthStore   AuthStrategy
	Collections CollectionServiceAPI // Service for managing collections
	Records     RecordServiceAPI     // Service for record CRUD operations
	Realtime    RealtimeServiceAPI   // Service for real-time subscriptions
	Admins      AdminServiceAPI      // Service for managing administrators
	Users       UserServiceAPI       // Service for general user-related operations
	Logs        LogServiceAPI        // Service for viewing logs
	Settings    SettingServiceAPI    // Service for viewing and updating settings
	Batch       BatchServiceAPI      // General batch service
	Legacy      LegacyServiceAPI     // Legacy API service
	Files       FileServiceAPI       // Service for file operations
}

type authInjector struct {
	client *Client
	next   http.RoundTripper
}

func (t *authInjector) RoundTrip(req *http.Request) (*http.Response, error) {
	// Avoid injecting a (possibly stale) token into auth bootstrap endpoints.
	// Ex: /auth-with-password, /auth-with-oauth2, /auth-with-otp.
	if strings.Contains(req.URL.Path, "/auth-with-") {
		return t.next.RoundTrip(req)
	}
	t.client.mu.RLock()
	authStore := t.client.AuthStore
	t.client.mu.RUnlock()

	if authStore != nil {
		var tok string
		var err error
		if withCtx, ok := authStore.(AuthStrategyWithContext); ok {
			tok, err = withCtx.TokenWithContext(req.Context(), t.client)
		} else {
			tok, err = authStore.Token(t.client)
		}
		if err != nil {
			return nil, err
		}
		if tok != "" {
			req.Header.Set("Authorization", tok)
		}
	}
	return t.next.RoundTrip(req)
}

// NewClient creates a new Client with the given baseURL.
func NewClient(baseURL string, opts ...ClientOption) *Client {
	c := &Client{
		BaseURL:    baseURL,
		HTTPClient: &http.Client{},
	}
	c.AuthStore = &NilAuth{}
	for _, opt := range opts {
		opt(c)
	}

	transport := c.HTTPClient.Transport
	if transport == nil {
		transport = http.DefaultTransport
	}
	c.HTTPClient.Transport = &authInjector{client: c, next: transport}
	c.Collections = &CollectionService{Client: c}
	c.Records = &RecordService{Client: c}
	c.Realtime = &RealtimeService{Client: c}
	c.Admins = &AdminService{Client: c}
	c.Users = &UserService{Client: c}
	c.Logs = &LogService{Client: c}
	c.Settings = &SettingService{Client: c}
	c.Batch = &BatchService{client: c}
	c.Legacy = &LegacyService{Client: c}
	c.Files = &FileService{Client: c}
	return c
}

// ClearAuthStore removes the stored authentication information.
func (c *Client) ClearAuthStore() {
	c.mu.Lock()
	defer c.mu.Unlock()

	if c.AuthStore != nil {
		c.AuthStore.Clear()
		c.AuthStore = &NilAuth{}
	}
}

// newRequest performs common request initialization.
func (c *Client) newRequest(ctx context.Context, method, path string, body io.Reader, contentType string) (*http.Request, error) {
	base, err := url.Parse(c.BaseURL)
	if err != nil {
		return nil, fmt.Errorf("pocketbase: invalid base URL: %w", err)
	}
	rel, err := url.Parse(path)
	if err != nil {
		return nil, fmt.Errorf("pocketbase: invalid path: %w", err)
	}
	endpoint := base.ResolveReference(rel).String()

	req, err := http.NewRequestWithContext(ctx, method, endpoint, body)
	if err != nil {
		return nil, fmt.Errorf("pocketbase: failed to create request: %w", err)
	}

	if contentType != "" {
		req.Header.Set("Content-Type", contentType)
	}

	return req, nil
}

// Send is a wrapper used by external service implementations.
func (c *Client) Send(ctx context.Context, method, path string, body, responseData any) error {
	return c.send(ctx, method, path, body, responseData)
}

// SendWithOptions sends a request with additional options.
func (c *Client) SendWithOptions(ctx context.Context, method, path string, body, responseData any, opts ...RequestOption) error {
	return c.send(ctx, method, path, body, responseData, opts...)
}

func (c *Client) sendStream(ctx context.Context, method, path string, body io.Reader, contentType string) (io.ReadCloser, error) {
	req, err := c.newRequest(ctx, method, path, body, contentType)
	if err != nil {
		return nil, err
	}

	res, err := c.HTTPClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("pocketbase: http request failed: %w", err)
	}

	if res.StatusCode >= http.StatusBadRequest {
		defer res.Body.Close()
		resBody, err := io.ReadAll(res.Body)
		if err != nil {
			return nil, fmt.Errorf("pocketbase: failed to read response body: %w", err)
		}
		return nil, ParseAPIErrorFromResponse(res, resBody)
	}

	return res.Body, nil
}

// send is the central handler for all API requests.
func (c *Client) send(ctx context.Context, method, path string, body, responseData any, opts ...RequestOption) error {
	var reqBody io.Reader
	if body != nil {
		b, err := json.Marshal(body)
		if err != nil {
			return fmt.Errorf("pocketbase: failed to marshal request body: %w", err)
		}
		reqBody = bytes.NewReader(b)
	}
	return c.do(ctx, method, path, reqBody, "application/json", responseData, opts...)
}

func (c *Client) do(ctx context.Context, method, path string, body io.Reader, contentType string, responseData any, opts ...RequestOption) error {
	var ropts requestOptions
	for _, opt := range opts {
		opt(&ropts)
	}
	if ropts.writer != nil && responseData != nil {
		return fmt.Errorf("pocketbase: WithResponseWriter and responseData cannot be used together")
	}

	req, err := c.newRequest(ctx, method, path, body, contentType)
	if err != nil {
		return err
	}

	res, err := c.HTTPClient.Do(req)
	if err != nil {
		return fmt.Errorf("pocketbase: http request failed: %w", err)
	}
	defer res.Body.Close()

	if res.StatusCode >= http.StatusBadRequest {
		resBody, err := io.ReadAll(res.Body)
		if err != nil {
			return fmt.Errorf("pocketbase: failed to read error response body: %w", err)
		}

		return ParseAPIErrorFromResponse(res, resBody)
	}

	if ropts.writer != nil {
		if _, err := copyWithFlush(ropts.writer, res.Body); err != nil {
			return fmt.Errorf("pocketbase: failed to stream response: %w", err)
		}
		return nil
	}

	resBody, err := io.ReadAll(res.Body)
	if err != nil {
		return fmt.Errorf("pocketbase: failed to read response body: %w", err)
	}

	if responseData != nil {
		if err := json.Unmarshal(resBody, responseData); err != nil {
			return fmt.Errorf("pocketbase: failed to unmarshal response: %w", err)
		}
	}

	return nil
}

// WithPassword creates a PasswordAuth strategy and sets it to the client.
func (c *Client) WithPassword(ctx context.Context, collection, identity, password string) (*AuthResponse, error) {
	authStrategy := NewPasswordAuth(c, collection, identity, password)
	c.mu.Lock()
	c.AuthStore = authStrategy
	c.mu.Unlock()

	// Get the first authentication token immediately.
	token, err := authStrategy.TokenWithContext(ctx, c)
	if err != nil {
		c.ClearAuthStore()
		return nil, err
	}

	currentAuth := authStrategy.auth.Load()
	if currentAuth == nil {
		return nil, fmt.Errorf("authentication succeeded but no auth data is available")
	}

	res := &AuthResponse{Token: token}
	if admin, ok := currentAuth.model.(*Admin); ok {
		res.Admin = admin
	}
	if record, ok := currentAuth.model.(*Record); ok {
		res.Record = record
	}

	return res, nil
}

// WithAdminPassword is a convenience method for WithPassword.
func (c *Client) WithAdminPassword(ctx context.Context, identity, password string) (*AuthResponse, error) {
	return c.WithPassword(ctx, "_superusers", identity, password)
}

// WithToken sets a TokenAuth strategy to the client.
func (c *Client) WithToken(token string) {
	c.mu.Lock()
	defer c.mu.Unlock()

	c.AuthStore = NewTokenAuth(token)
}

// WithAuthStrategy sets a custom auth strategy to the client.
// If nil is provided, the client falls back to an unauthenticated strategy.
//
// This is useful when you don't want the built-in PasswordAuth to keep credentials
// in memory, or when you have a custom token refresh flow.
func (c *Client) WithAuthStrategy(strategy AuthStrategy) {
	c.mu.Lock()
	defer c.mu.Unlock()

	if c.AuthStore != nil {
		c.AuthStore.Clear()
	}
	if strategy == nil {
		c.AuthStore = &NilAuth{}
		return
	}
	c.AuthStore = strategy
}

// UseAuthResponse receives an AuthResponse and sets the client authentication state.
func (c *Client) UseAuthResponse(res *AuthResponse) *Client {
	c.mu.Lock()
	defer c.mu.Unlock()

	if res == nil || res.Token == "" {
		if c.AuthStore != nil {
			c.AuthStore.Clear()
		}
		c.AuthStore = &NilAuth{}
		return c
	}

	c.AuthStore = NewTokenAuth(res.Token)

	return c
}

// HealthCheck checks the health status of the PocketBase server.
func (c *Client) HealthCheck(ctx context.Context) (map[string]any, error) {
	path := "/api/health"
	var result map[string]any
	if err := c.send(ctx, http.MethodGet, path, nil, &result); err != nil {
		return nil, err
	}
	return result, nil
}

func copyWithFlush(dst io.Writer, src io.Reader) (int64, error) {
	const copyWithFlushBufferSize = 32 * 1024
	var total int64
	buf := make([]byte, copyWithFlushBufferSize)
	for {
		n, err := src.Read(buf)
		if n > 0 {
			wn, werr := dst.Write(buf[:n])
			total += int64(wn)
			if werr != nil {
				return total, werr
			}
			if f, ok := dst.(interface{ Flush() }); ok {
				f.Flush()
			}
		}
		if err != nil {
			if err == io.EOF {
				break
			}
			return total, err
		}
	}
	return total, nil
}
