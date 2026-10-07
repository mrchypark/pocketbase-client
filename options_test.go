package pocketbase

import (
	"bytes"
	"net/http"
	"testing"
)

func TestWithHTTPClient(t *testing.T) {
	c := &Client{}
	hc := &http.Client{}

	WithHTTPClient(hc)(c)

	if c.HTTPClient == hc || c.HTTPClient.Transport != hc.Transport || c.HTTPClient.Timeout != hc.Timeout || c.HTTPClient.Jar != hc.Jar {
		t.Errorf("WithHTTPClient must copy the HTTP client settings")
	}
}

func TestWithResponseWriter(t *testing.T) {
	ro := &requestOptions{}
	w := &bytes.Buffer{}

	WithResponseWriter(w)(ro)

	if ro.writer != w {
		t.Errorf("WithResponseWriter did not set the writer correctly")
	}
}
