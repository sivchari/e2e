package e2e

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"
)

const (
	apiUsersPath   = "/api/users"
	httpMethodPost = "POST"
)

func TestBasicGET(t *testing.T) {
	// Create a test server
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/api/health" {
			w.WriteHeader(http.StatusOK)

			if _, err := fmt.Fprintln(w, `{"status": "ok"}`); err != nil {
				t.Logf("Failed to write response: %v", err)
			}
		} else {
			w.WriteHeader(http.StatusNotFound)
		}
	}))
	defer server.Close()

	// Test GET request
	New(t, Config{BaseURL: server.URL}).
		GET("/api/health").
		Execute(t.Context()).
		ExpectStatus(200)
}

func TestBasicPOST(t *testing.T) {
	// Create a test server
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == apiUsersPath && r.Method == httpMethodPost {
			w.WriteHeader(http.StatusCreated)

			if _, err := fmt.Fprintln(w, `{"id": "123", "name": "Alice"}`); err != nil {
				t.Logf("Failed to write response: %v", err)
			}
		} else {
			w.WriteHeader(http.StatusNotFound)
		}
	}))
	defer server.Close()

	// Test POST request
	New(t, Config{BaseURL: server.URL}).
		POST("/api/users").
		Body(`{"name": "Alice"}`).
		Execute(t.Context()).
		ExpectStatus(201)
}

func TestCombinedRequests(t *testing.T) {
	// Create a test server
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case "/api/health":
			w.WriteHeader(http.StatusOK)

			if _, err := fmt.Fprintln(w, `{"status": "ok"}`); err != nil {
				t.Logf("Failed to write response: %v", err)
			}
		case apiUsersPath:
			if r.Method == httpMethodPost {
				w.WriteHeader(http.StatusCreated)
				if _, err := fmt.Fprintln(w, `{"id": "123", "name": "Alice"}`); err != nil {
					t.Logf("Failed to write response: %v", err)
				}
			}
		default:
			w.WriteHeader(http.StatusNotFound)
		}
	}))
	defer server.Close()

	// Test combined requests
	suite := New(t, Config{BaseURL: server.URL})

	suite.GET("/api/health").Execute(t.Context()).ExpectStatus(200)
	suite.POST("/api/users").Body(`{"name": "Alice"}`).Execute(t.Context()).ExpectStatus(201)
}

func TestMultipleEndpoints(t *testing.T) {
	// Create API server
	apiServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == apiUsersPath {
			w.WriteHeader(http.StatusOK)
			if _, err := fmt.Fprintln(w, `{"users": []}`); err != nil {
				t.Logf("Failed to write response: %v", err)
			}
		} else {
			w.WriteHeader(http.StatusNotFound)
		}
	}))
	defer apiServer.Close()

	// Create Auth server
	authServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/auth/login" && r.Method == "POST" {
			w.WriteHeader(http.StatusOK)

			if _, err := fmt.Fprintln(w, `{"token": "abc123"}`); err != nil {
				t.Logf("Failed to write response: %v", err)
			}
		} else {
			w.WriteHeader(http.StatusNotFound)
		}
	}))
	defer authServer.Close()

	// Test multiple endpoints with separate clients
	apiClient := New(t, Config{BaseURL: apiServer.URL})
	authClient := New(t, Config{BaseURL: authServer.URL})

	apiClient.GET("/api/users").Execute(t.Context()).ExpectStatus(200)
	authClient.POST("/auth/login").Body(`{"username": "test"}`).Execute(t.Context()).ExpectStatus(200)
}

func TestStructBody(t *testing.T) {
	// Define test structs
	type User struct {
		Name  string `json:"name"`
		Email string `json:"email"`
	}

	type LoginRequest struct {
		Username string `json:"username"`
		Password string `json:"password"`
	}

	// Create a test server
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch {
		case r.URL.Path == apiUsersPath && r.Method == httpMethodPost:
			var user User
			if err := json.NewDecoder(r.Body).Decode(&user); err != nil {
				w.WriteHeader(http.StatusBadRequest)

				return
			}

			if user.Name == "Alice" && user.Email == "alice@example.com" {
				w.WriteHeader(http.StatusCreated)

				if _, err := fmt.Fprintln(w, `{"id": "123", "name": "Alice", "email": "alice@example.com"}`); err != nil {
					t.Logf("Failed to write response: %v", err)
				}
			} else {
				w.WriteHeader(http.StatusBadRequest)
			}
		case r.URL.Path == "/auth/login" && r.Method == httpMethodPost:
			var login LoginRequest
			if err := json.NewDecoder(r.Body).Decode(&login); err != nil {
				w.WriteHeader(http.StatusBadRequest)

				return
			}

			if login.Username == "testuser" {
				w.WriteHeader(http.StatusOK)
				if _, err := fmt.Fprintln(w, `{"token": "abc123"}`); err != nil {
					t.Logf("Failed to write response: %v", err)
				}
			} else {
				w.WriteHeader(http.StatusUnauthorized)
			}
		default:
			w.WriteHeader(http.StatusNotFound)
		}
	}))
	defer server.Close()

	// Test struct body
	client := New(t, Config{BaseURL: server.URL})

	// Test with struct
	user := User{Name: "Alice", Email: "alice@example.com"}
	client.POST("/api/users").Body(user).Execute(t.Context()).ExpectStatus(201)

	// Test with map
	loginData := map[string]interface{}{
		"username": "testuser",
		"password": "secret123",
	}
	client.POST("/auth/login").Body(loginData).Execute(t.Context()).ExpectStatus(200)

	// Test with string (backward compatibility)
	client.POST("/auth/login").Body(`{"username": "testuser", "password": "secret123"}`).Execute(t.Context()).ExpectStatus(200)
}
