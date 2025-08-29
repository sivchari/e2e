package e2e

import (
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
)

// TestNew verifies the test suite initialization.
func TestNew(t *testing.T) {
	config := Config{
		BaseURL: "http://example.com",
		Timeout: 10 * time.Second,
	}

	suite := New(t, config)

	if suite.config.BaseURL != config.BaseURL {
		t.Errorf("Expected BaseURL %s, got %s", config.BaseURL, suite.config.BaseURL)
	}

	if suite.config.Timeout != config.Timeout {
		t.Errorf("Expected Timeout %v, got %v", config.Timeout, suite.config.Timeout)
	}

	if suite.t != t {
		t.Error("Testing instance not properly set")
	}

	if suite.client == nil {
		t.Error("HTTP client not initialized")
	}
}

// TestDefaultTimeout verifies default timeout is set when not specified.
func TestDefaultTimeout(t *testing.T) {
	config := Config{
		BaseURL: "http://example.com",
		// Timeout not specified
	}

	suite := New(t, config)

	expectedTimeout := 30 * time.Second
	if suite.config.Timeout != expectedTimeout {
		t.Errorf("Expected default timeout %v, got %v", expectedTimeout, suite.config.Timeout)
	}
}

// TestBasicIntegration runs a simple integration test.
func TestBasicIntegration(t *testing.T) {
	// Create a test server
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte("OK"))
	}))
	defer server.Close()

	// Create client and make request
	client := New(t, Config{BaseURL: server.URL})
	client.GET("/test").Execute(t.Context()).ExpectStatus(200)
}
