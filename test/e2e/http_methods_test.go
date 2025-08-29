package e2e_test

import (
	"testing"

	"github.com/sivchari/e2e"
	"github.com/sivchari/e2e/test/e2e/testserver"
)

func TestHTTPMethods(t *testing.T) {
	server := testserver.NewEchoServer()
	defer server.Close()

	client := e2e.New(t, e2e.Config{BaseURL: server.URL})

	tests := []struct {
		name   string
		method func(path string) *e2e.HTTPBuilder
	}{
		{"GET", client.GET},
		{"POST", client.POST},
		{"PUT", client.PUT},
		{"DELETE", client.DELETE},
		{"PATCH", client.PATCH},
		{"HEAD", client.HEAD},
		{"OPTIONS", client.OPTIONS},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.method("/test").
				Execute(t.Context()).
				ExpectStatus(200)
		})
	}
}
