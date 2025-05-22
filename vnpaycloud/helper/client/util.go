package client

import (
	"fmt"
	"net/url"
	"reflect"
	"strconv"
	"strings"
	"time"
)

func BuildQueryString(opts any) (*url.URL, error) {
	optsValue := reflect.ValueOf(opts)
	if optsValue.Kind() == reflect.Ptr {
		optsValue = optsValue.Elem()
	}

	optsType := reflect.TypeOf(opts)
	if optsType.Kind() == reflect.Ptr {
		optsType = optsType.Elem()
	}

	params := url.Values{}

	if optsValue.Kind() == reflect.Struct {
		for i := 0; i < optsValue.NumField(); i++ {
			v := optsValue.Field(i)
			f := optsType.Field(i)
			qTag := f.Tag.Get("q")

			if qTag != "" {
				tags := strings.Split(qTag, ",")

				if !isZero(v) {
				loop:
					switch v.Kind() {
					case reflect.Ptr:
						v = v.Elem()
						goto loop
					case reflect.String:
						params.Add(tags[0], v.String())
					case reflect.Int:
						params.Add(tags[0], strconv.FormatInt(v.Int(), 10))
					case reflect.Bool:
						params.Add(tags[0], strconv.FormatBool(v.Bool()))
					case reflect.Slice:
						var values []string
						switch v.Type().Elem() {
						case reflect.TypeOf(0):
							for i := 0; i < v.Len(); i++ {
								values = append(values, strconv.FormatInt(v.Index(i).Int(), 10))
							}
						default:
							for i := 0; i < v.Len(); i++ {
								values = append(values, v.Index(i).String())
							}
						}
						if sliceFormat := f.Tag.Get("format"); sliceFormat == "comma-separated" {
							params.Add(tags[0], strings.Join(values, ","))
						} else {
							params[tags[0]] = append(params[tags[0]], values...)
						}
					case reflect.Map:
						if v.Type().Key().Kind() == reflect.String && v.Type().Elem().Kind() == reflect.String {
							var s []string
							for _, k := range v.MapKeys() {
								value := v.MapIndex(k).String()
								s = append(s, fmt.Sprintf("'%s':'%s'", k.String(), value))
							}
							params.Add(tags[0], fmt.Sprintf("{%s}", strings.Join(s, ", ")))
						}
					}
				} else {
					if requiredTag := f.Tag.Get("required"); requiredTag == "true" {
						return &url.URL{}, fmt.Errorf("Required query parameter [%s] not set.", f.Name)
					}
				}
			}
		}

		return &url.URL{RawQuery: params.Encode()}, nil
	}
	return nil, fmt.Errorf("Options type is not a struct.")
}

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
