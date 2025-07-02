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

	AuthStore   *AuthStore           // Authentication state manager
	Collections CollectionServiceAPI // Service for managing collections
	Records     RecordServiceAPI     // Service for record CRUD operations
	Realtime    RealtimeServiceAPI   // Service for real-time subscriptions
	Admins      AdminServiceAPI      // Service for managing administrators
	Users       UserServiceAPI       // Service for general user-related operations
	Logs        LogServiceAPI        // Service for viewing logs
	Settings    SettingServiceAPI    // Service for viewing and updating settings
	Batch       BatchServiceAPI      // General batch service
	Legacy      LegacyServiceAPI     // Legacy API service
}

type authInjector struct {
	client *Client
	next   http.RoundTripper
}

func (t *authInjector) RoundTrip(req *http.Request) (*http.Response, error) {
	if t.client.AuthStore != nil {
		tok, err := t.client.AuthStore.Token()
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
	c.AuthStore = newAuthStore(c)
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
	return c
}

// ClearAuthStore removes the stored authentication information, effectively logging out.
func (c *Client) ClearAuthStore() {
	if c.AuthStore != nil {
		c.AuthStore.Clear()
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
		apiErr := &APIError{}
		if err := json.Unmarshal(resBody, apiErr); err != nil {
			return nil, fmt.Errorf("pocketbase: http error %d: %s", res.StatusCode, string(resBody))
		}
		return nil, apiErr
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
			return fmt.Errorf("pocketbase: failed to read response body: %w", err)
		}
		apiErr := &APIError{}
		if err := json.Unmarshal(resBody, apiErr); err != nil {
			return fmt.Errorf("pocketbase: http error %d: %s", res.StatusCode, string(resBody))
		}
		return apiErr
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

// AuthenticateAsAdmin authenticates as an admin and stores the authentication information.
func (c *Client) AuthenticateAsAdmin(ctx context.Context, identity, password string) (*AuthResponse, error) {
	c.ClearAuthStore()

	return c.AuthenticateWithPassword(ctx, "_superusers", identity, password)

}

// AuthenticateWithPassword authenticates as a regular user and stores the authentication information.
func (c *Client) AuthenticateWithPassword(ctx context.Context, collection, identity, password string) (*AuthResponse, error) {
	c.ClearAuthStore()

	reqBody := map[string]string{
		"identity": identity,
		"password": password,
	}

	var authResponse AuthResponse
	path := fmt.Sprintf("/api/collections/%s/auth-with-password", url.PathEscape(collection))
	if err := c.send(ctx, http.MethodPost, path, reqBody, &authResponse); err != nil {
		return nil, err
	}

	c.AuthStore.Set(authResponse.Token, authResponse.Record)

	return &authResponse, nil
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
