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
	"time"
)

// Config holds configuration for the test suite.
type Config struct {
	BaseURL string
	Timeout time.Duration // Default timeout for requests
}

// TestSuite represents the main test suite.
type TestSuite struct {
	config Config
	t      *testing.T
	client *http.Client
}

// HTTPBuilder builds HTTP requests.
type HTTPBuilder struct {
	suite   *TestSuite
	method  string
	path    string
	body    interface{}
	headers map[string]string
	query   map[string]string
	timeout time.Duration
	resp    *http.Response
}

// New creates a new test suite.
func New(t *testing.T, config Config) *TestSuite {
	t.Helper()

	// Set default timeout if not specified
	if config.Timeout == 0 {
		config.Timeout = 30 * time.Second
	}

	return &TestSuite{
		config: config,
		t:      t,
		client: &http.Client{
			Timeout: config.Timeout,
		},
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

// PUT creates a PUT request builder.
func (s *TestSuite) PUT(path string) *HTTPBuilder {
	return &HTTPBuilder{
		suite:  s,
		method: "PUT",
		path:   path,
	}
}

// DELETE creates a DELETE request builder.
func (s *TestSuite) DELETE(path string) *HTTPBuilder {
	return &HTTPBuilder{
		suite:  s,
		method: "DELETE",
		path:   path,
	}
}

// PATCH creates a PATCH request builder.
func (s *TestSuite) PATCH(path string) *HTTPBuilder {
	return &HTTPBuilder{
		suite:  s,
		method: "PATCH",
		path:   path,
	}
}

// HEAD creates a HEAD request builder.
func (s *TestSuite) HEAD(path string) *HTTPBuilder {
	return &HTTPBuilder{
		suite:  s,
		method: "HEAD",
		path:   path,
	}
}

// OPTIONS creates an OPTIONS request builder.
func (s *TestSuite) OPTIONS(path string) *HTTPBuilder {
	return &HTTPBuilder{
		suite:  s,
		method: "OPTIONS",
		path:   path,
	}
}

// Body sets the request body (accepts string, struct, or map).
func (h *HTTPBuilder) Body(body interface{}) *HTTPBuilder {
	h.body = body

	return h
}

// Header sets a request header.
func (h *HTTPBuilder) Header(key, value string) *HTTPBuilder {
	if h.headers == nil {
		h.headers = make(map[string]string)
	}

	h.headers[key] = value

	return h
}

// Query adds a query parameter.
func (h *HTTPBuilder) Query(key, value string) *HTTPBuilder {
	if h.query == nil {
		h.query = make(map[string]string)
	}

	h.query[key] = value

	return h
}

// Authorization sets the Authorization header.
func (h *HTTPBuilder) Authorization(value string) *HTTPBuilder {
	return h.Header("Authorization", value)
}

// Timeout sets a custom timeout for this specific request.
func (h *HTTPBuilder) Timeout(timeout time.Duration) *HTTPBuilder {
	h.timeout = timeout

	return h
}

// Execute performs the HTTP request.
func (h *HTTPBuilder) Execute(ctx context.Context) *HTTPBuilder {
	reqURL := h.buildURL()
	bodyReader := h.prepareBody()
	ctx = h.applyTimeout(ctx)
	req := h.createRequest(ctx, reqURL, bodyReader)
	h.setHeaders(req)
	h.executeRequest(req, reqURL)

	return h
}

func (h *HTTPBuilder) buildURL() *url.URL {
	baseURL, err := url.Parse(h.suite.config.BaseURL)
	if err != nil {
		h.suite.t.Fatalf("Failed to parse base URL: %v", err)
	}

	path, err := url.Parse(h.path)
	if err != nil {
		h.suite.t.Fatalf("Failed to parse path: %v", err)
	}

	reqURL := baseURL.ResolveReference(path)

	// Add query parameters
	if len(h.query) > 0 {
		q := reqURL.Query()
		for key, value := range h.query {
			q.Add(key, value)
		}

		reqURL.RawQuery = q.Encode()
	}

	return reqURL
}

func (h *HTTPBuilder) prepareBody() io.Reader {
	if h.body == nil {
		return nil
	}

	bodyBytes, err := h.serializeBody()
	if err != nil {
		h.suite.t.Fatalf("Failed to serialize body: %v", err)
	}

	return bytes.NewBuffer(bodyBytes)
}

func (h *HTTPBuilder) applyTimeout(ctx context.Context) context.Context {
	if h.timeout <= 0 {
		return ctx
	}

	newCtx, cancel := context.WithTimeout(ctx, h.timeout)
	h.suite.t.Cleanup(cancel)

	return newCtx
}

func (h *HTTPBuilder) createRequest(ctx context.Context, reqURL *url.URL, bodyReader io.Reader) *http.Request {
	req, err := http.NewRequestWithContext(ctx, h.method, reqURL.String(), bodyReader)
	if err != nil {
		h.suite.t.Fatalf("Failed to create request: %v", err)
	}

	return req
}

func (h *HTTPBuilder) setHeaders(req *http.Request) {
	if h.body != nil {
		req.Header.Set("Content-Type", "application/json")
	}

	for key, value := range h.headers {
		req.Header.Set(key, value)
	}
}

func (h *HTTPBuilder) executeRequest(req *http.Request, reqURL *url.URL) {
	client := h.suite.client
	if h.timeout > 0 {
		client = &http.Client{
			Timeout: h.timeout,
		}
	}

	resp, err := client.Do(req)
	if err != nil {
		h.suite.t.Fatalf("Failed to execute %s request to %s: %v", h.method, reqURL.String(), err)
	}

	h.suite.t.Cleanup(func() {
		if err := resp.Body.Close(); err != nil {
			h.suite.t.Logf("Failed to close response body: %v", err)
		}
	})

	h.resp = resp
}

// ExpectStatus validates the response status code.
func (h *HTTPBuilder) ExpectStatus(statusCode int) *HTTPBuilder {
	if h.resp == nil {
		h.suite.t.Fatal("Request not executed. Call Execute() first.")
	}

	if h.resp.StatusCode != statusCode {
		body, _ := io.ReadAll(h.resp.Body)
		h.suite.t.Fatalf("Status code mismatch for %s %s:\n  Expected: %d\n  Actual: %d\n  Response body: %s",
			h.method, h.path, statusCode, h.resp.StatusCode, string(body))
	}

	return h
}

// ExpectJSON validates the JSON response body.
func (h *HTTPBuilder) ExpectJSON(expected interface{}) *HTTPBuilder {
	if h.resp == nil {
		h.suite.t.Fatal("Request not executed. Call Execute() first.")
	}

	body, err := io.ReadAll(h.resp.Body)
	if err != nil {
		h.suite.t.Fatalf("Failed to read response body: %v", err)
	}

	// Parse actual response
	var actual interface{}
	if err := json.Unmarshal(body, &actual); err != nil {
		h.suite.t.Fatalf("Failed to parse JSON response: %v. Body: %s", err, string(body))
	}

	// Compare based on expected type
	switch exp := expected.(type) {
	case string:
		// If expected is a JSON string, parse it first
		var expectedParsed interface{}
		if err := json.Unmarshal([]byte(exp), &expectedParsed); err != nil {
			h.suite.t.Fatalf("Failed to parse expected JSON: %v", err)
		}

		if !jsonEqual(expectedParsed, actual) {
			expectedJSON, _ := json.MarshalIndent(expectedParsed, "", "  ")
			actualJSON, _ := json.MarshalIndent(actual, "", "  ")
			h.suite.t.Fatalf("JSON response mismatch for %s %s:\n\nExpected:\n%s\n\nActual:\n%s",
				h.method, h.path, string(expectedJSON), string(actualJSON))
		}
	default:
		// Marshal expected to JSON and back to normalize it
		expectedBytes, err := json.Marshal(expected)
		if err != nil {
			h.suite.t.Fatalf("Failed to marshal expected value: %v", err)
		}

		var expectedNormalized interface{}
		if err := json.Unmarshal(expectedBytes, &expectedNormalized); err != nil {
			h.suite.t.Fatalf("Failed to normalize expected value: %v", err)
		}

		if !jsonEqual(expectedNormalized, actual) {
			expectedJSON, _ := json.MarshalIndent(expectedNormalized, "", "  ")
			actualJSON, _ := json.MarshalIndent(actual, "", "  ")
			h.suite.t.Fatalf("JSON response mismatch for %s %s:\n\nExpected:\n%s\n\nActual:\n%s",
				h.method, h.path, string(expectedJSON), string(actualJSON))
		}
	}

	return h
}

// ExpectHeader validates a response header.
func (h *HTTPBuilder) ExpectHeader(key, value string) *HTTPBuilder {
	if h.resp == nil {
		h.suite.t.Fatal("Request not executed. Call Execute() first.")
	}

	actualValue := h.resp.Header.Get(key)
	if actualValue != value {
		h.suite.t.Fatalf("Header mismatch for %s %s:\n  Header: %s\n  Expected: %q\n  Actual: %q",
			h.method, h.path, key, value, actualValue)
	}

	return h
}

// jsonEqual compares two JSON values for equality.
func jsonEqual(a, b interface{}) bool {
	aBytes, _ := json.Marshal(a)
	bBytes, _ := json.Marshal(b)

	return bytes.Equal(aBytes, bBytes)
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
