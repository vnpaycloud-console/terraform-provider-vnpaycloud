package client

import (
	"bytes"
	"errors"
	"fmt"
	"net/http"
)

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

func (e ErrUnexpectedResponseCode) Error() string {
	e.DefaultErrString = fmt.Sprintf(
		"Expected HTTP response code %v when accessing [%s %s], but got %d instead: %s",
		e.Expected, e.Method, e.URL, e.Actual, bytes.TrimSpace(e.Body),
	)
	return e.choseErrString()
}

func (e ErrUnexpectedResponseCode) GetStatusCode() int {
	return e.Actual
}

func ResponseCodeIs(err error, status int) bool {
	var codeError ErrUnexpectedResponseCode
	if errors.As(err, &codeError) {
		return codeError.Actual == status
	}
	return false
}

type ErrInvalidInput struct {
	ErrMissingInput
	Value any
}

func (e ErrInvalidInput) Error() string {
	e.DefaultErrString = fmt.Sprintf("Invalid input provided for argument [%s]: [%+v]", e.Argument, e.Value)
	return e.choseErrString()
}

type ErrMissingInput struct {
	BaseError
	Argument string
}

func (e ErrMissingInput) Error() string {
	e.DefaultErrString = fmt.Sprintf("Missing input for argument [%s]", e.Argument)
	return e.choseErrString()
}

type ErrResourceNotFound struct {
	BaseError
	Name         string
	ResourceType string
}

func (e ErrResourceNotFound) Error() string {
	e.DefaultErrString = fmt.Sprintf("Unable to find %s with name %s", e.ResourceType, e.Name)
	return e.choseErrString()
}

type ErrMultipleResourcesFound struct {
	BaseError
	Name         string
	Count        int
	ResourceType string
}

func (e ErrMultipleResourcesFound) Error() string {
	e.DefaultErrString = fmt.Sprintf("Found %d %ss matching %s", e.Count, e.ResourceType, e.Name)
	return e.choseErrString()
}
