package e2e

import (
	"bytes"
	"encoding/json"
	"io"
	"net/http"
	"strings"
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
	suite      *TestSuite
	method     string
	path       string
	body       interface{}
	statusCode int
}

// New creates a new test suite.
func New(t *testing.T, config Config) *TestSuite {
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

// ExpectStatus sets the expected status code and executes the request.
func (h *HTTPBuilder) ExpectStatus(statusCode int) *HTTPBuilder {
	h.statusCode = statusCode
	h.execute()

	return h
}

// execute performs the HTTP request and validates the response.
func (h *HTTPBuilder) execute() {
	url := strings.TrimSuffix(h.suite.config.BaseURL, "/") + "/" + strings.TrimPrefix(h.path, "/")

	var bodyReader io.Reader

	if h.body != nil {
		bodyBytes, err := h.serializeBody()
		if err != nil {
			h.suite.t.Fatalf("Failed to serialize body: %v", err)
		}

		bodyReader = bytes.NewBuffer(bodyBytes)
	}

	req, err := http.NewRequest(h.method, url, bodyReader)
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
	defer resp.Body.Close()

	if resp.StatusCode != h.statusCode {
		body, _ := io.ReadAll(resp.Body)
		h.suite.t.Fatalf("Expected status code %d, got %d. Response body: %s",
			h.statusCode, resp.StatusCode, string(body))
	}
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
	return json.Marshal(h.body)
}
