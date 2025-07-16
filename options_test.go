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

	if c.HTTPClient != hc {
		t.Errorf("WithHTTPClient did not set the HTTP client correctly")
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
