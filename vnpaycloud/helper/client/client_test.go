package client

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"sync/atomic"
	"testing"
	"time"
)

func TestNewClient(t *testing.T) {
	tests := []struct {
		name    string
		cfg     *ClientConfig
		wantErr string
	}{
		{
			name:    "missing token",
			cfg:     &ClientConfig{BaseURL: "http://localhost"},
			wantErr: "token is required",
		},
		{
			name:    "missing base_url",
			cfg:     &ClientConfig{Token: "vtx_pat_xxx"},
			wantErr: "base_url is required",
		},
		{
			name: "valid config",
			cfg:  &ClientConfig{BaseURL: "http://localhost", Token: "vtx_pat_xxx"},
		},
		{
			name: "trailing slash trimmed",
			cfg:  &ClientConfig{BaseURL: "http://localhost/", Token: "vtx_pat_xxx"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			c, err := NewClient(context.Background(), tt.cfg)
			if tt.wantErr != "" {
				if err == nil || err.Error() != tt.wantErr {
					t.Fatalf("expected error %q, got %v", tt.wantErr, err)
				}
				return
			}
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if c == nil {
				t.Fatal("expected non-nil client")
			}
			if strings.HasSuffix(c.baseURL, "/") {
				t.Error("baseURL should not have trailing slash")
			}
		})
	}
}

func newTestClient(t *testing.T, serverURL string) *Client {
	t.Helper()
	c, err := NewClient(context.Background(), &ClientConfig{
		BaseURL: serverURL,
		Token:   "test-token",
	})
	if err != nil {
		t.Fatalf("failed to create client: %v", err)
	}
	return c
}

func TestHTTPMethods(t *testing.T) {
	tests := []struct {
		name       string
		callMethod func(ctx context.Context, c *Client, path string) (*http.Response, error)
		wantMethod string
	}{
		{
			name: "GET",
			callMethod: func(ctx context.Context, c *Client, path string) (*http.Response, error) {
				return c.Get(ctx, path, nil, nil)
			},
			wantMethod: "GET",
		},
		{
			name: "POST",
			callMethod: func(ctx context.Context, c *Client, path string) (*http.Response, error) {
				return c.Post(ctx, path, nil, nil, nil)
			},
			wantMethod: "POST",
		},
		{
			name: "PUT",
			callMethod: func(ctx context.Context, c *Client, path string) (*http.Response, error) {
				return c.Put(ctx, path, nil, nil, nil)
			},
			wantMethod: "PUT",
		},
		{
			name: "PATCH",
			callMethod: func(ctx context.Context, c *Client, path string) (*http.Response, error) {
				return c.Patch(ctx, path, nil, nil, nil)
			},
			wantMethod: "PATCH",
		},
		{
			name: "DELETE",
			callMethod: func(ctx context.Context, c *Client, path string) (*http.Response, error) {
				return c.Delete(ctx, path, nil)
			},
			wantMethod: "DELETE",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var gotMethod string
			var gotPath string
			srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				gotMethod = r.Method
				gotPath = r.URL.Path
				w.WriteHeader(http.StatusOK)
			}))
			defer srv.Close()

			c := newTestClient(t, srv.URL)
			_, err := tt.callMethod(context.Background(), c, "/test-path")
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if gotMethod != tt.wantMethod {
				t.Errorf("expected method %s, got %s", tt.wantMethod, gotMethod)
			}
			if gotPath != "/test-path" {
				t.Errorf("expected path /test-path, got %s", gotPath)
			}
		})
	}
}

