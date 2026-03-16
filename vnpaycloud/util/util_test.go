package util

import (
	"errors"
	"fmt"
	"net/http"
	"reflect"
	"testing"
	"time"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"

	"terraform-provider-vnpaycloud/vnpaycloud/helper/client"
)

// ─── isZero ─────────────────────────────────────────────────────────

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

// ─── BuildRequestBody ───────────────────────────────────────────────

func TestBuildRequestBody_BasicStruct(t *testing.T) {
	type basicOpts struct {
		Name  string `json:"name"`
		Count int    `json:"count"`
	}

	result, err := BuildRequestBody(basicOpts{Name: "test", Count: 5}, "")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result["name"] != "test" {
		t.Errorf("expected name=test, got %v", result["name"])
	}
	if result["count"] != float64(5) {
		t.Errorf("expected count=5, got %v", result["count"])
	}
}

func TestBuildRequestBody_Parent(t *testing.T) {
	type opts struct {
		Name string `json:"name"`
	}

	result, err := BuildRequestBody(opts{Name: "wrapped"}, "server")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	inner, ok := result["server"].(map[string]any)
	if !ok {
		t.Fatalf("expected nested map under 'server', got %T", result["server"])
	}
	if inner["name"] != "wrapped" {
		t.Errorf("expected name=wrapped, got %v", inner["name"])
	}
}

func TestBuildRequestBody_Pointer(t *testing.T) {
	type opts struct {
		Name string `json:"name"`
	}

	result, err := BuildRequestBody(&opts{Name: "ptr"}, "")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result["name"] != "ptr" {
		t.Errorf("expected name=ptr, got %v", result["name"])
	}
}

func TestBuildRequestBody_RequiredTag(t *testing.T) {
	type opts struct {
		Name string `json:"name" required:"true"`
	}

	t.Run("missing required field", func(t *testing.T) {
		_, err := BuildRequestBody(opts{}, "")
		if err == nil {
			t.Fatal("expected error for missing required field")
		}
		var missing client.ErrMissingInput
		if !errors.As(err, &missing) {
			t.Fatalf("expected ErrMissingInput, got %T: %v", err, err)
		}
		if missing.Argument != "Name" {
			t.Errorf("expected Argument='Name', got %q", missing.Argument)
		}
	})

	t.Run("required field present", func(t *testing.T) {
		result, err := BuildRequestBody(opts{Name: "present"}, "")
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if result["name"] != "present" {
			t.Errorf("expected name=present, got %v", result["name"])
		}
	})
}

func TestBuildRequestBody_XorTag(t *testing.T) {
	type opts struct {
		FieldA string `json:"field_a" xor:"FieldB"`
		FieldB string `json:"field_b"`
	}

	t.Run("exactly one set - A", func(t *testing.T) {
		_, err := BuildRequestBody(opts{FieldA: "a"}, "")
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
	})

	t.Run("exactly one set - B", func(t *testing.T) {
		_, err := BuildRequestBody(opts{FieldB: "b"}, "")
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
	})

	t.Run("both set", func(t *testing.T) {
		_, err := BuildRequestBody(opts{FieldA: "a", FieldB: "b"}, "")
		if err == nil {
			t.Fatal("expected error when both xor fields are set")
		}
	})

	t.Run("neither set", func(t *testing.T) {
		_, err := BuildRequestBody(opts{}, "")
		if err == nil {
			t.Fatal("expected error when neither xor field is set")
		}
	})
}

func TestBuildRequestBody_OrTag(t *testing.T) {
	type opts struct {
		FieldA string `json:"field_a" or:"FieldB"`
		FieldB string `json:"field_b"`
	}

	t.Run("both set", func(t *testing.T) {
		_, err := BuildRequestBody(opts{FieldA: "a", FieldB: "b"}, "")
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
	})

	t.Run("only A set", func(t *testing.T) {
		_, err := BuildRequestBody(opts{FieldA: "a"}, "")
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
	})

	t.Run("only B set", func(t *testing.T) {
		_, err := BuildRequestBody(opts{FieldB: "b"}, "")
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
	})

	t.Run("neither set", func(t *testing.T) {
		_, err := BuildRequestBody(opts{}, "")
		if err == nil {
			t.Fatal("expected error when neither or field is set")
		}
	})
}

func TestBuildRequestBody_JsonDash(t *testing.T) {
	type opts struct {
		Name   string `json:"name"`
		Hidden string `json:"-"`
	}

	result, err := BuildRequestBody(opts{Name: "visible", Hidden: "secret"}, "")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if _, exists := result["Hidden"]; exists {
		t.Error("field with json:\"-\" should be excluded")
	}
	if result["name"] != "visible" {
		t.Errorf("expected name=visible, got %v", result["name"])
	}
}

