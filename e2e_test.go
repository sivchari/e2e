package e2e

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestBasicGET(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/api/health" {
			w.WriteHeader(http.StatusOK)
			w.Write([]byte(`{"status": "ok"}`))
		} else {
			w.WriteHeader(http.StatusNotFound)
		}
	}))
	defer server.Close()

	New(t, Config{BaseURL: server.URL}).
		GET("/api/health").
		Execute(t.Context()).
		ExpectStatus(200)
}

func TestBasicPOST(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/api/users" && r.Method == "POST" {
			w.WriteHeader(http.StatusCreated)
			w.Write([]byte(`{"id": "123", "name": "Alice"}`))
		} else {
			w.WriteHeader(http.StatusNotFound)
		}
	}))
	defer server.Close()

	New(t, Config{BaseURL: server.URL}).
		POST("/api/users").
		Body(`{"name": "Alice"}`).
		Execute(t.Context()).
		ExpectStatus(201)
}

func TestCombinedRequests(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case "/api/health":
			w.WriteHeader(http.StatusOK)
			w.Write([]byte(`{"status": "ok"}`))
		case "/api/users":
			if r.Method == "POST" {
				w.WriteHeader(http.StatusCreated)
				w.Write([]byte(`{"id": "123", "name": "Alice"}`))
			}
		default:
			w.WriteHeader(http.StatusNotFound)
		}
	}))
	defer server.Close()

	suite := New(t, Config{BaseURL: server.URL})
	suite.GET("/api/health").Execute(t.Context()).ExpectStatus(200)
	suite.POST("/api/users").Body(`{"name": "Alice"}`).Execute(t.Context()).ExpectStatus(201)
}

func TestStructBody(t *testing.T) {
	type User struct {
		Name  string `json:"name"`
		Email string `json:"email"`
	}

	type LoginRequest struct {
		Username string `json:"username"`
		Password string `json:"password"`
	}

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch {
		case r.URL.Path == "/api/users" && r.Method == "POST":
			var user User
			if err := json.NewDecoder(r.Body).Decode(&user); err != nil {
				w.WriteHeader(http.StatusBadRequest)
				return
			}
			if user.Name == "Alice" && user.Email == "alice@example.com" {
				w.WriteHeader(http.StatusCreated)
				w.Write([]byte(`{"id": "123", "name": "Alice", "email": "alice@example.com"}`))
			} else {
				w.WriteHeader(http.StatusBadRequest)
			}
		case r.URL.Path == "/auth/login" && r.Method == "POST":
			var login LoginRequest
			if err := json.NewDecoder(r.Body).Decode(&login); err != nil {
				w.WriteHeader(http.StatusBadRequest)
				return
			}
			if login.Username == "testuser" {
				w.WriteHeader(http.StatusOK)
				w.Write([]byte(`{"token": "abc123"}`))
			} else {
				w.WriteHeader(http.StatusUnauthorized)
			}
		default:
			w.WriteHeader(http.StatusNotFound)
		}
	}))
	defer server.Close()

	client := New(t, Config{BaseURL: server.URL})

	user := User{Name: "Alice", Email: "alice@example.com"}
	client.POST("/api/users").Body(user).Execute(t.Context()).ExpectStatus(201)

	loginData := map[string]any{
		"username": "testuser",
		"password": "secret123",
	}
	client.POST("/auth/login").Body(loginData).Execute(t.Context()).ExpectStatus(200)

	client.POST("/auth/login").Body(`{"username": "testuser", "password": "secret123"}`).Execute(t.Context()).ExpectStatus(200)
}