func TestRequestHeaders(t *testing.T) {
	var gotHeaders http.Header
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotHeaders = r.Header.Clone()
		w.WriteHeader(http.StatusOK)
	}))
	defer srv.Close()

	c := newTestClient(t, srv.URL)

	t.Run("default headers", func(t *testing.T) {
		_, err := c.Get(context.Background(), "/headers", nil, nil)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if got := gotHeaders.Get("Authorization"); got != "Bearer test-token" {
			t.Errorf("expected Authorization 'Bearer test-token', got %q", got)
		}
		if got := gotHeaders.Get("User-Agent"); got != DefaultUserAgent {
			t.Errorf("expected User-Agent %q, got %q", DefaultUserAgent, got)
		}
		if got := gotHeaders.Get("Accept"); got != "application/json" {
			t.Errorf("expected Accept 'application/json', got %q", got)
		}
	})

	t.Run("content-type set for JSON body", func(t *testing.T) {
		body := map[string]string{"key": "value"}
		_, err := c.Post(context.Background(), "/headers", body, nil, nil)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if got := gotHeaders.Get("Content-Type"); got != "application/json" {
			t.Errorf("expected Content-Type 'application/json', got %q", got)
		}
	})

	t.Run("more headers", func(t *testing.T) {
		opts := &RequestOpts{
			MoreHeaders: map[string]string{"X-Custom": "test-value"},
		}
		_, err := c.Get(context.Background(), "/headers", nil, opts)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if got := gotHeaders.Get("X-Custom"); got != "test-value" {
			t.Errorf("expected X-Custom 'test-value', got %q", got)
		}
	})

	t.Run("omit headers", func(t *testing.T) {
		opts := &RequestOpts{
			OmitHeaders: []string{"Accept"},
		}
		_, err := c.Get(context.Background(), "/headers", nil, opts)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if got := gotHeaders.Get("Accept"); got != "" {
			t.Errorf("expected Accept to be omitted, got %q", got)
		}
	})
}

func TestJSONBodySerialization(t *testing.T) {
	type reqBody struct {
		Name  string `json:"name"`
		Count int    `json:"count"`
	}

	var gotBody reqBody
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if err := json.NewDecoder(r.Body).Decode(&gotBody); err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}
		w.WriteHeader(http.StatusOK)
	}))
	defer srv.Close()

	c := newTestClient(t, srv.URL)
	input := reqBody{Name: "test", Count: 42}
	_, err := c.Post(context.Background(), "/body", input, nil, nil)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if gotBody.Name != "test" || gotBody.Count != 42 {
		t.Errorf("expected {test 42}, got %+v", gotBody)
	}
}

func TestRawBody(t *testing.T) {
	var gotBody string
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		b, _ := io.ReadAll(r.Body)
		gotBody = string(b)
		w.WriteHeader(http.StatusOK)
	}))
	defer srv.Close()

	c := newTestClient(t, srv.URL)
	raw := strings.NewReader("raw-data")
	// Pass io.Reader as JSONBody — initReqOpts should detect and use RawBody
	_, err := c.Post(context.Background(), "/raw", raw, nil, nil)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if gotBody != "raw-data" {
		t.Errorf("expected 'raw-data', got %q", gotBody)
	}
}

func TestJSONBodyAndRawBodyConflict(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))
	defer srv.Close()

	c := newTestClient(t, srv.URL)
	opts := &RequestOpts{
		JSONBody: map[string]string{"key": "value"},
		RawBody:  strings.NewReader("raw"),
	}
	_, err := c.Post(context.Background(), "/conflict", nil, nil, opts)
	if err == nil {
		t.Fatal("expected error when both JSONBody and RawBody are set")
	}
	if !strings.Contains(err.Error(), "please provide only one of JSONBody or RawBody") {
		t.Errorf("unexpected error: %v", err)
	}
}

func TestKeepResponseBodyAndJSONResponseConflict(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))
	defer srv.Close()

	c := newTestClient(t, srv.URL)
	var resp map[string]string
	opts := &RequestOpts{
		JSONResponse:     &resp,
		KeepResponseBody: true,
	}
	_, err := c.Get(context.Background(), "/conflict", nil, opts)
	if err == nil {
		t.Fatal("expected error when KeepResponseBody and JSONResponse are both set")
	}
	if !strings.Contains(err.Error(), "cannot use KeepResponseBody") {
		t.Errorf("unexpected error: %v", err)
	}
}

func TestJSONResponseDeserialization(t *testing.T) {
	type respBody struct {
		ID   string `json:"id"`
		Name string `json:"name"`
	}

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(respBody{ID: "abc", Name: "test-resource"})
	}))
	defer srv.Close()

	c := newTestClient(t, srv.URL)
	var got respBody
	_, err := c.Get(context.Background(), "/resource", &got, nil)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if got.ID != "abc" || got.Name != "test-resource" {
		t.Errorf("expected {abc test-resource}, got %+v", got)
	}
}

