package http

import (
	"bytes"
	"context"
	"crypto/tls"
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/url"
	"reflect"
	"strconv"
	"strings"
	"time"

	"github.com/golang/glog"
	"github.com/hashicorp/terraform-plugin-log/tflog"
	"google.golang.org/genproto/googleapis/api/httpbody"
	"google.golang.org/grpc"
	"google.golang.org/grpc/metadata"
	"google.golang.org/protobuf/types/known/anypb"
)

func BuildURL(endpoint, path string, requestStruct interface{}) (string, error) {
	// Let's start with a base url

	baseUrl, err := url.Parse(endpoint)
	if err != nil {
		return "", errors.New(fmt.Sprintf("BuildURL err = %s", err))
	}

	// Add a Path Segment (Path segment is automatically escaped)
	baseUrl.Path += path

	// Add Query Parameters to the URL
	params, err := getParameterURLfromRequestStruct(requestStruct)
	if err != nil {
		return "", errors.New(fmt.Sprintf("BuildURL err = %s", err))
	}

	if params != "" {
		baseUrl.RawQuery = params
	}

	return baseUrl.String(), nil
}

func SendRequest(ctx context.Context, url, method string, body interface{}) (*httpbody.HttpBody, error) {
	md, _ := metadata.FromIncomingContext(ctx)

	header := http.Header{}
	for k, v := range md {
		if strings.EqualFold(k, "authorization") {
			header.Add(k, strings.Join(v, "; "))
		}
	}

	httpresp := &http.Response{}
	_, err := SendHttp(ctx, url,
		WithBody(body),
		WithMethod(method),
		WithHeader(header),
		WithFullResponse(&httpresp),
		WithSkipTls(),
		WithEnableContext(),
	)
	if err != nil {
		tflog.Error(ctx, "http.SendRequest err =", map[string]interface{}{"err": err})
		return nil, errors.New(fmt.Sprintf("http.SendRequest err = %s", err))
	}

	//forward httpresp header
	metad := metadata.Pairs()

	for names, values := range httpresp.Header {
		metad.Append(names, strings.Join(values, "; "))
	}
	statusCode := strconv.Itoa(httpresp.StatusCode)
	metad.Append("x-http-code", statusCode)

	err = grpc.SendHeader(ctx, metad)
	if err != nil {
		tflog.Error(ctx, "http.SendRequest err =", map[string]interface{}{"err": err})
		return nil, errors.New(fmt.Sprintf("http.SendRequest err = %s", err))
	}

	contentType := httpresp.Header.Get("Content-Type")
	if contentType == "" {
		contentType = "application/json"
	}

	data, err := ioutil.ReadAll(httpresp.Body)
	if err != nil {
		tflog.Error(ctx, "http.SendRequest err =", map[string]interface{}{"err": err})
		return nil, errors.New(fmt.Sprintf("http.SendRequest err = %s", err))
	}

	resp := &httpbody.HttpBody{
		ContentType: contentType,
		Data:        data,
		Extensions: []*anypb.Any{
			{
				TypeUrl: "StatusCode",
				Value:   []byte(statusCode),
			},
		},
	}

	return resp, nil
}

func getParameterURLfromRequestStruct(requestStruct interface{}) (string, error) {
	// Prepare Query Parameters
	if requestStruct == nil {
		return "", nil
	}
	typeOfRequest := reflect.TypeOf(requestStruct)
	valueOfRequest := reflect.ValueOf(requestStruct)

	switch typeOfRequest.Kind() {
	case reflect.Ptr:
		if typeOfRequest.Elem().Kind() != reflect.Struct {
			return "", errors.New(fmt.Sprintf("getParameterURLfromRequestStruct requestStruct is not a pointer to struct, requestStruct = %v", requestStruct))
		}
		valueOfRequest = valueOfRequest.Elem()
	case reflect.Struct:
		break
	default:
		return "", errors.New(fmt.Sprintf("getParameterURLfromRequestStruct requestStruct is not struct, requestStruct = %v", requestStruct))
	}

	typeOfS := valueOfRequest.Type()

	values := make([]string, 0)
	fields := make([]string, 0)

	for i := 0; i < valueOfRequest.NumField(); i++ {
		valueField := valueOfRequest.Field(i)
		typeField := typeOfS.Field(i)
		tag := typeField.Tag.Get("url")

		if valueField.CanInterface() && tag == "param" {

			str := ""

			switch valueField.Kind() {
			case reflect.Float64:
				num := valueField.Interface().(float64)
				str = fmt.Sprintf("%.3f", num)
			case reflect.Float32:
				num := valueField.Interface().(float32)
				str = fmt.Sprintf("%.3f", num)
			case reflect.Int64:
				num := valueField.Interface().(int64)
				str = fmt.Sprintf("%d", num)
			case reflect.Int32:
				num := valueField.Interface().(int32)
				str = fmt.Sprintf("%d", num)
			case reflect.String:
				str = valueField.String()
			default:
				break
			}

			if str != "" {
				values = append(values, str)
				fields = append(fields, strings.ToLower(typeOfS.Field(i).Name))
			}
		}
	}

	params := url.Values{}

	for idx, value := range values {
		params.Add(fields[idx], value)
	}

	return params.Encode(), nil
}

