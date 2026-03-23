package client

import (
	"errors"
	"fmt"
	"net/http"
	"testing"
)

func TestBaseError(t *testing.T) {
	t.Run("default message", func(t *testing.T) {
		e := BaseError{}
		got := e.Error()
		want := "An error occurred while executing a request."
		if got != want {
			t.Errorf("expected %q, got %q", want, got)
		}
	})

	t.Run("custom info overrides default", func(t *testing.T) {
		e := BaseError{Info: "custom error info"}
		got := e.Error()
		if got != "custom error info" {
			t.Errorf("expected 'custom error info', got %q", got)
		}
	})
}

func TestErrUnexpectedResponseCode(t *testing.T) {
	t.Run("error message format", func(t *testing.T) {
		e := ErrUnexpectedResponseCode{
			URL:      "http://localhost/test",
			Method:   "GET",
			Expected: []int{200},
			Actual:   404,
			Body:     []byte("not found"),
		}
		got := e.Error()
		expected := fmt.Sprintf(
			"Expected HTTP response code %v when accessing [%s %s], but got %d instead: %s",
			[]int{200}, "GET", "http://localhost/test", 404, "not found",
		)
		if got != expected {
			t.Errorf("expected %q, got %q", expected, got)
		}
	})

	t.Run("info overrides default message", func(t *testing.T) {
		e := ErrUnexpectedResponseCode{
			URL:      "http://localhost/test",
			Method:   "GET",
			Expected: []int{200},
			Actual:   500,
			Body:     []byte("server error"),
		}
		e.Info = "custom info"
		got := e.Error()
		if got != "custom info" {
			t.Errorf("expected 'custom info', got %q", got)
		}
	})

	t.Run("GetStatusCode", func(t *testing.T) {
		e := ErrUnexpectedResponseCode{Actual: 503}
		if got := e.GetStatusCode(); got != 503 {
			t.Errorf("expected 503, got %d", got)
		}
	})

	t.Run("implements error interface", func(t *testing.T) {
		var _ error = ErrUnexpectedResponseCode{}
	})
}

func TestResponseCodeIs(t *testing.T) {
	tests := []struct {
		name   string
		err    error
		status int
		want   bool
	}{
		{
			name:   "matching status code",
			err:    ErrUnexpectedResponseCode{Actual: 404},
			status: 404,
			want:   true,
		},
		{
			name:   "non-matching status code",
			err:    ErrUnexpectedResponseCode{Actual: 500},
			status: 404,
			want:   false,
		},
		{
			name:   "different error type",
			err:    errors.New("some error"),
			status: 404,
			want:   false,
		},
		{
			name:   "nil error",
			err:    nil,
			status: 404,
			want:   false,
		},
		{
			name:   "wrapped error",
			err:    fmt.Errorf("wrapped: %w", ErrUnexpectedResponseCode{Actual: 429}),
			status: 429,
			want:   true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := ResponseCodeIs(tt.err, tt.status)
			if got != tt.want {
				t.Errorf("ResponseCodeIs() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestErrorsAsErrUnexpectedResponseCode(t *testing.T) {
	original := ErrUnexpectedResponseCode{
		URL:      "http://test.com/api",
		Method:   "POST",
		Expected: []int{200, 201},
		Actual:   422,
		Body:     []byte("unprocessable"),
		ResponseHeader: http.Header{
			"X-Request-Id": []string{"abc123"},
		},
	}

	wrapped := fmt.Errorf("request failed: %w", original)

	var target ErrUnexpectedResponseCode
	if !errors.As(wrapped, &target) {
		t.Fatal("errors.As should find ErrUnexpectedResponseCode in wrapped error")
	}
	if target.Actual != 422 {
		t.Errorf("expected Actual 422, got %d", target.Actual)
	}
	if target.URL != "http://test.com/api" {
		t.Errorf("expected URL 'http://test.com/api', got %q", target.URL)
	}
	if string(target.Body) != "unprocessable" {
		t.Errorf("expected Body 'unprocessable', got %q", string(target.Body))
	}
	if target.ResponseHeader.Get("X-Request-Id") != "abc123" {
		t.Errorf("expected X-Request-Id 'abc123', got %q", target.ResponseHeader.Get("X-Request-Id"))
	}
}

func TestErrMissingInput(t *testing.T) {
	t.Run("default message", func(t *testing.T) {
		e := ErrMissingInput{Argument: "project_id"}
		got := e.Error()
		if got != "Missing input for argument [project_id]" {
			t.Errorf("unexpected error: %q", got)
		}
	})

	t.Run("custom info overrides", func(t *testing.T) {
		e := ErrMissingInput{Argument: "project_id"}
		e.Info = "custom missing input"
		got := e.Error()
		if got != "custom missing input" {
			t.Errorf("expected 'custom missing input', got %q", got)
		}
	})
}

func TestErrInvalidInput(t *testing.T) {
	t.Run("default message", func(t *testing.T) {
		e := ErrInvalidInput{
			ErrMissingInput: ErrMissingInput{Argument: "zone_id"},
			Value:           "invalid-zone",
		}
		got := e.Error()
		if got != "Invalid input provided for argument [zone_id]: [invalid-zone]" {
			t.Errorf("unexpected error: %q", got)
		}
	})

	t.Run("custom info overrides", func(t *testing.T) {
		e := ErrInvalidInput{
			ErrMissingInput: ErrMissingInput{Argument: "zone_id"},
			Value:           "bad",
		}
		e.Info = "custom invalid input"
		got := e.Error()
		if got != "custom invalid input" {
			t.Errorf("expected 'custom invalid input', got %q", got)
		}
	})
}

func TestErrResourceNotFound(t *testing.T) {
	t.Run("default message", func(t *testing.T) {
		e := ErrResourceNotFound{Name: "my-vpc", ResourceType: "vpc"}
		got := e.Error()
		if got != "Unable to find vpc with name my-vpc" {
			t.Errorf("unexpected error: %q", got)
		}
	})

	t.Run("custom info overrides", func(t *testing.T) {
		e := ErrResourceNotFound{Name: "my-vpc", ResourceType: "vpc"}
		e.Info = "custom not found"
		got := e.Error()
		if got != "custom not found" {
			t.Errorf("expected 'custom not found', got %q", got)
		}
	})
}

func TestErrMultipleResourcesFound(t *testing.T) {
	t.Run("default message", func(t *testing.T) {
		e := ErrMultipleResourcesFound{Name: "web-server", Count: 3, ResourceType: "instance"}
		got := e.Error()
		if got != "Found 3 instances matching web-server" {
			t.Errorf("unexpected error: %q", got)
		}
	})

	t.Run("custom info overrides", func(t *testing.T) {
		e := ErrMultipleResourcesFound{Name: "web-server", Count: 3, ResourceType: "instance"}
		e.Info = "custom multiple found"
		got := e.Error()
		if got != "custom multiple found" {
			t.Errorf("expected 'custom multiple found', got %q", got)
		}
	})
}
