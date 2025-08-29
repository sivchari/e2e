// Package testserver provides test servers for e2e testing
package testserver

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
)

// NewEchoServer creates a test server that echoes request details.
func NewEchoServer() *httptest.Server {
	return httptest.NewServer(http.HandlerFunc(echoHandler))
}

// echoHandler echoes back request information for testing.
func echoHandler(w http.ResponseWriter, r *http.Request) {
	// Echo headers back with X-Echo- prefix
	echoHeaders(w, r.Header)

	response := buildEchoResponse(r)
	sendJSONResponse(w, response)
}

// buildEchoResponse constructs the echo response from the request.
func buildEchoResponse(r *http.Request) map[string]interface{} {
	response := map[string]interface{}{
		"method": r.Method,
		"path":   r.URL.Path,
	}

	// Include query parameters if present
	if query := r.URL.Query(); len(query) > 0 {
		response["query"] = query
	}

	// Include body for non-GET/HEAD requests
	if shouldIncludeBody(r.Method) {
		if body := parseRequestBody(r); body != nil {
			response["body"] = body
		}
	}

	return response
}

// shouldIncludeBody checks if the request method should have a body.
func shouldIncludeBody(method string) bool {
	return method != http.MethodGet &&
		method != http.MethodHead &&
		method != http.MethodOptions
}

// parseRequestBody attempts to parse the request body as JSON.
func parseRequestBody(r *http.Request) interface{} {
	if r.Body == nil {
		return nil
	}

	var body interface{}
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		return nil
	}

	return body
}

// echoHeaders copies request headers to response with X-Echo- prefix.
func echoHeaders(w http.ResponseWriter, headers http.Header) {
	for key, values := range headers {
		if len(values) > 0 {
			w.Header().Set("X-Echo-"+key, values[0])
		}
	}
}

// sendJSONResponse sends the response as JSON.
func sendJSONResponse(w http.ResponseWriter, data interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	_ = json.NewEncoder(w).Encode(data)
}