func GetHttpClient(ctx context.Context) *http.Client {
	return &http.Client{}
}

func SendHttp(ctx context.Context, endpoint string, opts ...RequestOption) ([]byte, error) {
	client := GetHttpClient(ctx)

	reqOpt := defaultRequest()
	for _, option := range opts {
		option(reqOpt)
	}

	if reqOpt.query != nil {
		if strings.HasSuffix(endpoint, "?") {
			endpoint += reqOpt.query.Encode()
		} else {
			endpoint += "?" + reqOpt.query.Encode()
		}
	}

	request, err := http.NewRequest(reqOpt.method, endpoint, bytes.NewBuffer(reqOpt.bodyBuffer))
	if err != nil {
		return nil, err
	}

	if reqOpt.enableCtx {
		request, err = http.NewRequestWithContext(ctx, reqOpt.method, endpoint, bytes.NewBuffer(reqOpt.bodyBuffer))
		if err != nil {
			return nil, err
		}
	}

	if reqOpt.timeout.Nanoseconds() > 0 {
		client = &http.Client{
			Timeout: reqOpt.timeout,
		}
	}

	if reqOpt.httpProxy != "" {
		proxyUrl, err := url.Parse(reqOpt.httpProxy)
		if err != nil {
			return nil, err
		}
		client.Transport = &http.Transport{Proxy: http.ProxyURL(proxyUrl)}
	}

	if reqOpt.skipTls == true {
		customTransport := http.DefaultTransport.(*http.Transport).Clone()
		if client.Transport != nil {
			customTransport = client.Transport.(*http.Transport).Clone()
		}
		customTransport.TLSClientConfig = &tls.Config{InsecureSkipVerify: true}
		client.Transport = customTransport
	}

	request.Header = reqOpt.header
	if len(reqOpt.contentType) > 0 {
		request.Header.Set("Content-Type", string(reqOpt.contentType))
	}

	response, err := client.Do(request)
	if err != nil {
		glog.Errorf("SendHttp: endpoint: %s, error: %v", endpoint, err)
		return nil, err
	}

	if reqOpt.httpResponse != nil {
		*reqOpt.httpResponse = response
		return []byte{}, nil
	}

	defer func() {
		if response.Body != nil {
			response.Body.Close()
		}
	}()

	if response.StatusCode != http.StatusOK {
		msg := ""
		if response.Body != nil {
			body, _ := ioutil.ReadAll(response.Body)
			msg = string(body)
		}

		glog.Errorf("SendHttp: request: %s, response status: %s, msg: %s\n", endpoint, response.Status, msg)
		return nil, errors.New(fmt.Sprintf("SendHttp: request: %s, response status: %s, msg: %s\n", endpoint, response.Status, msg))
	}

	body, err := ioutil.ReadAll(response.Body)
	if err != nil {
		glog.Errorln("SendHttp: read all data from body error: ", err)
		return nil, err
	}

	return body, nil
}

type requestOptions struct {
	query        *url.Values
	header       http.Header
	method       string
	contentType  string
	httpProxy    string
	bodyBuffer   []byte
	httpResponse **http.Response
	timeout      time.Duration
	enableCtx    bool
	skipTls      bool
}

type RequestOption func(*requestOptions)

func defaultRequest() *requestOptions {
	return &requestOptions{
		method: http.MethodGet,
		header: http.Header{},
	}
}

func WithMethod(method string) RequestOption {
	return func(opt *requestOptions) {
		opt.method = method
	}
}

func WithContentType(contentType string) RequestOption {
	return func(opt *requestOptions) {
		opt.contentType = contentType
	}
}

func WithHeader(header http.Header) RequestOption {
	return func(opt *requestOptions) {
		opt.header = header
	}
}

func WithAddQuery(key, value string) RequestOption {
	return func(opt *requestOptions) {
		if opt.query == nil {
			opt.query = &url.Values{}
		}

		opt.query.Add(key, value)
	}
}

func WithBody(body interface{}) RequestOption {
	return func(opt *requestOptions) {
		if str, ok := body.([]byte); ok {
			opt.bodyBuffer = []byte(str)
			return
		}

		if str, ok := body.(string); ok {
			opt.bodyBuffer = []byte(str)
			return
		}

		if urlEncoded, ok := body.(interface{ Encode() string }); ok {
			opt.bodyBuffer = []byte(urlEncoded.Encode())
			return
		}

		rawData, err := json.Marshal(body)
		if err != nil {
			glog.Errorln("http.WithBody: marshal body error: ", err)
			return
		}

		opt.bodyBuffer = rawData
	}
}

func WithFullResponse(resp **http.Response) RequestOption {
	return func(opt *requestOptions) {
		opt.httpResponse = resp
	}
}

func WithHttpProxy(httpProxy string) RequestOption {
	return func(opt *requestOptions) {
		opt.httpProxy = httpProxy
	}
}

func WithEnableContext() RequestOption {
	return func(opt *requestOptions) {
		opt.enableCtx = true
	}
}

func WithSkipTls() RequestOption {
	return func(opt *requestOptions) {
		opt.skipTls = true
	}
}
