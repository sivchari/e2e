# e2e

A simple and intuitive E2E testing framework for Go, supporting HTTP, gRPC, and GraphQL.

## Features

- **Simple API** - Fluent method chaining for readable tests
- **Multi-Protocol** - HTTP, gRPC, and GraphQL support
- **Type Safe** - Compile-time error detection
- **IDE Friendly** - Full autocomplete support
- **Zero Config** - Start testing immediately

## Installation

```bash
go get github.com/sivchari/e2e
```

## Quick Start

### HTTP

```go
func TestAPI(t *testing.T) {
    client := e2e.New(t, e2e.Config{BaseURL: "http://localhost:8080"})

    client.GET("/users/1").
        Execute(context.Background()).
        ExpectStatus(200).
        ExpectJSON(map[string]any{
            "id":   "1",
            "name": "Alice",
        })
}
```

### gRPC

```go
func TestGRPC(t *testing.T) {
    client := e2e.NewGRPC(t, e2e.GRPCConfig{Target: "localhost:9090"})

    client.Call("user.UserService/GetUser").
        Body(&pb.GetUserRequest{Id: "1"}).
        Execute(context.Background()).
        ExpectCode(codes.OK).
        ExpectMessage(&pb.GetUserResponse{
            Id:   "1",
            Name: "Alice",
        })
}
```

### GraphQL

```go
func TestGraphQL(t *testing.T) {
    client := e2e.NewGraphQL(t, e2e.GraphQLConfig{BaseURL: "http://localhost:8080/graphql"})

    client.Query(`query($id: ID!) { user(id: $id) { id name } }`).
        Variables(map[string]any{"id": "1"}).
        Execute(context.Background()).
        ExpectNoErrors().
        ExpectData(map[string]any{
            "user": map[string]any{
                "id":   "1",
                "name": "Alice",
            },
        })
}
```

## HTTP API

### Creating a Client

```go
client := e2e.New(t, e2e.Config{
    BaseURL: "http://localhost:8080",
    Timeout: 30 * time.Second, // optional, default: 30s
})
```

### HTTP Methods

```go
client.GET("/path")
client.POST("/path")
client.PUT("/path")
client.DELETE("/path")
client.PATCH("/path")
client.HEAD("/path")
client.OPTIONS("/path")
```

### Request Configuration

```go
client.POST("/users").
    Body(User{Name: "Alice"}).           // struct, map, or JSON string
    Header("X-Request-ID", "123").        // custom header
    Query("page", "1").                   // query parameter
    Authorization("Bearer token").        // Authorization header
    Timeout(5 * time.Second).             // request timeout
    Execute(ctx)
```

### Response Validation

```go
client.GET("/users/1").
    Execute(ctx).
    ExpectStatus(200).                    // status code
    ExpectJSON(expected).                 // JSON body (struct, map, or string)
    ExpectHeader("Content-Type", "application/json")
```

## gRPC API

### Creating a Client

```go
client := e2e.NewGRPC(t, e2e.GRPCConfig{
    Target:  "localhost:9090",
    Timeout: 30 * time.Second, // optional
})
```

### Making Calls

```go
client.Call("package.Service/Method").
    Body(&pb.Request{}).                  // proto message
    Metadata("authorization", "Bearer token").  // gRPC metadata
    Timeout(5 * time.Second).
    Execute(ctx)
```

### Response Validation

```go
client.Call("user.UserService/GetUser").
    Body(&pb.GetUserRequest{Id: "1"}).
    Execute(ctx).
    ExpectCode(codes.OK).                 // gRPC status code
    ExpectMessage(&pb.GetUserResponse{    // proto message
        Id:   "1",
        Name: "Alice",
    }).
    ExpectErrorMessage("user not found")  // error message (for error cases)
```

## GraphQL API

### Creating a Client

```go
client := e2e.NewGraphQL(t, e2e.GraphQLConfig{
    BaseURL: "http://localhost:8080/graphql",
    Timeout: 30 * time.Second, // optional
})
```

### Queries and Mutations

```go
// Query
client.Query(`query { users { id name } }`).
    Execute(ctx)

// Query with variables
client.Query(`query($id: ID!) { user(id: $id) { id name } }`).
    Variables(map[string]any{"id": "1"}).
    Execute(ctx)

// Mutation
client.Mutation(`mutation($input: CreateUserInput!) { createUser(input: $input) { id } }`).
    Variables(map[string]any{
        "input": map[string]any{"name": "Alice"},
    }).
    Execute(ctx)
```

### Request Configuration

```go
client.Query(`query { users { id } }`).
    Variables(map[string]any{"limit": 10}).
    Header("Authorization", "Bearer token").
    Timeout(5 * time.Second).
    Execute(ctx)
```

### Response Validation

```go
client.Query(`query { user(id: "1") { name } }`).
    Execute(ctx).
    ExpectNoErrors().                     // no GraphQL errors
    ExpectData(expected).                 // full data validation
    ExpectDataPath("user.name", "Alice")  // path-based validation

// Error validation
client.Query(`query { user(id: "999") { name } }`).
    Execute(ctx).
    ExpectErrors().                       // expect errors
    ExpectErrorMessage("user not found")  // specific error message
```

## Error Messages

When assertions fail, detailed error messages are provided:

```
=== HTTP Request Failed ===
Request:  GET /users/1
URL:      http://localhost:8080/users/1
Headers:  Authorization: Bearer token

Response: 404 Not Found
Headers:  Content-Type: application/json
Body:     {"error": "user not found"}

Status code mismatch:
Expected: 200 OK
Actual:   404 Not Found
```

## API Comparison

| Feature | HTTP | gRPC | GraphQL |
|---------|------|------|---------|
| Factory | `New()` | `NewGRPC()` | `NewGraphQL()` |
| Request | `GET()/POST()/...` | `Call()` | `Query()/Mutation()` |
| Body | `Body()` | `Body()` | `Variables()` |
| Headers | `Header()` | `Metadata()` | `Header()` |
| Timeout | `Timeout()` | `Timeout()` | `Timeout()` |
| Execute | `Execute()` | `Execute()` | `Execute()` |
| Status | `ExpectStatus()` | `ExpectCode()` | `ExpectNoErrors()` |
| Response | `ExpectJSON()` | `ExpectMessage()` | `ExpectData()` |

## License

MIT License
