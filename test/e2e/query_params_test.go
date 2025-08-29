package e2e_test

import (
	"testing"

	"github.com/sivchari/e2e"
	"github.com/sivchari/e2e/test/e2e/testserver"
)

func TestQueryParams(t *testing.T) {
	server := testserver.NewEchoServer()
	defer server.Close()

	client := e2e.New(t, e2e.Config{BaseURL: server.URL})

	client.GET("/search").
		Query("q", "golang").
		Query("limit", "10").
		Query("offset", "20").
		Execute(t.Context()).
		ExpectStatus(200).
		ExpectJSON(map[string]interface{}{
			"method": "GET",
			"path":   "/search",
			"query": map[string]interface{}{
				"q":      []interface{}{"golang"},
				"limit":  []interface{}{"10"},
				"offset": []interface{}{"20"},
			},
		})
}
