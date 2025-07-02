// Package httputil contains helper functions for sending HTTP requests with JSON bodies.
package httputil

import (
	"bytes"
	"fmt"
	"io"
	"net/http"

	"github.com/goccy/go-json"
)

// Do sends req using client and returns the response body.
// Responses with status code >= 400 are returned as errors including the body
// text to make debugging easier.
func Do(client *http.Client, req *http.Request) ([]byte, error) {
	resp, err := client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("httputil: send request: %w", err)
	}
	defer resp.Body.Close()
	data, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("httputil: read response body: %w", err)
	}
	if resp.StatusCode >= http.StatusBadRequest {
		return nil, fmt.Errorf("http error %d: %s", resp.StatusCode, string(data))
	}
	return data, nil
}

// JSONRequest creates a new http.Request with the given method and URL.
// When body is non-nil it is marshaled to JSON and the Content-Type header
// is set accordingly.
func JSONRequest(method, url string, body any) (*http.Request, error) {
	var reader io.Reader
	if body != nil {
		data, err := json.Marshal(body)
		if err != nil {
			return nil, fmt.Errorf("httputil: marshal body: %w", err)
		}
		reader = bytes.NewBuffer(data)
	}
	req, err := http.NewRequest(method, url, reader)
	if err != nil {
		return nil, fmt.Errorf("httputil: create request: %w", err)
	}
	if body != nil {
		req.Header.Set("Content-Type", "application/json")
	}
	return req, nil
}

// RawRequest creates a new http.Request using the provided reader as the body.
// The Content-Type header is set when contentType is not empty and body is not nil.
func RawRequest(method, url string, body io.Reader, contentType string) (*http.Request, error) {
	var reader io.Reader
	if body != nil {
		data, err := io.ReadAll(body)
		if err != nil {
			return nil, fmt.Errorf("httputil: read body: %w", err)
		}
		reader = bytes.NewReader(data)
	}
	req, err := http.NewRequest(method, url, reader)
	if err != nil {
		return nil, fmt.Errorf("httputil: create request: %w", err)
	}
	if contentType != "" && reader != nil {
		req.Header.Set("Content-Type", contentType)
	}
	return req, nil
}
