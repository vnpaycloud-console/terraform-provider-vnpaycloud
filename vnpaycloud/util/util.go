package util

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"reflect"
	"strings"
	"time"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"

	"terraform-provider-vnpaycloud/vnpaycloud/helper/client"
)

var t time.Time

func isZero(v reflect.Value) bool {
	switch v.Kind() {
	case reflect.Ptr:
		if v.IsNil() {
			return true
		}
		return false
	case reflect.Func, reflect.Map, reflect.Slice:
		return v.IsNil()
	case reflect.Array:
		z := true
		for i := 0; i < v.Len(); i++ {
			z = z && isZero(v.Index(i))
		}
		return z
	case reflect.Struct:
		if v.Type() == reflect.TypeOf(t) {
			return v.Interface().(time.Time).IsZero()
		}
		z := true
		for i := 0; i < v.NumField(); i++ {
			z = z && isZero(v.Field(i))
		}
		return z
	}
	z := reflect.Zero(v.Type())
	return v.Interface() == z.Interface()
}

func BuildRequestBody(opts any, parent string) (map[string]any, error) {
	optsValue := reflect.ValueOf(opts)
	if optsValue.Kind() == reflect.Ptr {
		optsValue = optsValue.Elem()
	}

	optsType := reflect.TypeOf(opts)
	if optsType.Kind() == reflect.Ptr {
		optsType = optsType.Elem()
	}

	optsMap := make(map[string]any)
	switch optsValue.Kind() {
	case reflect.Struct:
		for i := 0; i < optsValue.NumField(); i++ {
			v := optsValue.Field(i)
			f := optsType.Field(i)

			if len(f.Name) == 0 || f.Name[0] < 'A' || f.Name[0] > 'Z' {
				continue
			}

			zero := isZero(v)

			if requiredTag := f.Tag.Get("required"); requiredTag == "true" {
				if zero {
					err := client.ErrMissingInput{}
					err.Argument = f.Name
					return nil, err
				}
			}

			if xorTag := f.Tag.Get("xor"); xorTag != "" {
				xorField := optsValue.FieldByName(xorTag)
				var xorFieldIsZero bool
				if reflect.ValueOf(xorField.Interface()) == reflect.Zero(xorField.Type()) {
					xorFieldIsZero = true
				} else {
					if xorField.Kind() == reflect.Ptr {
						xorField = xorField.Elem()
					}
					xorFieldIsZero = isZero(xorField)
				}
				if !(zero != xorFieldIsZero) {
					err := client.ErrMissingInput{}
					err.Argument = fmt.Sprintf("%s/%s", f.Name, xorTag)
					err.Info = fmt.Sprintf("Exactly one of %s and %s must be provided", f.Name, xorTag)
					return nil, err
				}
			}

			if orTag := f.Tag.Get("or"); orTag != "" {
				if zero {
					orField := optsValue.FieldByName(orTag)
					var orFieldIsZero bool
					if reflect.ValueOf(orField.Interface()) == reflect.Zero(orField.Type()) {
						orFieldIsZero = true
					} else {
						if orField.Kind() == reflect.Ptr {
							orField = orField.Elem()
						}
						orFieldIsZero = isZero(orField)
					}
					if orFieldIsZero {
						err := client.ErrMissingInput{}
						err.Argument = fmt.Sprintf("%s/%s", f.Name, orTag)
						err.Info = fmt.Sprintf("At least one of %s and %s must be provided", f.Name, orTag)
						return nil, err
					}
				}
			}

			jsonTag := f.Tag.Get("json")
			if jsonTag == "-" {
				continue
			}

			if v.Kind() == reflect.Slice || (v.Kind() == reflect.Ptr && v.Elem().Kind() == reflect.Slice) {
				sliceValue := v
				if sliceValue.Kind() == reflect.Ptr {
					sliceValue = sliceValue.Elem()
				}

				for i := 0; i < sliceValue.Len(); i++ {
					element := sliceValue.Index(i)
					if element.Kind() == reflect.Struct || (element.Kind() == reflect.Ptr && element.Elem().Kind() == reflect.Struct) {
						_, err := BuildRequestBody(element.Interface(), "")
						if err != nil {
							return nil, err
						}
					}
				}
			}
			if v.Kind() == reflect.Struct || (v.Kind() == reflect.Ptr && v.Elem().Kind() == reflect.Struct) {
				if zero {
					if jsonTag != "" {
						jsonTagPieces := strings.Split(jsonTag, ",")
						if len(jsonTagPieces) > 1 && jsonTagPieces[1] == "omitempty" {
							if v.CanSet() {
								if !v.IsNil() {
									if v.Kind() == reflect.Ptr {
										v.Set(reflect.Zero(v.Type()))
									}
								}
							}
						}
					}
					continue
				}

				_, err := BuildRequestBody(v.Interface(), f.Name)
				if err != nil {
					return nil, err
				}
			}
		}

		b, err := json.Marshal(opts)
		if err != nil {
			return nil, err
		}

		err = json.Unmarshal(b, &optsMap)
		if err != nil {
			return nil, err
		}

		if parent != "" {
			optsMap = map[string]any{parent: optsMap}
		}
		return optsMap, nil
	case reflect.Slice, reflect.Array:
		optsMaps := make([]map[string]any, optsValue.Len())
		for i := 0; i < optsValue.Len(); i++ {
			b, err := BuildRequestBody(optsValue.Index(i).Interface(), "")
			if err != nil {
				return nil, err
			}
			optsMaps[i] = b
		}
		if parent == "" {
			return nil, fmt.Errorf("Parent is required when passing an array or a slice.")
		}
		return map[string]any{parent: optsMaps}, nil
	}
	return nil, fmt.Errorf("Options type is not a struct, a slice, or an array.")
}

// CheckDeleted checks the error to see if it's a 404 (Not Found) and, if so,
// sets the resource ID to the empty string instead of throwing an error.
func CheckDeleted(d *schema.ResourceData, err error, msg string) error {
	if client.ResponseCodeIs(err, http.StatusNotFound) {
		d.SetId("")
		return nil
	}

	return fmt.Errorf("%s %s: %s", msg, d.Id(), err)
}

func CheckNotFound(d *schema.ResourceData, err error, msg string) error {
	var codeError client.ErrUnexpectedResponseCode
	if errors.As(err, &codeError) && codeError.Actual == http.StatusNotFound {
		d.SetId("")
		return nil
	}

	return fmt.Errorf("%s %s: %s", msg, d.Id(), err)
}

func ResponseCodeIs(err error, status int) bool {
	var codeError client.ErrUnexpectedResponseCode
	if errors.As(err, &codeError) {
		return codeError.Actual == status
	}
	return false
}