func TestNoContentResponse(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusNoContent)
	}))
	defer srv.Close()

	c := newTestClient(t, srv.URL)
	var got map[string]string
	resp, err := c.Delete(context.Background(), "/resource", &RequestOpts{
		JSONResponse: &got,
		OkCodes:      []int{204},
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if resp.StatusCode != http.StatusNoContent {
		t.Errorf("expected 204, got %d", resp.StatusCode)
	}
}

func TestKeepResponseBody(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{"data":"keep-me"}`))
	}))
	defer srv.Close()

	c := newTestClient(t, srv.URL)
	resp, err := c.Get(context.Background(), "/keep", nil, &RequestOpts{
		KeepResponseBody: true,
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	defer resp.Body.Close()
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		t.Fatalf("failed to read body: %v", err)
	}
	if string(body) != `{"data":"keep-me"}` {
		t.Errorf("expected body to be kept, got %q", string(body))
	}
}

func TestNilBody(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		body, _ := io.ReadAll(r.Body)
		if len(body) != 0 {
			http.Error(w, "expected empty body", http.StatusBadRequest)
			return
		}
		w.WriteHeader(http.StatusOK)
	}))
	defer srv.Close()

	c := newTestClient(t, srv.URL)
	resp, err := c.Get(context.Background(), "/nil-body", nil, nil)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if resp.StatusCode != http.StatusOK {
		t.Errorf("expected 200, got %d", resp.StatusCode)
	}
}

func TestDefaultOkCodes(t *testing.T) {
	tests := []struct {
		method string
		want   []int
	}{
		{"GET", []int{200}},
		{"HEAD", []int{200}},
		{"POST", []int{200, 201, 202}},
		{"PUT", []int{200, 201, 202}},
		{"PATCH", []int{200, 202, 204}},
		{"DELETE", []int{200, 202, 204}},
		{"OPTIONS", []int{}},
	}

	for _, tt := range tests {
		t.Run(tt.method, func(t *testing.T) {
			got := defaultOkCodes(tt.method)
			if len(got) != len(tt.want) {
				t.Fatalf("expected %v, got %v", tt.want, got)
			}
			for i, code := range got {
				if code != tt.want[i] {
					t.Errorf("code[%d]: expected %d, got %d", i, tt.want[i], code)
				}
			}
		})
	}
}

func TestCustomOkCodes(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusAccepted) // 202
	}))
	defer srv.Close()

	c := newTestClient(t, srv.URL)

	// 202 is not in default GET ok codes (only 200)
	_, err := c.Get(context.Background(), "/custom", nil, nil)
	if err == nil {
		t.Fatal("expected error for 202 with default GET ok codes")
	}

	// But with custom ok codes it should succeed
	_, err = c.Get(context.Background(), "/custom", nil, &RequestOpts{
		OkCodes: []int{202},
	})
	if err != nil {
		t.Fatalf("unexpected error with custom OkCodes: %v", err)
	}
}

func TestUnexpectedResponseCode(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusNotFound)
		w.Write([]byte(`{"error":"not found"}`))
	}))
	defer srv.Close()

	c := newTestClient(t, srv.URL)
	_, err := c.Get(context.Background(), "/missing", nil, nil)
	if err == nil {
		t.Fatal("expected error for 404")
	}

	var respErr ErrUnexpectedResponseCode
	if !errors.As(err, &respErr) {
		t.Fatalf("expected ErrUnexpectedResponseCode, got %T: %v", err, err)
	}
	if respErr.Actual != 404 {
		t.Errorf("expected status 404, got %d", respErr.Actual)
	}
	if !strings.Contains(string(respErr.Body), "not found") {
		t.Errorf("expected body to contain 'not found', got %q", string(respErr.Body))
	}
}

func TestRetryOn429(t *testing.T) {
	var attempts int32
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		n := atomic.AddInt32(&attempts, 1)
		if n <= 2 {
			w.WriteHeader(http.StatusTooManyRequests)
			w.Write([]byte("rate limited"))
			return
		}
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{"ok":true}`))
	}))
	defer srv.Close()

	c := newTestClient(t, srv.URL)
	_, err := c.Get(context.Background(), "/retry", nil, nil)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if got := atomic.LoadInt32(&attempts); got != 3 {
		t.Errorf("expected 3 attempts, got %d", got)
	}
}

func TestRetryOn503(t *testing.T) {
	var attempts int32
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		n := atomic.AddInt32(&attempts, 1)
		if n <= 1 {
			w.WriteHeader(http.StatusServiceUnavailable)
			w.Write([]byte("service unavailable"))
			return
		}
		w.WriteHeader(http.StatusOK)
	}))
	defer srv.Close()

	c := newTestClient(t, srv.URL)
	_, err := c.Get(context.Background(), "/retry503", nil, nil)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if got := atomic.LoadInt32(&attempts); got != 2 {
		t.Errorf("expected 2 attempts, got %d", got)
	}
}

