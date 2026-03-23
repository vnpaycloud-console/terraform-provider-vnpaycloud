package client

import (
	"net/url"
	"reflect"
	"testing"
	"time"
)

func TestBuildQueryString(t *testing.T) {
	type listOpts struct {
		Name   string `q:"name"`
		Limit  int    `q:"limit"`
		Active bool   `q:"active"`
	}

	t.Run("basic struct fields", func(t *testing.T) {
		opts := listOpts{Name: "test", Limit: 10, Active: true}
		u, err := BuildQueryString(opts)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		q := u.Query()
		if got := q.Get("name"); got != "test" {
			t.Errorf("expected name=test, got %q", got)
		}
		if got := q.Get("limit"); got != "10" {
			t.Errorf("expected limit=10, got %q", got)
		}
		if got := q.Get("active"); got != "true" {
			t.Errorf("expected active=true, got %q", got)
		}
	})

	t.Run("pointer to struct", func(t *testing.T) {
		opts := &listOpts{Name: "ptr-test"}
		u, err := BuildQueryString(opts)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if got := u.Query().Get("name"); got != "ptr-test" {
			t.Errorf("expected name=ptr-test, got %q", got)
		}
	})

	t.Run("zero value fields omitted", func(t *testing.T) {
		opts := listOpts{Name: "only-name"}
		u, err := BuildQueryString(opts)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		q := u.Query()
		if got := q.Get("name"); got != "only-name" {
			t.Errorf("expected name=only-name, got %q", got)
		}
		if got := q.Get("limit"); got != "" {
			t.Errorf("expected limit to be omitted, got %q", got)
		}
	})

	t.Run("required field missing", func(t *testing.T) {
		type reqOpts struct {
			ID string `q:"id" required:"true"`
		}
		_, err := BuildQueryString(reqOpts{})
		if err == nil {
			t.Fatal("expected error for missing required field")
		}
	})

	t.Run("non-struct returns error", func(t *testing.T) {
		_, err := BuildQueryString("not-a-struct")
		if err == nil {
			t.Fatal("expected error for non-struct input")
		}
	})

	t.Run("slice field", func(t *testing.T) {
		type sliceOpts struct {
			Tags []string `q:"tags"`
		}
		opts := sliceOpts{Tags: []string{"a", "b"}}
		u, err := BuildQueryString(opts)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		q := u.Query()
		vals := q["tags"]
		if len(vals) != 2 || vals[0] != "a" || vals[1] != "b" {
			t.Errorf("expected tags=[a,b], got %v", vals)
		}
	})

	t.Run("slice with comma-separated format", func(t *testing.T) {
		type csvOpts struct {
			IDs []string `q:"ids" format:"comma-separated"`
		}
		opts := csvOpts{IDs: []string{"1", "2", "3"}}
		u, err := BuildQueryString(opts)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if got := u.Query().Get("ids"); got != "1,2,3" {
			t.Errorf("expected ids=1,2,3, got %q", got)
		}
	})

	t.Run("int slice", func(t *testing.T) {
		type intSliceOpts struct {
			Ports []int `q:"ports"`
		}
		opts := intSliceOpts{Ports: []int{80, 443}}
		u, err := BuildQueryString(opts)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		q := u.Query()
		vals := q["ports"]
		if len(vals) != 2 || vals[0] != "80" || vals[1] != "443" {
			t.Errorf("expected ports=[80,443], got %v", vals)
		}
	})

	t.Run("map field", func(t *testing.T) {
		type mapOpts struct {
			Labels map[string]string `q:"labels"`
		}
		opts := mapOpts{Labels: map[string]string{"env": "prod"}}
		u, err := BuildQueryString(opts)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if got := u.Query().Get("labels"); got == "" {
			t.Error("expected labels to be set")
		}
	})

	t.Run("pointer field", func(t *testing.T) {
		type ptrOpts struct {
			Name *string `q:"name"`
		}
		name := "ptr-value"
		opts := ptrOpts{Name: &name}
		u, err := BuildQueryString(opts)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if got := u.Query().Get("name"); got != "ptr-value" {
			t.Errorf("expected name=ptr-value, got %q", got)
		}
	})

	t.Run("nil pointer field omitted", func(t *testing.T) {
		type ptrOpts struct {
			Name *string `q:"name"`
		}
		opts := ptrOpts{Name: nil}
		u, err := BuildQueryString(opts)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if got := u.Query().Get("name"); got != "" {
			t.Errorf("expected name to be omitted, got %q", got)
		}
	})

	t.Run("field without q tag ignored", func(t *testing.T) {
		type mixedOpts struct {
			Name    string `q:"name"`
			Ignored string
		}
		opts := mixedOpts{Name: "test", Ignored: "should-not-appear"}
		u, err := BuildQueryString(opts)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		q := u.Query()
		if got := q.Get("name"); got != "test" {
			t.Errorf("expected name=test, got %q", got)
		}
		if raw := u.RawQuery; len(q) != 1 {
			t.Errorf("expected only 1 query param, got query: %s", raw)
		}
	})
}

// Verify BuildQueryString returns *url.URL
func TestBuildQueryStringReturnType(t *testing.T) {
	type opts struct {
		Name string `q:"name"`
	}
	result, err := BuildQueryString(opts{Name: "test"})
	if err != nil {
		t.Fatal(err)
	}
	var _ *url.URL = result // compile-time check
}

func TestIsZero(t *testing.T) {
	tests := []struct {
		name string
		val  any
		want bool
	}{
		{"zero string", "", true},
		{"non-zero string", "hello", false},
		{"zero int", 0, true},
		{"non-zero int", 42, false},
		{"zero bool", false, true},
		{"non-zero bool", true, false},
		{"nil slice", ([]string)(nil), true},
		{"empty slice", []string{}, false},
		{"nil map", (map[string]string)(nil), true},
		{"zero time", time.Time{}, true},
		{"non-zero time", time.Now(), false},
		{"zero struct", struct{ A int }{}, true},
		{"non-zero struct", struct{ A int }{A: 1}, false},
		{"zero array", [2]int{}, true},
		{"non-zero array", [2]int{1, 0}, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := isZero(reflect.ValueOf(tt.val))
			if got != tt.want {
				t.Errorf("isZero(%v) = %v, want %v", tt.val, got, tt.want)
			}
		})
	}

	t.Run("nil pointer", func(t *testing.T) {
		var p *string
		got := isZero(reflect.ValueOf(&p).Elem())
		if !got {
			t.Error("expected nil pointer to be zero")
		}
	})

	t.Run("non-nil pointer", func(t *testing.T) {
		s := "hello"
		got := isZero(reflect.ValueOf(&s))
		if got {
			t.Error("expected non-nil pointer to not be zero")
		}
	})

	t.Run("nil func", func(t *testing.T) {
		var fn func()
		got := isZero(reflect.ValueOf(&fn).Elem())
		if !got {
			t.Error("expected nil func to be zero")
		}
	})
}
