package client

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"math/rand/v2"
	"net"
	"net/http"
	"strings"
	"time"

	"github.com/hashicorp/terraform-plugin-log/tflog"
)

const (
	DefaultUserAgent = "terraform-provider-vnpaycloud/v2"
)

var applicationJSON = "application/json"

type Client struct {
	baseURL    string
	token      string
	httpClient http.Client
}

type ClientConfig struct {
	BaseURL string
	Token   string
}

func NewClient(_ context.Context, cfg *ClientConfig) (*Client, error) {
	if cfg.Token == "" {
		return nil, errors.New("token is required")
	}
	if cfg.BaseURL == "" {
		return nil, errors.New("base_url is required")
	}

	transport := &http.Transport{
		DialContext: (&net.Dialer{
			Timeout:   30 * time.Second,
			KeepAlive: 30 * time.Second,
		}).DialContext,
		TLSHandshakeTimeout: 10 * time.Second,
	}

	return &Client{
		baseURL: strings.TrimRight(cfg.BaseURL, "/"),
		token:   cfg.Token,
		httpClient: http.Client{
			Transport: transport,
			Timeout:   60 * time.Second,
		},
	}, nil
}

type RequestOpts struct {
	JSONBody         any
	RawBody          io.Reader
	JSONResponse     any
	OkCodes          []int
	MoreHeaders      map[string]string
	OmitHeaders      []string
	KeepResponseBody bool
}

func (client *Client) Get(ctx context.Context, path string, JSONResponse any, opts *RequestOpts) (*http.Response, error) {
	if opts == nil {
		opts = new(RequestOpts)
	}
	client.initReqOpts(nil, JSONResponse, opts)
	return client.doRequest(ctx, "GET", client.baseURL+path, opts)
}

func (client *Client) Post(ctx context.Context, path string, JSONBody any, JSONResponse any, opts *RequestOpts) (*http.Response, error) {
	if opts == nil {
		opts = new(RequestOpts)
	}
	client.initReqOpts(JSONBody, JSONResponse, opts)
	return client.doRequest(ctx, "POST", client.baseURL+path, opts)
}

func (client *Client) Put(ctx context.Context, path string, JSONBody any, JSONResponse any, opts *RequestOpts) (*http.Response, error) {
	if opts == nil {
		opts = new(RequestOpts)
	}
	client.initReqOpts(JSONBody, JSONResponse, opts)
	return client.doRequest(ctx, "PUT", client.baseURL+path, opts)
}

func (client *Client) Patch(ctx context.Context, path string, JSONBody any, JSONResponse any, opts *RequestOpts) (*http.Response, error) {
	if opts == nil {
		opts = new(RequestOpts)
	}
	client.initReqOpts(JSONBody, JSONResponse, opts)
	return client.doRequest(ctx, "PATCH", client.baseURL+path, opts)
}

func (client *Client) Delete(ctx context.Context, path string, opts *RequestOpts) (*http.Response, error) {
	if opts == nil {
		opts = new(RequestOpts)
	}
	client.initReqOpts(nil, nil, opts)
	return client.doRequest(ctx, "DELETE", client.baseURL+path, opts)
}

func (client *Client) doRequest(ctx context.Context, method, url string, options *RequestOpts) (*http.Response, error) {
	const (
		maxTransientRetries = 3 // 503 Service Unavailable (network blips)
		maxRateLimitRetries = 4 // 429 / "Too Many Requests" — longer backoff, backend rate-limits aggressively
	)

	transientAttempts := 0
	rateLimitAttempts := 0

	for {
		resp, err := client.doRequestOnce(ctx, method, url, options)
		if err == nil {
			return resp, nil
		}

		var respErr ErrUnexpectedResponseCode
		if !errors.As(err, &respErr) {
			return resp, err
		}

		isRateLimited := respErr.Actual == http.StatusTooManyRequests ||
			strings.Contains(string(respErr.Body), "Too Many Requests")
		isTransient := respErr.Actual == http.StatusServiceUnavailable && !isRateLimited

		var backoff time.Duration
		switch {
		case isRateLimited:
			if rateLimitAttempts >= maxRateLimitRetries {
				return resp, err
			}

			backoff = time.Duration(30*(rateLimitAttempts+1)) * time.Second // 30s, 60s, 90s, 120s
			backoff += time.Duration(rand.Int64N(int64(backoff) / 4))       // +0–25% jitter
			rateLimitAttempts++
		case isTransient:
			if transientAttempts >= maxTransientRetries {
				return resp, err
			}
			backoff = time.Duration(1<<uint(transientAttempts)) * time.Second // 1s, 2s, 4s
			transientAttempts++
		default:
			return resp, err
		}

		select {
		case <-ctx.Done():
			return resp, ctx.Err()
		case <-time.After(backoff):
		}
	}
}