func TestRetryOnBodyContainsTooManyRequests(t *testing.T) {
	var attempts int32
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		n := atomic.AddInt32(&attempts, 1)
		if n <= 1 {
			// Simulate gRPC error forwarded as non-429 status but with "Too Many Requests" in body
			w.WriteHeader(http.StatusBadRequest)
			w.Write([]byte(`{"error":"Too Many Requests"}`))
			return
		}
		w.WriteHeader(http.StatusOK)
	}))
	defer srv.Close()

	c := newTestClient(t, srv.URL)
	_, err := c.Get(context.Background(), "/grpc-rate-limit", nil, nil)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if got := atomic.LoadInt32(&attempts); got != 2 {
		t.Errorf("expected 2 attempts, got %d", got)
	}
}

func TestNoRetryOnOtherErrors(t *testing.T) {
	statusCodes := []int{400, 401, 403, 404, 500}
	for _, code := range statusCodes {
		t.Run(http.StatusText(code), func(t *testing.T) {
			var attempts int32
			srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				atomic.AddInt32(&attempts, 1)
				w.WriteHeader(code)
				w.Write([]byte("error"))
			}))
			defer srv.Close()

			c := newTestClient(t, srv.URL)
			_, err := c.Get(context.Background(), "/no-retry", nil, nil)
			if err == nil {
				t.Fatal("expected error")
			}
			if got := atomic.LoadInt32(&attempts); got != 1 {
				t.Errorf("expected 1 attempt (no retry), got %d", got)
			}
		})
	}
}

func TestMaxRetriesExhausted(t *testing.T) {
	var attempts int32
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		atomic.AddInt32(&attempts, 1)
		w.WriteHeader(http.StatusTooManyRequests)
		w.Write([]byte("rate limited"))
	}))
	defer srv.Close()

	c := newTestClient(t, srv.URL)
	_, err := c.Get(context.Background(), "/max-retry", nil, nil)
	if err == nil {
		t.Fatal("expected error after max retries")
	}
	// Initial attempt + 3 retries = 4 total
	if got := atomic.LoadInt32(&attempts); got != 4 {
		t.Errorf("expected 4 attempts (1 + 3 retries), got %d", got)
	}
}

func TestContextCancellationStopsRetry(t *testing.T) {
	var attempts int32
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		atomic.AddInt32(&attempts, 1)
		w.WriteHeader(http.StatusTooManyRequests)
		w.Write([]byte("rate limited"))
	}))
	defer srv.Close()

	c := newTestClient(t, srv.URL)
	ctx, cancel := context.WithTimeout(context.Background(), 500*time.Millisecond)
	defer cancel()

	_, err := c.Get(ctx, "/cancel-retry", nil, nil)
	if err == nil {
		t.Fatal("expected error on context cancellation")
	}
	if !errors.Is(err, context.DeadlineExceeded) {
		// Could also be the rate limit error if context expires between attempts
		var respErr ErrUnexpectedResponseCode
		if !errors.As(err, &respErr) {
			t.Fatalf("expected context or response error, got: %v", err)
		}
	}
	// Should have fewer than 4 attempts due to cancellation
	if got := atomic.LoadInt32(&attempts); got >= 4 {
		t.Errorf("expected fewer than 4 attempts due to cancellation, got %d", got)
	}
}

func TestInitReqOptsWithIOReader(t *testing.T) {
	c := &Client{}
	reader := bytes.NewReader([]byte("test"))
	opts := &RequestOpts{}
	c.initReqOpts(reader, nil, opts)

	if opts.RawBody == nil {
		t.Error("expected RawBody to be set when JSONBody is an io.Reader")
	}
	if opts.JSONBody != nil {
		t.Error("expected JSONBody to be nil when io.Reader is passed")
	}
}

func TestInitReqOptsWithStruct(t *testing.T) {
	c := &Client{}
	body := map[string]string{"key": "val"}
	var resp map[string]string
	opts := &RequestOpts{}
	c.initReqOpts(body, &resp, opts)

	if opts.JSONBody == nil {
		t.Error("expected JSONBody to be set")
	}
	if opts.JSONResponse == nil {
		t.Error("expected JSONResponse to be set")
	}
	if opts.RawBody != nil {
		t.Error("expected RawBody to be nil")
	}
}
