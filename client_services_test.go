package pocketbase

import "testing"

// TestNewClientServices tests that all client services are initialized.
func TestNewClientServices(t *testing.T) {
	c := NewClient("http://example.com")
	if c.Collections == nil {
		t.Fatal("Collections service not initialized")
	}
	if c.Admins == nil {
		t.Fatal("Admins service not initialized")
	}
	if c.Users == nil {
		t.Fatal("Users service not initialized")
	}
	if c.Logs == nil {
		t.Fatal("Logs service not initialized")
	}
	if c.Settings == nil {
		t.Fatal("Settings service not initialized")
	}
}
