// Package e2e provides a simple and intuitive HTTP testing framework for Go.
package e2e

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"testing"
)

// Config holds configuration for the test suite.
type Config struct {
	BaseURL string
}

// TestSuite represents the main test suite.
type TestSuite struct {
	config Config
	t      *testing.T
	client *http.Client
}

// HTTPBuilder builds HTTP requests.
type HTTPBuilder struct {
	suite  *TestSuite
	method string
	path   string
	body   interface{}
	resp   *http.Response
}

// New creates a new test suite.
func New(t *testing.T, config Config) *TestSuite {
	t.Helper()

	return &TestSuite{
		config: config,
		t:      t,
		client: &http.Client{},
	}
}

// GET creates a GET request builder.
func (s *TestSuite) GET(path string) *HTTPBuilder {
	return &HTTPBuilder{
		suite:  s,
		method: "GET",
		path:   path,
	}
}

// POST creates a POST request builder.
func (s *TestSuite) POST(path string) *HTTPBuilder {
	return &HTTPBuilder{
		suite:  s,
		method: "POST",
		path:   path,
	}
}

// Body sets the request body (accepts string, struct, or map).
func (h *HTTPBuilder) Body(body interface{}) *HTTPBuilder {
	h.body = body

	return h
}

// Execute performs the HTTP request.
func (h *HTTPBuilder) Execute(ctx context.Context) *HTTPBuilder {
	baseURL, err := url.Parse(h.suite.config.BaseURL)
	if err != nil {
		h.suite.t.Fatalf("Failed to parse base URL: %v", err)
	}

	path, err := url.Parse(h.path)
	if err != nil {
		h.suite.t.Fatalf("Failed to parse path: %v", err)
	}

	reqURL := baseURL.ResolveReference(path)

	var bodyReader io.Reader

	if h.body != nil {
		bodyBytes, err := h.serializeBody()
		if err != nil {
			h.suite.t.Fatalf("Failed to serialize body: %v", err)
		}

		bodyReader = bytes.NewBuffer(bodyBytes)
	}

	req, err := http.NewRequestWithContext(ctx, h.method, reqURL.String(), bodyReader)
	if err != nil {
		h.suite.t.Fatalf("Failed to create request: %v", err)
	}

	if h.body != nil {
		req.Header.Set("Content-Type", "application/json")
	}

	resp, err := h.suite.client.Do(req)
	if err != nil {
		h.suite.t.Fatalf("Failed to execute request: %v", err)
	}

	h.suite.t.Cleanup(func() {
		if err := resp.Body.Close(); err != nil {
			h.suite.t.Logf("Failed to close response body: %v", err)
		}
	})

	h.resp = resp

	return h
}

// ExpectStatus validates the response status code.
func (h *HTTPBuilder) ExpectStatus(statusCode int) *HTTPBuilder {
	if h.resp == nil {
		h.suite.t.Fatal("Request not executed. Call Execute() first.")
	}

	if h.resp.StatusCode != statusCode {
		body, _ := io.ReadAll(h.resp.Body)
		h.suite.t.Fatalf("Expected status code %d, got %d. Response body: %s",
			statusCode, h.resp.StatusCode, string(body))
	}

	return h
}


// serializeBody converts the body to JSON bytes.
func (h *HTTPBuilder) serializeBody() ([]byte, error) {
	if h.body == nil {
		return nil, nil
	}

	// If it's already a string, return as-is
	if str, ok := h.body.(string); ok {
		return []byte(str), nil
	}

	// Otherwise, marshal as JSON
	bytes, err := json.Marshal(h.body)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal body: %w", err)
	}

	return bytes, nil
}
