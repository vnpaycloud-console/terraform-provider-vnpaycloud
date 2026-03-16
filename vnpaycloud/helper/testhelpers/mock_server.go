package testhelpers

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
)

// Route defines a single API route handler for the mock server.
type Route struct {
	Method  string
	Pattern string
	Handler http.HandlerFunc
}

// NewMockServer creates an httptest.Server with the given route handlers.
// Routes are matched by method and path prefix. If no route matches,
// the server returns 404. The server is automatically closed when the
// test completes.
func NewMockServer(t *testing.T, routes []Route) *httptest.Server {
	t.Helper()

	mux := http.NewServeMux()
	for _, r := range routes {
		route := r // capture
		mux.HandleFunc(route.Pattern, func(w http.ResponseWriter, req *http.Request) {
			if route.Method != "" && req.Method != route.Method {
				http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
				return
			}
			route.Handler(w, req)
		})
	}

	srv := httptest.NewServer(mux)
	t.Cleanup(srv.Close)
	return srv
}

// JSONHandler returns an http.HandlerFunc that responds with the given
// status code and JSON-encoded body.
func JSONHandler(t *testing.T, statusCode int, body any) http.HandlerFunc {
	t.Helper()
	return func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(statusCode)
		if body != nil {
			if err := json.NewEncoder(w).Encode(body); err != nil {
				t.Errorf("failed to encode response body: %v", err)
			}
		}
	}
}

// EmptyHandler returns an http.HandlerFunc that responds with the given
// status code and no body (e.g. 204 No Content, 202 Accepted).
func EmptyHandler(statusCode int) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(statusCode)
	}
}
