// Package testserver provides a GraphQL server for testing.
package testserver

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

// User represents a user in the test data.
type User struct {
	ID    string `json:"id"`
	Name  string `json:"name"`
	Email string `json:"email"`
}

// GraphQLRequest represents a GraphQL request.
type GraphQLRequest struct {
	Query     string         `json:"query"`
	Variables map[string]any `json:"variables,omitempty"`
}

// GraphQLResponse represents a GraphQL response.
type GraphQLResponse struct {
	Data   any            `json:"data,omitempty"`
	Errors []GraphQLError `json:"errors,omitempty"`
}

// GraphQLError represents a GraphQL error.
type GraphQLError struct {
	Message string `json:"message"`
}

// Server implements a simple GraphQL server for testing.
type Server struct {
	users map[string]*User
}

// NewServer creates a new test server.
func NewServer() *Server {
	return &Server{
		users: map[string]*User{
			"1": {ID: "1", Name: "Alice", Email: "alice@example.com"},
			"2": {ID: "2", Name: "Bob", Email: "bob@example.com"},
		},
	}
}

// ServeHTTP handles GraphQL requests.
func (s *Server) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)

		return
	}

	var req GraphQLRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		s.writeError(w, "Invalid request body")

		return
	}

	resp := s.handleQuery(req, r.Header)
	w.Header().Set("Content-Type", "application/json")

	if err := json.NewEncoder(w).Encode(resp); err != nil {
		http.Error(w, "Failed to encode response", http.StatusInternalServerError)
	}
}

func (s *Server) handleQuery(req GraphQLRequest, headers http.Header) GraphQLResponse {
	query := strings.TrimSpace(req.Query)

	// Simple query parsing for testing
	switch {
	case strings.Contains(query, "users"):
		return s.handleUsersQuery()
	case strings.Contains(query, "user("):
		return s.handleUserQuery(req.Variables)
	case strings.Contains(query, "createUser"):
		return s.handleCreateUser(req.Variables)
	case strings.Contains(query, "echo"):
		return s.handleEcho(req.Variables, headers)
	case strings.Contains(query, "error"):
		return s.handleError()
	default:
		return GraphQLResponse{
			Errors: []GraphQLError{{Message: "Unknown query"}},
		}
	}
}

func (s *Server) handleUsersQuery() GraphQLResponse {
	users := make([]*User, 0, len(s.users))
	for _, u := range s.users {
		users = append(users, u)
	}

	return GraphQLResponse{
		Data: map[string]any{"users": users},
	}
}

func (s *Server) handleUserQuery(vars map[string]any) GraphQLResponse {
	id, ok := vars["id"].(string)
	if !ok {
		return GraphQLResponse{
			Errors: []GraphQLError{{Message: "id is required"}},
		}
	}

	user, exists := s.users[id]
	if !exists {
		return GraphQLResponse{
			Errors: []GraphQLError{{Message: "user not found"}},
		}
	}

	return GraphQLResponse{
		Data: map[string]any{"user": user},
	}
}

func (s *Server) handleCreateUser(vars map[string]any) GraphQLResponse {
	input, ok := vars["input"].(map[string]any)
	if !ok {
		return GraphQLResponse{
			Errors: []GraphQLError{{Message: "input is required"}},
		}
	}

	name, _ := input["name"].(string)
	email, _ := input["email"].(string)

	user := &User{
		ID:    "3",
		Name:  name,
		Email: email,
	}

	return GraphQLResponse{
		Data: map[string]any{"createUser": user},
	}
}

func (s *Server) handleEcho(vars map[string]any, headers http.Header) GraphQLResponse {
	msg, _ := vars["message"].(string)

	return GraphQLResponse{
		Data: map[string]any{
			"echo": map[string]any{
				"message":       msg,
				"authorization": headers.Get("Authorization"),
			},
		},
	}
}

func (s *Server) handleError() GraphQLResponse {
	return GraphQLResponse{
		Errors: []GraphQLError{
			{Message: "Something went wrong"},
			{Message: "Another error"},
		},
	}
}

func (s *Server) writeError(w http.ResponseWriter, msg string) {
	w.Header().Set("Content-Type", "application/json")

	resp := GraphQLResponse{
		Errors: []GraphQLError{{Message: msg}},
	}

	if err := json.NewEncoder(w).Encode(resp); err != nil {
		http.Error(w, "Failed to encode response", http.StatusInternalServerError)
	}
}

// StartTestServer starts a GraphQL server for testing and returns the URL.
func StartTestServer(t *testing.T) string {
	t.Helper()

	srv := httptest.NewServer(NewServer())
	t.Cleanup(srv.Close)

	return srv.URL
}
