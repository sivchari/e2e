package e2e_test

import (
	"context"
	"testing"
	"time"

	"google.golang.org/grpc/codes"

	"github.com/sivchari/e2e"
	"github.com/sivchari/e2e/test/e2e/grpc/testserver"
	pb "github.com/sivchari/e2e/test/e2e/grpc/testserver/proto"
)

func TestGRPC_NewGRPC(t *testing.T) {
	addr := testserver.StartTestServer(t)

	// Should create a gRPC test suite
	client := e2e.NewGRPC(t, e2e.GRPCConfig{
		Target: addr,
	})

	if client == nil {
		t.Fatal("Expected non-nil client")
	}
}

func TestGRPC_Call_GetUser(t *testing.T) {
	addr := testserver.StartTestServer(t)
	client := e2e.NewGRPC(t, e2e.GRPCConfig{Target: addr})

	// Should call GetUser and expect OK
	client.Call("testpb.TestService/GetUser").
		Body(&pb.GetUserRequest{Id: "1"}).
		Execute(context.Background()).
		ExpectCode(codes.OK).
		ExpectMessage(&pb.GetUserResponse{
			Id:    "1",
			Name:  "Alice",
			Email: "alice@example.com",
		})
}

func TestGRPC_Call_NotFound(t *testing.T) {
	addr := testserver.StartTestServer(t)
	client := e2e.NewGRPC(t, e2e.GRPCConfig{Target: addr})

	// Should return NotFound for non-existent user
	client.Call("testpb.TestService/GetUser").
		Body(&pb.GetUserRequest{Id: "999"}).
		Execute(context.Background()).
		ExpectCode(codes.NotFound).
		ExpectErrorMessage("user 999 not found")
}

func TestGRPC_Call_InvalidArgument(t *testing.T) {
	addr := testserver.StartTestServer(t)
	client := e2e.NewGRPC(t, e2e.GRPCConfig{Target: addr})

	// Should return InvalidArgument for empty ID
	client.Call("testpb.TestService/GetUser").
		Body(&pb.GetUserRequest{Id: ""}).
		Execute(context.Background()).
		ExpectCode(codes.InvalidArgument)
}

func TestGRPC_Metadata(t *testing.T) {
	addr := testserver.StartTestServer(t)
	client := e2e.NewGRPC(t, e2e.GRPCConfig{Target: addr})

	// Should pass metadata to the server
	client.Call("testpb.TestService/Echo").
		Metadata("x-custom-header", "test-value").
		Body(&pb.EchoRequest{Message: "hello"}).
		Execute(context.Background()).
		ExpectCode(codes.OK)
}

func TestGRPC_Timeout(t *testing.T) {
	addr := testserver.StartTestServer(t)
	client := e2e.NewGRPC(t, e2e.GRPCConfig{Target: addr})

	// Should respect timeout
	client.Call("testpb.TestService/GetUser").
		Body(&pb.GetUserRequest{Id: "1"}).
		Timeout(5 * time.Second).
		Execute(context.Background()).
		ExpectCode(codes.OK)
}

func TestGRPC_CreateUser(t *testing.T) {
	addr := testserver.StartTestServer(t)
	client := e2e.NewGRPC(t, e2e.GRPCConfig{Target: addr})

	// Should create a new user
	client.Call("testpb.TestService/CreateUser").
		Body(&pb.CreateUserRequest{
			Name:  "Charlie",
			Email: "charlie@example.com",
		}).
		Execute(context.Background()).
		ExpectCode(codes.OK).
		ExpectMessage(&pb.CreateUserResponse{
			Id:    "3",
			Name:  "Charlie",
			Email: "charlie@example.com",
		})
}
