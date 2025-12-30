package pocketbase

import (
	"io"
	"net/http"
)

// ClientOption configures a Client instance.
// ClientOption configures a Client instance.
type ClientOption func(*Client)

// WithHTTPClient sets the HTTP client used for requests.
func WithHTTPClient(hc *http.Client) ClientOption {
	return func(c *Client) {
		c.HTTPClient = hc
	}
}

// WithAuthStrategy sets the auth strategy used for requests.
// If nil is provided, the client falls back to an unauthenticated strategy.
func WithAuthStrategy(strategy AuthStrategy) ClientOption {
	return func(c *Client) {
		if strategy == nil {
			c.AuthStore = &NilAuth{}
			return
		}
		c.AuthStore = strategy
	}
}

type requestOptions struct {
	writer io.Writer
}

// RequestOption configures the behavior of a single request.
type RequestOption func(*requestOptions)

// WithResponseWriter streams the response body to the provided writer.
// If the writer implements http.Flusher, Flush is called after each write.
// This option cannot be used together with the responseData argument
// of Client.send or SendWithOptions; doing so results in an error.
func WithResponseWriter(w io.Writer) RequestOption {
	return func(o *requestOptions) {
		o.writer = w
	}
}
