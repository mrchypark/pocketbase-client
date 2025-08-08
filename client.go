// Package pocketbase provides a Go client for interacting with the PocketBase backend.
//
// This client covers almost all API endpoints exposed by PocketBase, including authentication,
// record CRUD, real-time subscriptions, and file management. All network requests can be
// controlled via context.Context, allowing callers to easily set timeouts or cancel requests.
package pocketbase

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"

	"github.com/goccy/go-json"
)

// Client interacts with the PocketBase API.
//
// Client is the primary object for interacting with the PocketBase API.
//
// BaseURL represents the root URL of the PocketBase server.
// HTTPClient is used for all requests and can be customized as needed.
// Service fields like Records provide specific endpoint operations.
// Client is the primary object for interacting with the PocketBase API.
//
// BaseURL represents the root URL of the PocketBase server.
// HTTPClient is used for all requests and can be customized as needed.
// Service fields like Records provide specific endpoint operations.
type Client struct {
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
	if strings.Contains(req.URL.Path, "auth-with-password") {
		return t.next.RoundTrip(req)
	}
	if t.client.AuthStore != nil {
		tok, err := t.client.AuthStore.Token(t.client)
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

// ClearAuthStore removes the stored authentication information, effectively logging out.
func (c *Client) ClearAuthStore() {
	if c.AuthStore != nil {
		c.AuthStore.Clear()
		// Replace with NilAuth again to make it unauthenticated state.
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

// Send is a wrapper used by external service implementations like RecordService.
func (c *Client) Send(ctx context.Context, method, path string, body, responseData interface{}) error {
	return c.send(ctx, method, path, body, responseData)
}

// SendWithOptions sends a request with additional options.
func (c *Client) SendWithOptions(ctx context.Context, method, path string, body, responseData interface{}, opts ...RequestOption) error {
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
		return nil, ParseAPIError(res, resBody, path)
	}

	return res.Body, nil
}

// send is the central handler for all API requests.
func (c *Client) send(ctx context.Context, method, path string, body, responseData interface{}, opts ...RequestOption) error {
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

func (c *Client) do(ctx context.Context, method, path string, body io.Reader, contentType string, responseData interface{}, opts ...RequestOption) error {
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

		// Parse error using new error system
		return ParseAPIError(res, resBody, path)
	}

	// The success response handling logic
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
	// Set new PasswordAuth strategy
	authStrategy := NewPasswordAuth(c, collection, identity, password)
	c.AuthStore = authStrategy

	// Get the first authentication token immediately.
	token, err := authStrategy.Token(c)
	if err != nil {
		c.ClearAuthStore() // Clear auth info on failure
		return nil, err
	}

	// ✅ Modified part: Instead of getting model directly from authStrategy,
	// load the internal authToken pointer to get model information.
	currentAuth := authStrategy.auth.Load()
	if currentAuth == nil {
		// Abnormal situation where token was obtained but no auth data is available
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
	c.AuthStore = NewTokenAuth(token)
}

// UseAuthResponse receives an AuthResponse and sets the client authentication state.
// Use this method to configure the client when you already have a token from OAuth2 or token refresh.
func (c *Client) UseAuthResponse(res *AuthResponse) *Client {
	if res == nil || res.Token == "" {
		c.ClearAuthStore()
		return c
	}

	// Use TokenAuth strategy since the received token is static.
	c.AuthStore = NewTokenAuth(res.Token)

	return c
}

// HealthCheck checks the health status of the PocketBase server.
func (c *Client) HealthCheck(ctx context.Context) (map[string]interface{}, error) {
	path := "/api/health"
	var result map[string]interface{}
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
