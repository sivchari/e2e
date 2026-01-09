package e2e

import (
	"bytes"
	"context"
	"fmt"
	"testing"
	"time"

	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/encoding/prototext"
	"google.golang.org/protobuf/proto"
)

// protoString formats a proto message as a string.
func protoString(m proto.Message) string {
	return prototext.Format(m)
}

// GRPCConfig holds configuration for the gRPC test suite.
type GRPCConfig struct {
	Target  string        // gRPC server address (e.g., "localhost:9090")
	Timeout time.Duration // Default timeout for requests
}

// GRPCSuite represents the gRPC test suite.
type GRPCSuite struct {
	config GRPCConfig
	t      testing.TB
	conn   *grpc.ClientConn
}

// GRPCBuilder builds gRPC requests.
type GRPCBuilder struct {
	suite    *GRPCSuite
	method   string
	body     proto.Message
	md       metadata.MD
	timeout  time.Duration
	respErr  error
	respBody []byte
}

// NewGRPC creates a new gRPC test suite.
func NewGRPC(tb testing.TB, config GRPCConfig) *GRPCSuite {
	tb.Helper()

	if config.Timeout == 0 {
		config.Timeout = 30 * time.Second
	}

	conn, err := grpc.NewClient(
		config.Target,
		grpc.WithTransportCredentials(insecure.NewCredentials()),
	)
	if err != nil {
		tb.Fatalf("Failed to connect to gRPC server: %v", err)
	}

	tb.Cleanup(func() {
		if err := conn.Close(); err != nil {
			tb.Logf("Failed to close gRPC connection: %v", err)
		}
	})

	return &GRPCSuite{
		config: config,
		t:      tb,
		conn:   conn,
	}
}

// Call creates a gRPC call builder.
func (s *GRPCSuite) Call(method string) *GRPCBuilder {
	return &GRPCBuilder{
		suite:  s,
		method: method,
		md:     metadata.New(nil),
	}
}

// Body sets the request message.
func (b *GRPCBuilder) Body(message proto.Message) *GRPCBuilder {
	b.body = message

	return b
}

// Metadata adds gRPC metadata.
func (b *GRPCBuilder) Metadata(key, value string) *GRPCBuilder {
	b.md.Append(key, value)

	return b
}

// Timeout sets request timeout.
func (b *GRPCBuilder) Timeout(d time.Duration) *GRPCBuilder {
	b.timeout = d

	return b
}

// Execute performs the gRPC call.
func (b *GRPCBuilder) Execute(ctx context.Context) *GRPCBuilder {
	if b.timeout > 0 {
		var cancel context.CancelFunc
		ctx, cancel = context.WithTimeout(ctx, b.timeout)
		b.suite.t.Cleanup(cancel)
	}

	if len(b.md) > 0 {
		ctx = metadata.NewOutgoingContext(ctx, b.md)
	}

	// Serialize request
	reqBody, err := b.serializeBody()
	if err != nil {
		b.suite.t.Fatalf("Failed to marshal request: %v", err)
	}

	// Execute RPC
	var respBody []byte

	err = b.suite.conn.Invoke(
		ctx,
		"/"+b.method,
		reqBody,
		&respBody,
		grpc.ForceCodec(rawCodec{}),
	)

	b.respErr = err
	b.respBody = respBody

	return b
}

// ExpectCode validates the gRPC status code.
func (b *GRPCBuilder) ExpectCode(code codes.Code) *GRPCBuilder {
	st := status.Convert(b.respErr)
	if st.Code() != code {
		b.suite.t.Fatal(b.formatError(
			"gRPC status code mismatch",
			code.String(),
			st.Code().String(),
		))
	}

	return b
}

// ExpectMessage validates the response message.
func (b *GRPCBuilder) ExpectMessage(expected proto.Message) *GRPCBuilder {
	if b.respErr != nil {
		st := status.Convert(b.respErr)
		b.suite.t.Fatalf("Cannot validate message: RPC failed with %s: %s", st.Code(), st.Message())
	}

	// Create a new instance of the same type
	actual := proto.Clone(expected)
	proto.Reset(actual)

	if err := proto.Unmarshal(b.respBody, actual); err != nil {
		b.suite.t.Fatalf("Failed to unmarshal response: %v", err)
	}

	if !proto.Equal(expected, actual) {
		b.suite.t.Fatal(b.formatError(
			"gRPC response message mismatch",
			protoString(expected),
			protoString(actual),
		))
	}

	return b
}

// ExpectErrorMessage validates the error message.
func (b *GRPCBuilder) ExpectErrorMessage(msg string) *GRPCBuilder {
	st := status.Convert(b.respErr)
	if st.Message() != msg {
		b.suite.t.Fatal(b.formatError(
			"gRPC error message mismatch",
			msg,
			st.Message(),
		))
	}

	return b
}

// formatError creates a detailed error message.
func (b *GRPCBuilder) formatError(assertion, expected, actual string) string {
	var sb bytes.Buffer

	sb.WriteString("\n=== gRPC Request Failed ===\n")
	sb.WriteString(fmt.Sprintf("Method:   %s\n", b.method))
	sb.WriteString(fmt.Sprintf("Target:   %s\n", b.suite.config.Target))

	if b.body != nil {
		sb.WriteString(fmt.Sprintf("Request:  %s\n", protoString(b.body)))
	}

	if b.respErr != nil {
		st := status.Convert(b.respErr)
		sb.WriteString(fmt.Sprintf("\nResponse: %s\n", st.Code()))
		sb.WriteString(fmt.Sprintf("Message:  %s\n", st.Message()))
	}

	sb.WriteString("\n")
	sb.WriteString(fmt.Sprintf("%s:\n", assertion))
	sb.WriteString(fmt.Sprintf("Expected: %s\n", expected))
	sb.WriteString(fmt.Sprintf("Actual:   %s\n", actual))

	return sb.String()
}

// rawCodec is a codec that passes through raw bytes.
type rawCodec struct{}

func (rawCodec) Marshal(v interface{}) ([]byte, error) {
	if b, ok := v.([]byte); ok {
		return b, nil
	}

	return nil, fmt.Errorf("rawCodec: expected []byte, got %T", v)
}

func (rawCodec) Unmarshal(data []byte, v interface{}) error {
	if b, ok := v.(*[]byte); ok {
		*b = data

		return nil
	}

	return fmt.Errorf("rawCodec: expected *[]byte, got %T", v)
}

func (rawCodec) Name() string {
	return "raw"
}

// serializeBody serializes the request body to bytes.
func (b *GRPCBuilder) serializeBody() ([]byte, error) {
	if b.body == nil {
		return nil, nil
	}

	data, err := proto.Marshal(b.body)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal proto message: %w", err)
	}

	return data, nil
}
