package e2e

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"testing"
	"time"
)

// GraphQLConfig holds configuration for the GraphQL test suite.
type GraphQLConfig struct {
	BaseURL string        // GraphQL endpoint URL
	Timeout time.Duration // Default timeout for requests
}

// GraphQLSuite represents the GraphQL test suite.
type GraphQLSuite struct {
	config GraphQLConfig
	t      testing.TB
	client *http.Client
}

// GraphQLBuilder builds GraphQL requests.
type GraphQLBuilder struct {
	suite     *GraphQLSuite
	query     string
	variables map[string]any
	headers   map[string]string
	timeout   time.Duration

	// Response
	response *GraphQLResponse
}

// GraphQLResponse represents a GraphQL response.
type GraphQLResponse struct {
	Data   json.RawMessage `json:"data,omitempty"`
	Errors []GraphQLError  `json:"errors,omitempty"`
}

// GraphQLError represents a GraphQL error.
type GraphQLError struct {
	Message   string `json:"message"`
	Locations []struct {
		Line   int `json:"line"`
		Column int `json:"column"`
	} `json:"locations,omitempty"`
	Path []any `json:"path,omitempty"`
}

// NewGraphQL creates a new GraphQL test suite.
func NewGraphQL(tb testing.TB, config GraphQLConfig) *GraphQLSuite {
	tb.Helper()

	if config.Timeout == 0 {
		config.Timeout = 30 * time.Second
	}

	return &GraphQLSuite{
		config: config,
		t:      tb,
		client: &http.Client{
			Timeout: config.Timeout,
		},
	}
}

// Query creates a GraphQL query builder.
func (s *GraphQLSuite) Query(query string) *GraphQLBuilder {
	return &GraphQLBuilder{
		suite: s,
		query: query,
	}
}

// Mutation creates a GraphQL mutation builder.
func (s *GraphQLSuite) Mutation(mutation string) *GraphQLBuilder {
	return &GraphQLBuilder{
		suite: s,
		query: mutation,
	}
}

// Variables sets GraphQL variables.
func (b *GraphQLBuilder) Variables(vars map[string]any) *GraphQLBuilder {
	b.variables = vars

	return b
}

// Header sets an HTTP header.
func (b *GraphQLBuilder) Header(key, value string) *GraphQLBuilder {
	if b.headers == nil {
		b.headers = make(map[string]string)
	}

	b.headers[key] = value

	return b
}

// Timeout sets request timeout.
func (b *GraphQLBuilder) Timeout(d time.Duration) *GraphQLBuilder {
	b.timeout = d

	return b
}

// Execute performs the GraphQL request.
func (b *GraphQLBuilder) Execute(ctx context.Context) *GraphQLBuilder {
	if b.timeout > 0 {
		var cancel context.CancelFunc
		ctx, cancel = context.WithTimeout(ctx, b.timeout)
		b.suite.t.Cleanup(cancel)
	}

	// Build request body
	reqBody := map[string]any{
		"query": b.query,
	}

	if len(b.variables) > 0 {
		reqBody["variables"] = b.variables
	}

	bodyBytes, err := json.Marshal(reqBody)
	if err != nil {
		b.suite.t.Fatalf("Failed to marshal GraphQL request: %v", err)
	}

	// Create HTTP request
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, b.suite.config.BaseURL, bytes.NewReader(bodyBytes))
	if err != nil {
		b.suite.t.Fatalf("Failed to create request: %v", err)
	}

	req.Header.Set("Content-Type", "application/json")

	for key, value := range b.headers {
		req.Header.Set(key, value)
	}

	// Execute request
	client := b.suite.client
	if b.timeout > 0 {
		client = &http.Client{Timeout: b.timeout}
	}

	resp, err := client.Do(req)
	if err != nil {
		b.suite.t.Fatalf("Failed to execute GraphQL request: %v", err)
	}

	defer func() {
		if closeErr := resp.Body.Close(); closeErr != nil {
			b.suite.t.Logf("Failed to close response body: %v", closeErr)
		}
	}()

	// Read response
	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		b.suite.t.Fatalf("Failed to read response body: %v", err)
	}

	// Parse GraphQL response
	var gqlResp GraphQLResponse
	if err := json.Unmarshal(respBody, &gqlResp); err != nil {
		b.suite.t.Fatalf("Failed to parse GraphQL response: %v. Body: %s", err, string(respBody))
	}

	b.response = &gqlResp

	return b
}

// ExpectNoErrors validates no errors in response.
func (b *GraphQLBuilder) ExpectNoErrors() *GraphQLBuilder {
	if b.response == nil {
		b.suite.t.Fatal("Request not executed. Call Execute() first.")
	}

	if len(b.response.Errors) > 0 {
		var msgs []string
		for _, e := range b.response.Errors {
			msgs = append(msgs, e.Message)
		}

		b.suite.t.Fatal(b.formatError(
			"Expected no GraphQL errors",
			"no errors",
			strings.Join(msgs, ", "),
		))
	}

	return b
}

