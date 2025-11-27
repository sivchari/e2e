package e2e_test

import (
	"context"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/sivchari/e2e"
)

// mockT is a mock testing.TB to capture fatal messages.
type mockT struct {
	testing.TB
	fatalMsg string
}

func (m *mockT) Fatal(args ...interface{}) {
	if len(args) > 0 {
		if msg, ok := args[0].(string); ok {
			m.fatalMsg = msg
		}
	}

	panic("fatal called")
}

func (m *mockT) Fatalf(format string, _ ...interface{}) {
	m.fatalMsg = format

	panic("fatal called")
}

func (m *mockT) Helper() {}

func (m *mockT) Cleanup(_ func()) {}

func (m *mockT) Logf(_ string, _ ...interface{}) {}

func TestErrorMessageContainsRequestDetails(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusBadRequest)
		_, _ = w.Write([]byte(`{"error":"validation failed"}`))
	}))
	defer server.Close()

	mt := &mockT{TB: t}

	defer func() {
		if r := recover(); r == nil {
			t.Fatal("Expected panic from Fatal call")
		}

		// Verify error message contains key information
		if !strings.Contains(mt.fatalMsg, "=== HTTP Request Failed ===") {
			t.Errorf("Error message should contain header, got: %s", mt.fatalMsg)
		}

		if !strings.Contains(mt.fatalMsg, "POST /api/users") {
			t.Errorf("Error message should contain request method and path, got: %s", mt.fatalMsg)
		}

		if !strings.Contains(mt.fatalMsg, "URL:") {
			t.Errorf("Error message should contain URL, got: %s", mt.fatalMsg)
		}

		if !strings.Contains(mt.fatalMsg, "Response: 400 Bad Request") {
			t.Errorf("Error message should contain response status, got: %s", mt.fatalMsg)
		}

		if !strings.Contains(mt.fatalMsg, "Expected: 201 Created") {
			t.Errorf("Error message should contain expected status, got: %s", mt.fatalMsg)
		}

		if !strings.Contains(mt.fatalMsg, "Actual:") || !strings.Contains(mt.fatalMsg, "400 Bad Request") {
			t.Errorf("Error message should contain actual status, got: %s", mt.fatalMsg)
		}
	}()

	client := e2e.New(mt, e2e.Config{BaseURL: server.URL})
	client.POST("/api/users").
		Body(map[string]string{"name": "Alice"}).
		Execute(context.Background()).
		ExpectStatus(http.StatusCreated)
}

func TestErrorMessageContainsRequestBody(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusBadRequest)
	}))
	defer server.Close()

	mt := &mockT{TB: t}

	defer func() {
		if r := recover(); r == nil {
			t.Fatal("Expected panic from Fatal call")
		}

		if !strings.Contains(mt.fatalMsg, `"name":"Alice"`) {
			t.Errorf("Error message should contain request body, got: %s", mt.fatalMsg)
		}
	}()

	client := e2e.New(mt, e2e.Config{BaseURL: server.URL})
	client.POST("/api/users").
		Body(map[string]string{"name": "Alice"}).
		Execute(context.Background()).
		ExpectStatus(http.StatusCreated)
}

func TestErrorMessageContainsRequestHeaders(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusUnauthorized)
	}))
	defer server.Close()

	mt := &mockT{TB: t}

	defer func() {
		if r := recover(); r == nil {
			t.Fatal("Expected panic from Fatal call")
		}

		if !strings.Contains(mt.fatalMsg, "Authorization: Bearer token123") {
			t.Errorf("Error message should contain request headers, got: %s", mt.fatalMsg)
		}
	}()

	client := e2e.New(mt, e2e.Config{BaseURL: server.URL})
	client.GET("/api/users").
		Authorization("Bearer token123").
		Execute(context.Background()).
		ExpectStatus(http.StatusOK)
}

func TestErrorMessageContainsResponseBody(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusBadRequest)
		_, _ = w.Write([]byte(`{"error":"invalid email format"}`))
	}))
	defer server.Close()

	mt := &mockT{TB: t}

	defer func() {
		if r := recover(); r == nil {
			t.Fatal("Expected panic from Fatal call")
		}

		if !strings.Contains(mt.fatalMsg, "invalid email format") {
			t.Errorf("Error message should contain response body, got: %s", mt.fatalMsg)
		}
	}()

	client := e2e.New(mt, e2e.Config{BaseURL: server.URL})
	client.POST("/api/users").
		Execute(context.Background()).
		ExpectStatus(http.StatusCreated)
}

func TestErrorMessageHeaderMismatch(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("X-Custom-Header", "wrong-value")
		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()

	mt := &mockT{TB: t}

	defer func() {
		if r := recover(); r == nil {
			t.Fatal("Expected panic from Fatal call")
		}

		if !strings.Contains(mt.fatalMsg, "Header mismatch (X-Custom-Header)") {
			t.Errorf("Error message should contain header mismatch info, got: %s", mt.fatalMsg)
		}

		if !strings.Contains(mt.fatalMsg, "Expected: expected-value") {
			t.Errorf("Error message should contain expected header value, got: %s", mt.fatalMsg)
		}

		if !strings.Contains(mt.fatalMsg, "Actual:") || !strings.Contains(mt.fatalMsg, "wrong-value") {
			t.Errorf("Error message should contain actual header value, got: %s", mt.fatalMsg)
		}
	}()

	client := e2e.New(mt, e2e.Config{BaseURL: server.URL})
	client.GET("/").
		Execute(context.Background()).
		ExpectHeader("X-Custom-Header", "expected-value")
}
