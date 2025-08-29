package e2e_test

import (
	"testing"

	"github.com/sivchari/e2e"
	"github.com/sivchari/e2e/test/e2e/testserver"
)

// User represents a test user struct.
type User struct {
	Name  string `json:"name"`
	Email string `json:"email"`
}

func TestRequestBodyStruct(t *testing.T) {
	server := testserver.NewEchoServer()
	defer server.Close()

	client := e2e.New(t, e2e.Config{BaseURL: server.URL})

	user := User{
		Name:  "Alice",
		Email: "alice@example.com",
	}

	client.POST("/users").
		Body(user).
		Execute(t.Context()).
		ExpectStatus(200).
		ExpectJSON(map[string]interface{}{
			"method": "POST",
			"path":   "/users",
			"body": map[string]interface{}{
				"name":  "Alice",
				"email": "alice@example.com",
			},
		})
}

func TestRequestBodyMap(t *testing.T) {
	server := testserver.NewEchoServer()
	defer server.Close()

	client := e2e.New(t, e2e.Config{BaseURL: server.URL})

	data := map[string]interface{}{
		"key":   "value",
		"count": 42,
	}

	client.POST("/data").
		Body(data).
		Execute(t.Context()).
		ExpectStatus(200).
		ExpectJSON(map[string]interface{}{
			"method": "POST",
			"path":   "/data",
			"body": map[string]interface{}{
				"key":   "value",
				"count": 42.0,
			},
		})
}

func TestRequestBodyJSONString(t *testing.T) {
	server := testserver.NewEchoServer()
	defer server.Close()

	client := e2e.New(t, e2e.Config{BaseURL: server.URL})

	jsonStr := `{"message":"hello world"}`

	client.POST("/message").
		Body(jsonStr).
		Execute(t.Context()).
		ExpectStatus(200).
		ExpectJSON(map[string]interface{}{
			"method": "POST",
			"path":   "/message",
			"body": map[string]interface{}{
				"message": "hello world",
			},
		})
}
