package e2e_test

import (
	"testing"

	"github.com/sivchari/e2e"
	"github.com/sivchari/e2e/test/e2e/testserver"
)

func TestHeaders(t *testing.T) {
	server := testserver.NewEchoServer()
	defer server.Close()

	client := e2e.New(t, e2e.Config{BaseURL: server.URL})

	t.Run("CustomHeaders", func(t *testing.T) {
		client.GET("/test").
			Header("X-Request-ID", "123").
			Header("X-Custom-Header", "value").
			Execute(t.Context()).
			ExpectStatus(200).
			ExpectHeader("X-Echo-X-Request-ID", "123").
			ExpectHeader("X-Echo-X-Custom-Header", "value")
	})

	t.Run("Authorization", func(t *testing.T) {
		client.GET("/test").
			Authorization("Bearer token123").
			Execute(t.Context()).
			ExpectStatus(200).
			ExpectHeader("X-Echo-Authorization", "Bearer token123")
	})
}