func TestBuildRequestBody_Omitempty(t *testing.T) {
	type inner struct {
		Value string `json:"value"`
	}
	type opts struct {
		Name  string `json:"name"`
		Inner *inner `json:"inner,omitempty"`
	}

	t.Run("nil pointer omitted", func(t *testing.T) {
		result, err := BuildRequestBody(opts{Name: "test", Inner: nil}, "")
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if _, exists := result["inner"]; exists {
			t.Error("nil inner with omitempty should be excluded")
		}
	})

	t.Run("non-nil pointer included", func(t *testing.T) {
		result, err := BuildRequestBody(opts{Name: "test", Inner: &inner{Value: "present"}}, "")
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		innerMap, ok := result["inner"].(map[string]any)
		if !ok {
			t.Fatalf("expected inner to be map, got %T", result["inner"])
		}
		if innerMap["value"] != "present" {
			t.Errorf("expected value=present, got %v", innerMap["value"])
		}
	})
}

func TestBuildRequestBody_Slice(t *testing.T) {
	type opts struct {
		Tags []string `json:"tags"`
	}

	result, err := BuildRequestBody(opts{Tags: []string{"a", "b"}}, "")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	tags, ok := result["tags"].([]any)
	if !ok {
		t.Fatalf("expected tags to be []any, got %T", result["tags"])
	}
	if len(tags) != 2 || tags[0] != "a" || tags[1] != "b" {
		t.Errorf("expected [a b], got %v", tags)
	}
}

func TestBuildRequestBody_SliceOfStructs(t *testing.T) {
	type member struct {
		Name string `json:"name" required:"true"`
	}
	type opts struct {
		Members []member `json:"members"`
	}

	t.Run("valid slice of structs", func(t *testing.T) {
		_, err := BuildRequestBody(opts{Members: []member{{Name: "m1"}, {Name: "m2"}}}, "")
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
	})

	t.Run("invalid element in slice", func(t *testing.T) {
		_, err := BuildRequestBody(opts{Members: []member{{Name: "m1"}, {}}}, "")
		if err == nil {
			t.Fatal("expected error for missing required field in slice element")
		}
	})
}

func TestBuildRequestBody_NestedStruct(t *testing.T) {
	type network struct {
		CIDR string `json:"cidr" required:"true"`
	}
	type opts struct {
		Name    string  `json:"name"`
		Network network `json:"network"`
	}

	t.Run("valid nested struct", func(t *testing.T) {
		result, err := BuildRequestBody(opts{Name: "test", Network: network{CIDR: "10.0.0.0/8"}}, "")
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		net, ok := result["network"].(map[string]any)
		if !ok {
			t.Fatalf("expected network to be map, got %T", result["network"])
		}
		if net["cidr"] != "10.0.0.0/8" {
			t.Errorf("expected cidr=10.0.0.0/8, got %v", net["cidr"])
		}
	})

	t.Run("zero nested struct is skipped", func(t *testing.T) {
		// Zero-valued nested structs are skipped (not validated)
		result, err := BuildRequestBody(opts{Name: "test", Network: network{}}, "")
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if result["name"] != "test" {
			t.Errorf("expected name=test, got %v", result["name"])
		}
	})

	t.Run("non-zero nested struct with missing required", func(t *testing.T) {
		type innerReq struct {
			A string `json:"a" required:"true"`
			B string `json:"b"`
		}
		type outer struct {
			Name  string   `json:"name"`
			Inner innerReq `json:"inner"`
		}
		// Non-zero nested struct triggers validation
		_, err := BuildRequestBody(outer{Name: "test", Inner: innerReq{B: "has-value"}}, "")
		if err == nil {
			t.Fatal("expected error for missing required field in non-zero nested struct")
		}
	})
}

func TestBuildRequestBody_UnexportedFields(t *testing.T) {
	type opts struct {
		Name     string `json:"name"`
		internal string //nolint:unused
	}

	result, err := BuildRequestBody(opts{Name: "test"}, "")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result["name"] != "test" {
		t.Errorf("expected name=test, got %v", result["name"])
	}
}

func TestBuildRequestBody_ArrayInput(t *testing.T) {
	type item struct {
		Name string `json:"name"`
	}
	items := []item{{Name: "a"}, {Name: "b"}}

	t.Run("with parent", func(t *testing.T) {
		result, err := BuildRequestBody(items, "items")
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		arr, ok := result["items"].([]map[string]any)
		if !ok {
			t.Fatalf("expected []map[string]any, got %T", result["items"])
		}
		if len(arr) != 2 {
			t.Errorf("expected 2 items, got %d", len(arr))
		}
	})

	t.Run("without parent", func(t *testing.T) {
		_, err := BuildRequestBody(items, "")
		if err == nil {
			t.Fatal("expected error when passing slice without parent")
		}
	})
}

func TestBuildRequestBody_InvalidInput(t *testing.T) {
	_, err := BuildRequestBody("not-a-struct", "")
	if err == nil {
		t.Fatal("expected error for non-struct input")
	}
}

