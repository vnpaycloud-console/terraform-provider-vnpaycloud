package client

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"io"
	"net/http"

	"github.com/hashicorp/terraform-plugin-log/tflog"
)

const (
	DefaultUserAgent = "vnpaycloud-console-gophercloud/v2.0.0"
)

var applicationJSON = "application/json"

type Client struct {
	baseURL    string
	token      string
	httpClient http.Client
}

type ClientConfig struct {
	BaseURL       string
	AppCredID     string
	AppCredSecret string
}

func NewClient(ctx context.Context, cfg *ClientConfig) (*Client, error) {
	c := &Client{baseURL: cfg.BaseURL}

	if err := c.Authenticate(ctx, cfg.AppCredID, cfg.AppCredSecret); err != nil {
		return nil, err
	}

	return c, nil
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

func (client *Client) Authenticate(ctx context.Context, appCredID, appCredSecret string) error {
	type AuthPayload struct {
		Auth struct {
			Identity struct {
				Methods               []string `json:"methods"`
				ApplicationCredential struct {
					ID     string `json:"id,omitempty"`
					Name   string `json:"name,omitempty"`
					Secret string `json:"secret"`
				} `json:"application_credential"`
			} `json:"identity"`
		} `json:"auth"`
	}

	payload := AuthPayload{}
	payload.Auth.Identity.Methods = []string{"application_credential"}
	payload.Auth.Identity.ApplicationCredential.ID = appCredID
	payload.Auth.Identity.ApplicationCredential.Secret = appCredSecret

	var respBody map[string]interface{}
	opts := &RequestOpts{
		JSONBody:     payload,
		JSONResponse: &respBody,
		OkCodes:      []int{201},
		OmitHeaders:  []string{"X-Auth-Token"},
	}

	resp, err := client.Post(ctx, ApiPath.Auth, payload, &respBody, opts)
	if err != nil {
		return err
	}

	token := resp.Header.Get("X-Subject-Token")

	if token == "" {
		return errors.New("no token found in X-Subject-Token header")
	}

	client.token = token
	return nil
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
	req.Header.Set("X-Auth-Token", client.token)

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
		resp.Body.Close()
		err := ErrUnexpectedResponseCode{
			URL:            url,
			Method:         method,
			Expected:       okc,
			Actual:         resp.StatusCode,
			Body:           body,
			ResponseHeader: resp.Header,
		}
		tflog.Error(ctx, "An error occurred while executing a request.", map[string]interface{}{
			"status":          err.Actual,
			"url":             err.URL,
			"method":          err.Method,
			"body":            string(err.Body),
			"response_header": err.ResponseHeader,
		})

		return resp, err
	}

	if options.JSONResponse != nil {
		defer resp.Body.Close()

		if resp.StatusCode == http.StatusNoContent {
			_, err = io.Copy(io.Discard, resp.Body)
			return resp, err
		}

		if err := json.NewDecoder(resp.Body).Decode(options.JSONResponse); err != nil {
			return nil, err
		}
	}

	if !options.KeepResponseBody && options.JSONResponse == nil {
		defer resp.Body.Close()

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
		return []int{201, 202}
	case "PUT":
		return []int{201, 202}
	case "PATCH":
		return []int{200, 202, 204}
	case "DELETE":
		return []int{202, 204}
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