func (client *Client) doRequestOnce(ctx context.Context, method, url string, options *RequestOpts) (*http.Response, error) {
	var body io.Reader
	var contentType *string

	if options.JSONBody != nil {
		if options.RawBody != nil {
			return nil, errors.New("please provide only one of JSONBody or RawBody to Request")
		}

		rendered, err := json.Marshal(options.JSONBody)
		if err != nil {
			return nil, err
		}

		body = bytes.NewReader(rendered)
		contentType = &applicationJSON
	}

	if options.KeepResponseBody && options.JSONResponse != nil {
		return nil, errors.New("cannot use KeepResponseBody when JSONResponse is not nil")
	}

	if options.RawBody != nil {
		body = options.RawBody
	}

	req, err := http.NewRequestWithContext(ctx, method, url, body)
	if err != nil {
		return nil, err
	}

	if contentType != nil {
		req.Header.Set("Content-Type", *contentType)
	}

	req.Header.Set("Accept", applicationJSON)
	req.Header.Set("User-Agent", DefaultUserAgent)
	req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", client.token))

	if options.MoreHeaders != nil {
		for k, v := range options.MoreHeaders {
			req.Header.Set(k, v)
		}
	}

	for _, v := range options.OmitHeaders {
		req.Header.Del(v)
	}

	resp, err := client.httpClient.Do(req)
	if err != nil {
		return nil, err
	}

	okc := options.OkCodes
	if okc == nil {
		okc = defaultOkCodes(method)
	}

	var ok bool
	for _, code := range okc {
		if resp.StatusCode == code {
			ok = true
			break
		}
	}

	if !ok {
		body, _ := io.ReadAll(resp.Body)
		_ = resp.Body.Close()
		respErr := ErrUnexpectedResponseCode{
			URL:            url,
			Method:         method,
			Expected:       okc,
			Actual:         resp.StatusCode,
			Body:           body,
			ResponseHeader: resp.Header,
		}
		respErr.Info = string(respErr.Body)

		tflog.Error(ctx, "An error occurred while executing a request.", map[string]interface{}{
			"status":          respErr.Actual,
			"url":             respErr.URL,
			"method":          respErr.Method,
			"body":            string(respErr.Body),
			"response_header": respErr.ResponseHeader,
		})

		return resp, respErr
	}

	if options.JSONResponse != nil {
		defer func() { _ = resp.Body.Close() }()

		if resp.StatusCode == http.StatusNoContent {
			_, err = io.Copy(io.Discard, resp.Body)
			return resp, err
		}

		if err := json.NewDecoder(resp.Body).Decode(options.JSONResponse); err != nil {
			return nil, err
		}
	}

	if !options.KeepResponseBody && options.JSONResponse == nil {
		defer func() { _ = resp.Body.Close() }()

		if _, err := io.Copy(io.Discard, resp.Body); err != nil {
			return nil, err
		}
	}

	return resp, nil
}

func defaultOkCodes(method string) []int {
	switch method {
	case "GET", "HEAD":
		return []int{200}
	case "POST":
		return []int{200, 201, 202}
	case "PUT":
		return []int{200, 201, 202}
	case "PATCH":
		return []int{200, 202, 204}
	case "DELETE":
		return []int{200, 202, 204}
	}

	return []int{}
}

func (client *Client) initReqOpts(JSONBody any, JSONResponse any, opts *RequestOpts) {
	if v, ok := (JSONBody).(io.Reader); ok {
		opts.RawBody = v
	} else if JSONBody != nil {
		opts.JSONBody = JSONBody
	}

	if JSONResponse != nil {
		opts.JSONResponse = JSONResponse
	}
}
