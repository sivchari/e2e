// Package testserver provides a gRPC server for testing.
package testserver

import (
	"context"
	"fmt"
	"net"
	"testing"

	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/status"

	pb "github.com/sivchari/e2e/test/e2e/grpc/testserver/proto"
)

// Server implements the TestService for testing.
type Server struct {
	pb.UnimplementedTestServiceServer
	users map[string]*pb.GetUserResponse
}

// NewServer creates a new test server.
func NewServer() *Server {
	return &Server{
		users: map[string]*pb.GetUserResponse{
			"1": {Id: "1", Name: "Alice", Email: "alice@example.com"},
			"2": {Id: "2", Name: "Bob", Email: "bob@example.com"},
		},
	}
}

// GetUser returns a user by ID.
func (s *Server) GetUser(_ context.Context, req *pb.GetUserRequest) (*pb.GetUserResponse, error) {
	if req.Id == "" {
		return nil, status.Error(codes.InvalidArgument, "id is required") //nolint:wrapcheck // gRPC status errors should not be wrapped
	}

	user, ok := s.users[req.Id]
	if !ok {
		return nil, status.Error(codes.NotFound, fmt.Sprintf("user %s not found", req.Id)) //nolint:wrapcheck // gRPC status errors should not be wrapped
	}

	return user, nil
}

// CreateUser creates a new user.
func (s *Server) CreateUser(_ context.Context, req *pb.CreateUserRequest) (*pb.CreateUserResponse, error) {
	if req.Name == "" {
		return nil, status.Error(codes.InvalidArgument, "name is required") //nolint:wrapcheck // gRPC status errors should not be wrapped
	}

	id := fmt.Sprintf("%d", len(s.users)+1)

	return &pb.CreateUserResponse{
		Id:    id,
		Name:  req.Name,
		Email: req.Email,
	}, nil
}

// Echo returns the message and metadata.
func (s *Server) Echo(ctx context.Context, req *pb.EchoRequest) (*pb.EchoResponse, error) {
	resp := &pb.EchoResponse{
		Message:  req.Message,
		Metadata: make(map[string]string),
	}

	// Extract metadata
	if md, ok := metadata.FromIncomingContext(ctx); ok {
		for key, values := range md {
			if len(values) > 0 {
				resp.Metadata[key] = values[0]
			}
		}
	}

	return resp, nil
}

// StartTestServer starts a gRPC server for testing and returns the address.
func StartTestServer(t *testing.T) string {
	t.Helper()

	var lc net.ListenConfig

	lis, err := lc.Listen(context.Background(), "tcp", "localhost:0")
	if err != nil {
		t.Fatalf("Failed to listen: %v", err)
	}

	srv := grpc.NewServer()
	pb.RegisterTestServiceServer(srv, NewServer())

	go func() {
		_ = srv.Serve(lis)
	}()

	t.Cleanup(func() {
		srv.GracefulStop()
	})

	return lis.Addr().String()
}