// ExpectErrors validates errors exist in response.
func (b *GraphQLBuilder) ExpectErrors() *GraphQLBuilder {
	if b.response == nil {
		b.suite.t.Fatal("Request not executed. Call Execute() first.")
	}

	if len(b.response.Errors) == 0 {
		b.suite.t.Fatal(b.formatError(
			"Expected GraphQL errors",
			"at least one error",
			"no errors",
		))
	}

	return b
}

// ExpectErrorMessage validates error message exists.
func (b *GraphQLBuilder) ExpectErrorMessage(msg string) *GraphQLBuilder {
	if b.response == nil {
		b.suite.t.Fatal("Request not executed. Call Execute() first.")
	}

	for _, e := range b.response.Errors {
		if e.Message == msg {
			return b
		}
	}

	msgs := make([]string, 0, len(b.response.Errors))
	for _, e := range b.response.Errors {
		msgs = append(msgs, e.Message)
	}

	b.suite.t.Fatal(b.formatError(
		"Expected error message not found",
		msg,
		strings.Join(msgs, ", "),
	))

	return b
}

// ExpectData validates the entire data response.
func (b *GraphQLBuilder) ExpectData(expected any) *GraphQLBuilder {
	if b.response == nil {
		b.suite.t.Fatal("Request not executed. Call Execute() first.")
	}

	// Marshal expected to JSON for comparison
	expectedBytes, err := json.Marshal(expected)
	if err != nil {
		b.suite.t.Fatalf("Failed to marshal expected data: %v", err)
	}

	// Normalize both for comparison
	var expectedNorm, actualNorm any
	if err := json.Unmarshal(expectedBytes, &expectedNorm); err != nil {
		b.suite.t.Fatalf("Failed to normalize expected data: %v", err)
	}

	if err := json.Unmarshal(b.response.Data, &actualNorm); err != nil {
		b.suite.t.Fatalf("Failed to parse response data: %v", err)
	}

	if !jsonEqual(expectedNorm, actualNorm) {
		expectedJSON, _ := json.MarshalIndent(expectedNorm, "", "  ")
		actualJSON, _ := json.MarshalIndent(actualNorm, "", "  ")
		b.suite.t.Fatal(b.formatError(
			"GraphQL data mismatch",
			string(expectedJSON),
			string(actualJSON),
		))
	}

	return b
}

// ExpectDataPath validates a specific path in data.
func (b *GraphQLBuilder) ExpectDataPath(path string, expected any) *GraphQLBuilder {
	if b.response == nil {
		b.suite.t.Fatal("Request not executed. Call Execute() first.")
	}

	// Parse response data
	var data map[string]any
	if err := json.Unmarshal(b.response.Data, &data); err != nil {
		b.suite.t.Fatalf("Failed to parse response data: %v", err)
	}

	// Navigate path
	parts := strings.Split(path, ".")

	var current any = data

	for _, part := range parts {
		switch v := current.(type) {
		case map[string]any:
			var ok bool
			current, ok = v[part]

			if !ok {
				b.suite.t.Fatal(b.formatError(
					fmt.Sprintf("Path %q not found at %q", path, part),
					fmt.Sprintf("%v", expected),
					"<missing>",
				))
			}
		default:
			b.suite.t.Fatalf("Cannot navigate path %q: unexpected type at %q", path, part)
		}
	}

	// Compare values
	if current != expected {
		b.suite.t.Fatal(b.formatError(
			fmt.Sprintf("Data at path %q mismatch", path),
			fmt.Sprintf("%v", expected),
			fmt.Sprintf("%v", current),
		))
	}

	return b
}

// formatError creates a detailed error message.
func (b *GraphQLBuilder) formatError(assertion, expected, actual string) string {
	var sb bytes.Buffer

	sb.WriteString("\n=== GraphQL Request Failed ===\n")
	sb.WriteString(fmt.Sprintf("Endpoint: %s\n", b.suite.config.BaseURL))
	sb.WriteString(fmt.Sprintf("Query:    %s\n", truncateString(b.query, 100)))

	if len(b.variables) > 0 {
		varsJSON, _ := json.Marshal(b.variables)
		sb.WriteString(fmt.Sprintf("Vars:     %s\n", truncateString(string(varsJSON), 100)))
	}

	if b.response != nil && len(b.response.Errors) > 0 {
		sb.WriteString("\nErrors:\n")

		for _, e := range b.response.Errors {
			sb.WriteString(fmt.Sprintf("  - %s\n", e.Message))
		}
	}

	sb.WriteString("\n")
	sb.WriteString(fmt.Sprintf("%s:\n", assertion))
	sb.WriteString(fmt.Sprintf("Expected: %s\n", expected))
	sb.WriteString(fmt.Sprintf("Actual:   %s\n", actual))

	return sb.String()
}

// truncateString truncates a string if it exceeds the maximum length.
func truncateString(s string, maxLen int) string {
	if len(s) <= maxLen {
		return s
	}

	return s[:maxLen] + "..."
}
