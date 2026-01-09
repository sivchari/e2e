package e2e_test

import (
	"context"
	"testing"
	"time"

	"github.com/sivchari/e2e"
	"github.com/sivchari/e2e/test/e2e/graphql/testserver"
)

func TestGraphQL_NewGraphQL(t *testing.T) {
	url := testserver.StartTestServer(t)

	client := e2e.NewGraphQL(t, e2e.GraphQLConfig{
		BaseURL: url,
	})

	if client == nil {
		t.Fatal("Expected non-nil client")
	}
}

func TestGraphQL_Query_Users(t *testing.T) {
	url := testserver.StartTestServer(t)
	client := e2e.NewGraphQL(t, e2e.GraphQLConfig{BaseURL: url})

	client.Query(`query { users { id name } }`).
		Execute(context.Background()).
		ExpectNoErrors()
}

func TestGraphQL_Query_User(t *testing.T) {
	url := testserver.StartTestServer(t)
	client := e2e.NewGraphQL(t, e2e.GraphQLConfig{BaseURL: url})

	client.Query(`query($id: ID!) { user(id: $id) { id name email } }`).
		Variables(map[string]any{"id": "1"}).
		Execute(context.Background()).
		ExpectNoErrors().
		ExpectData(map[string]any{
			"user": map[string]any{
				"id":    "1",
				"name":  "Alice",
				"email": "alice@example.com",
			},
		})
}

func TestGraphQL_Query_UserNotFound(t *testing.T) {
	url := testserver.StartTestServer(t)
	client := e2e.NewGraphQL(t, e2e.GraphQLConfig{BaseURL: url})

	client.Query(`query($id: ID!) { user(id: $id) { id name } }`).
		Variables(map[string]any{"id": "999"}).
		Execute(context.Background()).
		ExpectErrors().
		ExpectErrorMessage("user not found")
}

func TestGraphQL_Mutation_CreateUser(t *testing.T) {
	url := testserver.StartTestServer(t)
	client := e2e.NewGraphQL(t, e2e.GraphQLConfig{BaseURL: url})

	client.Mutation(`mutation($input: CreateUserInput!) { createUser(input: $input) { id name email } }`).
		Variables(map[string]any{
			"input": map[string]any{
				"name":  "Charlie",
				"email": "charlie@example.com",
			},
		}).
		Execute(context.Background()).
		ExpectNoErrors().
		ExpectData(map[string]any{
			"createUser": map[string]any{
				"id":    "3",
				"name":  "Charlie",
				"email": "charlie@example.com",
			},
		})
}

func TestGraphQL_Header(t *testing.T) {
	url := testserver.StartTestServer(t)
	client := e2e.NewGraphQL(t, e2e.GraphQLConfig{BaseURL: url})

	client.Query(`query($message: String!) { echo(message: $message) { message authorization } }`).
		Variables(map[string]any{"message": "hello"}).
		Header("Authorization", "Bearer test-token").
		Execute(context.Background()).
		ExpectNoErrors().
		ExpectDataPath("echo.authorization", "Bearer test-token")
}

func TestGraphQL_Timeout(t *testing.T) {
	url := testserver.StartTestServer(t)
	client := e2e.NewGraphQL(t, e2e.GraphQLConfig{BaseURL: url})

	client.Query(`query { users { id } }`).
		Timeout(5 * time.Second).
		Execute(context.Background()).
		ExpectNoErrors()
}

func TestGraphQL_MultipleErrors(t *testing.T) {
	url := testserver.StartTestServer(t)
	client := e2e.NewGraphQL(t, e2e.GraphQLConfig{BaseURL: url})

	client.Query(`query { error }`).
		Execute(context.Background()).
		ExpectErrors().
		ExpectErrorMessage("Something went wrong")
}
