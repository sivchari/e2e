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
	t      testing.TB
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

	// Request details for error reporting
	requestURL     string
	requestHeaders http.Header
	requestBody    []byte
	responseBody   []byte
}

// New creates a new test suite.
func New(tb testing.TB, config Config) *TestSuite {
	tb.Helper()

	// Set default timeout if not specified
	if config.Timeout == 0 {
		config.Timeout = 30 * time.Second
	}

	return &TestSuite{
		config: config,
		t:      tb,
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

	// Store request details for error reporting
	h.requestURL = reqURL.String()
	h.requestHeaders = req.Header.Clone()

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

	// Store request body for error reporting
	h.requestBody = bodyBytes

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
		expected := fmt.Sprintf("%d %s", statusCode, http.StatusText(statusCode))
		actual := fmt.Sprintf("%d %s", h.resp.StatusCode, http.StatusText(h.resp.StatusCode))
		h.suite.t.Fatal(h.formatError("Status code mismatch", expected, actual))
	}

	return h
}

// ExpectJSON validates the JSON response body.
func (h *HTTPBuilder) ExpectJSON(expected interface{}) *HTTPBuilder {
	if h.resp == nil {
		h.suite.t.Fatal("Request not executed. Call Execute() first.")
	}

	body := h.readResponseBody()

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
			h.suite.t.Fatal(h.formatError("JSON response mismatch", string(expectedJSON), string(actualJSON)))
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
			h.suite.t.Fatal(h.formatError("JSON response mismatch", string(expectedJSON), string(actualJSON)))
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
		assertion := fmt.Sprintf("Header mismatch (%s)", key)
		h.suite.t.Fatal(h.formatError(assertion, value, actualValue))
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

// readResponseBody reads the response body and caches it for reuse.
func (h *HTTPBuilder) readResponseBody() []byte {
	if h.responseBody != nil {
		return h.responseBody
	}

	body, err := io.ReadAll(h.resp.Body)
	if err != nil {
		h.suite.t.Fatalf("Failed to read response body: %v", err)
	}

	h.responseBody = body

	return body
}

const maxBodySize = 1024

// formatError creates a detailed error message with request/response information.
func (h *HTTPBuilder) formatError(assertion, expected, actual string) string {
	var sb bytes.Buffer

	sb.WriteString("\n=== HTTP Request Failed ===\n")
	sb.WriteString(fmt.Sprintf("Request:  %s %s\n", h.method, h.path))
	sb.WriteString(fmt.Sprintf("URL:      %s\n", h.requestURL))

	h.writeHeaders(&sb, "Request", h.requestHeaders)

	// Request body
	if len(h.requestBody) > 0 {
		sb.WriteString(fmt.Sprintf("Body:     %s\n", h.truncateBody(h.requestBody)))
	}

	sb.WriteString("\n")
	sb.WriteString(fmt.Sprintf("Response: %d %s\n", h.resp.StatusCode, http.StatusText(h.resp.StatusCode)))

	h.writeHeaders(&sb, "Response", h.resp.Header)

	// Response body
	respBody := h.readResponseBody()
	if len(respBody) > 0 {
		sb.WriteString(fmt.Sprintf("Body:     %s\n", h.truncateBody(respBody)))
	}

	sb.WriteString("\n")
	sb.WriteString(fmt.Sprintf("%s:\n", assertion))
	sb.WriteString(fmt.Sprintf("Expected: %s\n", expected))
	sb.WriteString(fmt.Sprintf("Actual:   %s\n", actual))

	return sb.String()
}

// writeHeaders writes headers to the buffer with proper formatting.
func (h *HTTPBuilder) writeHeaders(sb *bytes.Buffer, _ string, headers http.Header) {
	if len(headers) == 0 {
		return
	}

	sb.WriteString("Headers:  ")

	first := true
	for key, values := range headers {
		if !first {
			sb.WriteString("          ")
		}

		fmt.Fprintf(sb, "%s: %s\n", key, values[0])

		first = false
	}
}

// truncateBody truncates a body if it exceeds the maximum size.
func (h *HTTPBuilder) truncateBody(body []byte) string {
	if len(body) <= maxBodySize {
		return string(body)
	}

	return string(body[:maxBodySize]) + "... (truncated)"
}
