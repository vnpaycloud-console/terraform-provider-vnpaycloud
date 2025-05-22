package client

import "net/http"

type BaseError struct {
	DefaultErrString string
	Info             string
}

func (e BaseError) Error() string {
	e.DefaultErrString = "An error occurred while executing a request."
	return e.choseErrString()
}

func (e BaseError) choseErrString() string {
	if e.Info != "" {
		return e.Info
	}
	return e.DefaultErrString
}

type ErrUnexpectedResponseCode struct {
	BaseError
	URL            string
	Method         string
	Expected       []int
	Actual         int
	Body           []byte
	ResponseHeader http.Header
}