func TestBuildRequestBody_XorWithPointer(t *testing.T) {
	type opts struct {
		FieldA *string `json:"field_a" xor:"FieldB"`
		FieldB *string `json:"field_b"`
	}

	a := "hello"
	t.Run("pointer A set only", func(t *testing.T) {
		_, err := BuildRequestBody(opts{FieldA: &a}, "")
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
	})

	t.Run("both nil", func(t *testing.T) {
		_, err := BuildRequestBody(opts{}, "")
		if err == nil {
			t.Fatal("expected error when both pointer xor fields are nil")
		}
	})
}

func TestBuildRequestBody_OrWithPointer(t *testing.T) {
	type opts struct {
		FieldA *string `json:"field_a" or:"FieldB"`
		FieldB *string `json:"field_b"`
	}

	t.Run("both nil", func(t *testing.T) {
		_, err := BuildRequestBody(opts{}, "")
		if err == nil {
			t.Fatal("expected error when both or pointer fields are nil")
		}
	})

	b := "world"
	t.Run("B set only", func(t *testing.T) {
		_, err := BuildRequestBody(opts{FieldB: &b}, "")
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
	})
}

// ─── CheckDeleted ───────────────────────────────────────────────────

func TestCheckDeleted(t *testing.T) {
	r := &schema.Resource{
		Schema: map[string]*schema.Schema{
			"name": {Type: schema.TypeString, Optional: true},
		},
	}

	t.Run("404 clears resource ID", func(t *testing.T) {
		d := schema.TestResourceDataRaw(t, r.Schema, map[string]interface{}{"name": "test"})
		d.SetId("res-123")

		err404 := client.ErrUnexpectedResponseCode{Actual: http.StatusNotFound}
		result := CheckDeleted(d, err404, "Error reading resource")
		if result != nil {
			t.Fatalf("expected nil error, got: %v", result)
		}
		if d.Id() != "" {
			t.Errorf("expected empty ID, got %q", d.Id())
		}
	})

	t.Run("other error returned", func(t *testing.T) {
		d := schema.TestResourceDataRaw(t, r.Schema, map[string]interface{}{"name": "test"})
		d.SetId("res-456")

		err500 := client.ErrUnexpectedResponseCode{Actual: http.StatusInternalServerError}
		result := CheckDeleted(d, err500, "Error reading resource")
		if result == nil {
			t.Fatal("expected error to be returned")
		}
		if d.Id() != "res-456" {
			t.Errorf("expected ID to remain res-456, got %q", d.Id())
		}
	})
}

// ─── CheckNotFound ──────────────────────────────────────────────────

func TestCheckNotFound(t *testing.T) {
	r := &schema.Resource{
		Schema: map[string]*schema.Schema{
			"name": {Type: schema.TypeString, Optional: true},
		},
	}

	t.Run("404 clears resource ID", func(t *testing.T) {
		d := schema.TestResourceDataRaw(t, r.Schema, map[string]interface{}{"name": "test"})
		d.SetId("res-123")

		err404 := client.ErrUnexpectedResponseCode{Actual: http.StatusNotFound}
		result := CheckNotFound(d, err404, "Error reading resource")
		if result != nil {
			t.Fatalf("expected nil error, got: %v", result)
		}
		if d.Id() != "" {
			t.Errorf("expected empty ID, got %q", d.Id())
		}
	})

	t.Run("other error returned", func(t *testing.T) {
		d := schema.TestResourceDataRaw(t, r.Schema, map[string]interface{}{"name": "test"})
		d.SetId("res-789")

		err500 := client.ErrUnexpectedResponseCode{Actual: http.StatusInternalServerError}
		result := CheckNotFound(d, err500, "Error reading resource")
		if result == nil {
			t.Fatal("expected error to be returned")
		}
		if d.Id() != "res-789" {
			t.Errorf("expected ID to remain res-789, got %q", d.Id())
		}
	})

	t.Run("non-response error returned", func(t *testing.T) {
		d := schema.TestResourceDataRaw(t, r.Schema, map[string]interface{}{"name": "test"})
		d.SetId("res-abc")

		result := CheckNotFound(d, fmt.Errorf("connection refused"), "Error reading resource")
		if result == nil {
			t.Fatal("expected error to be returned")
		}
	})
}

// ─── ResponseCodeIs ─────────────────────────────────────────────────

func TestResponseCodeIs(t *testing.T) {
	tests := []struct {
		name   string
		err    error
		status int
		want   bool
	}{
		{
			name:   "matching status code",
			err:    client.ErrUnexpectedResponseCode{Actual: 404},
			status: 404,
			want:   true,
		},
		{
			name:   "non-matching status code",
			err:    client.ErrUnexpectedResponseCode{Actual: 500},
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
			err:    fmt.Errorf("wrapped: %w", client.ErrUnexpectedResponseCode{Actual: 429}),
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
