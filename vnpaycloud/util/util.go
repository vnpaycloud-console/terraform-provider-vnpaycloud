package util

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"reflect"
	"strconv"
	"strings"
	"time"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/retry"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"

	"terraform-provider-vnpaycloud/vnpaycloud/config"
	"terraform-provider-vnpaycloud/vnpaycloud/helper/client"
)

var t time.Time

func isZero(v reflect.Value) bool {
	//fmt.Printf("\n\nchecking isZero for value: %+v\n", v)
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
	// Compare other types directly:
	z := reflect.Zero(v.Type())
	//fmt.Printf("zero type for value: %+v\n\n\n", z)
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
		//fmt.Printf("optsValue.Kind() is a reflect.Struct: %+v\n", optsValue.Kind())
		for i := 0; i < optsValue.NumField(); i++ {
			v := optsValue.Field(i)
			f := optsType.Field(i)

			if f.Name != strings.Title(f.Name) {
				//fmt.Printf("Skipping field: %s...\n", f.Name)
				continue
			}

			//fmt.Printf("Starting on field: %s...\n", f.Name)

			zero := isZero(v)
			//fmt.Printf("v is zero?: %v\n", zero)

			// if the field has a required tag that's set to "true"
			if requiredTag := f.Tag.Get("required"); requiredTag == "true" {
				//fmt.Printf("Checking required field [%s]:\n\tv: %+v\n\tisZero:%v\n", f.Name, v.Interface(), zero)
				// if the field's value is zero, return a missing-argument error
				if zero {
					// if the field has a 'required' tag, it can't have a zero-value
					err := client.ErrMissingInput{}
					err.Argument = f.Name
					return nil, err
				}
			}

			if xorTag := f.Tag.Get("xor"); xorTag != "" {
				//fmt.Printf("Checking `xor` tag for field [%s] with value %+v:\n\txorTag: %s\n", f.Name, v, xorTag)
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
				//fmt.Printf("Checking `or` tag for field with:\n\tname: %+v\n\torTag:%s\n", f.Name, orTag)
				//fmt.Printf("field is zero?: %v\n", zero)
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
					//fmt.Printf("value before change: %+v\n", optsValue.Field(i))
					if jsonTag != "" {
						jsonTagPieces := strings.Split(jsonTag, ",")
						if len(jsonTagPieces) > 1 && jsonTagPieces[1] == "omitempty" {
							if v.CanSet() {
								if !v.IsNil() {
									if v.Kind() == reflect.Ptr {
										v.Set(reflect.Zero(v.Type()))
									}
								}
								//fmt.Printf("value after change: %+v\n", optsValue.Field(i))
							}
						}
					}
					continue
				}

				//fmt.Printf("Calling BuildRequestBody with:\n\tv: %+v\n\tf.Name:%s\n", v.Interface(), f.Name)
				_, err := BuildRequestBody(v.Interface(), f.Name)
				if err != nil {
					return nil, err
				}
			}
		}

		//fmt.Printf("opts: %+v \n", opts)

		b, err := json.Marshal(opts)
		if err != nil {
			return nil, err
		}

		//fmt.Printf("string(b): %s\n", string(b))

		err = json.Unmarshal(b, &optsMap)
		if err != nil {
			return nil, err
		}

		//fmt.Printf("optsMap: %+v\n", optsMap)

		if parent != "" {
			optsMap = map[string]any{parent: optsMap}
		}
		//fmt.Printf("optsMap after parent added: %+v\n", optsMap)
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
	// Return an error if we can't work with the underlying type of 'opts'
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

// GetRegion returns the region that was specified in the resource. If a
// region was not set, the provider-level region is checked. The provider-level
// region can either be set by the region argument or by OS_REGION_NAME.
func GetRegion(d *schema.ResourceData, config *config.Config) string {
	if v, ok := d.GetOk("region"); ok {
		return v.(string)
	}

	return "RegionOne"
}

// AddValueSpecs expands the 'value_specs' object and removes 'value_specs'
// from the reqeust body.
func AddValueSpecs(body map[string]interface{}) map[string]interface{} {
	if body["value_specs"] != nil {
		for k, v := range body["value_specs"].(map[string]interface{}) {
			// this hack allows to pass boolean values as strings
			if v == "true" || v == "false" {
				body[k] = v == "true"
				continue
			}
			body[k] = v
		}
		delete(body, "value_specs")
	}

	return body
}

// MapValueSpecs converts ResourceData into a map.
func MapValueSpecs(d *schema.ResourceData) map[string]string {
	m := make(map[string]string)
	for key, val := range d.Get("value_specs").(map[string]interface{}) {
		m[key] = val.(string)
	}
	return m
}

func CheckForRetryableError(err error) *retry.RetryError {
	e, ok := err.(client.ErrUnexpectedResponseCode)
	if !ok {
		return retry.NonRetryableError(err)
	}

	switch e.Actual {
	case http.StatusConflict, // 409
		http.StatusInternalServerError, // 500
		http.StatusBadGateway,          // 502
		http.StatusServiceUnavailable,  // 503
		http.StatusGatewayTimeout:      // 504
		return retry.RetryableError(err)
	}

	return retry.NonRetryableError(err)
}

func SuppressEquivalentTimeDiffs(k, old, new string, d *schema.ResourceData) bool {
	oldTime, err := time.Parse(time.RFC3339, old)
	if err != nil {
		return false
	}

	newTime, err := time.Parse(time.RFC3339, new)
	if err != nil {
		return false
	}

	return oldTime.Equal(newTime)
}

func ResourceNetworkingAvailabilityZoneHints(d *schema.ResourceData) []string {
	rawAZH := d.Get("availability_zone_hints").([]interface{})
	azh := make([]string, len(rawAZH))
	for i, raw := range rawAZH {
		azh[i] = raw.(string)
	}
	return azh
}

func ExpandVendorOptions(vendOptsRaw []interface{}) map[string]interface{} {
	vendorOptions := make(map[string]interface{})

	for _, option := range vendOptsRaw {
		for optKey, optValue := range option.(map[string]interface{}) {
			vendorOptions[optKey] = optValue
		}
	}

	return vendorOptions
}

func ExpandObjectReadTags(d *schema.ResourceData, tags []string) {
	d.Set("all_tags", tags)

	allTags := d.Get("all_tags").(*schema.Set)
	desiredTags := d.Get("tags").(*schema.Set)
	actualTags := allTags.Intersection(desiredTags)
	if !actualTags.Equal(desiredTags) {
		d.Set("tags", ExpandToStringSlice(actualTags.List()))
	}
}

func ExpandObjectUpdateTags(d *schema.ResourceData) []string {
	allTags := d.Get("all_tags").(*schema.Set)
	oldTagsRaw, newTagsRaw := d.GetChange("tags")
	oldTags, newTags := oldTagsRaw.(*schema.Set), newTagsRaw.(*schema.Set)

	allTagsWithoutOld := allTags.Difference(oldTags)

	return ExpandToStringSlice(allTagsWithoutOld.Union(newTags).List())
}

func ExpandObjectTags(d *schema.ResourceData) []string {
	rawTags := d.Get("tags").(*schema.Set).List()
	tags := make([]string, len(rawTags))

	for i, raw := range rawTags {
		tags[i] = raw.(string)
	}

	return tags
}

func ExpandToMapStringString(v map[string]interface{}) map[string]string {
	m := make(map[string]string, len(v))
	for key, val := range v {
		if strVal, ok := val.(string); ok {
			m[key] = strVal
		}
	}

	return m
}

func ExpandToStringSlice(v []interface{}) []string {
	s := make([]string, len(v))
	for i, val := range v {
		if strVal, ok := val.(string); ok {
			s[i] = strVal
		}
	}

	return s
}

// StrSliceContains checks if a given string is contained in a slice
// When anybody asks why Go needs generics, here you go.
func StrSliceContains(haystack []string, needle string) bool {
	for _, s := range haystack {
		if s == needle {
			return true
		}
	}
	return false
}

func SliceUnion(a, b []string) []string {
	var res []string
	for _, i := range a {
		if !StrSliceContains(res, i) {
			res = append(res, i)
		}
	}
	for _, k := range b {
		if !StrSliceContains(res, k) {
			res = append(res, k)
		}
	}
	return res
}

// compatibleMicroversion will determine if an obtained microversion is
// compatible with a given microversion.
func CompatibleMicroversion(direction, required, given string) (bool, error) {
	if direction != "min" && direction != "max" {
		return false, fmt.Errorf("Invalid microversion direction %s. Must be min or max", direction)
	}

	if required == "" || given == "" {
		return false, nil
	}

	requiredParts := strings.Split(required, ".")
	if len(requiredParts) != 2 {
		return false, fmt.Errorf("Not a valid microversion: %s", required)
	}

	givenParts := strings.Split(given, ".")
	if len(givenParts) != 2 {
		return false, fmt.Errorf("Not a valid microversion: %s", given)
	}

	requiredMajor, requiredMinor := requiredParts[0], requiredParts[1]
	givenMajor, givenMinor := givenParts[0], givenParts[1]

	requiredMajorInt, err := strconv.Atoi(requiredMajor)
	if err != nil {
		return false, fmt.Errorf("Unable to parse microversion: %s", required)
	}

	requiredMinorInt, err := strconv.Atoi(requiredMinor)
	if err != nil {
		return false, fmt.Errorf("Unable to parse microversion: %s", required)
	}

	givenMajorInt, err := strconv.Atoi(givenMajor)
	if err != nil {
		return false, fmt.Errorf("Unable to parse microversion: %s", given)
	}

	givenMinorInt, err := strconv.Atoi(givenMinor)
	if err != nil {
		return false, fmt.Errorf("Unable to parse microversion: %s", given)
	}

	switch direction {
	case "min":
		if requiredMajorInt == givenMajorInt {
			if requiredMinorInt <= givenMinorInt {
				return true, nil
			}
		}
	case "max":
		if requiredMajorInt == givenMajorInt {
			if requiredMinorInt >= givenMinorInt {
				return true, nil
			}
		}
	}

	return false, nil
}

func ValidateJSONObject(v interface{}, k string) ([]string, []error) {
	if v == nil || v.(string) == "" {
		return nil, []error{fmt.Errorf("%q value must not be empty", k)}
	}

	var j map[string]interface{}
	s := v.(string)

	err := json.Unmarshal([]byte(s), &j)
	if err != nil {
		return nil, []error{fmt.Errorf("%q must be a JSON object: %s", k, err)}
	}

	return nil, nil
}

func DiffSuppressJSONObject(k, old, new string, d *schema.ResourceData) bool {
	if StrSliceContains([]string{"{}", ""}, old) &&
		StrSliceContains([]string{"{}", ""}, new) {
		return true
	}
	return false
}

// Metadata in vnpaycloud are not fully replaced with a "set"
// operation, instead, it's only additive, and the existing
// metadata are only removed when set to `null` value in json.
func MapDiffWithNilValues(oldMap, newMap map[string]interface{}) (output map[string]interface{}) {
	output = make(map[string]interface{})

	for k, v := range newMap {
		output[k] = v
	}

	for key := range oldMap {
		_, ok := newMap[key]
		if !ok {
			output[key] = nil
		}
	}

	return
}

// parsePairedIDs is a helper function that parses a raw ID into two
// separate IDs. This is useful for resources that have a parent/child
// relationship.
func ParsePairedIDs(id string, res string) (string, string, error) {
	parts := strings.SplitN(id, "/", 2)
	if len(parts) != 2 {
		return "", "", fmt.Errorf("Unable to determine %s ID from raw ID: %s", res, id)
	}

	return parts[0], parts[1], nil
}

// getOkExists is a helper function that replaces the deprecated GetOkExists
// schema method. It returns the value of the key if it exists in the
// configuration, along with a boolean indicating if the key exists.
func GetOkExists(d *schema.ResourceData, key string) (interface{}, bool) {
	v := d.GetRawConfig().GetAttr(key)
	if v.IsNull() {
		return nil, false
	}
	return d.Get(key), true
}

type AuthScopeTokenInfo struct {
	UserID    string
	projectID string
	tokenID   string
}

func GetTokenInfo(ctx context.Context, sc *client.Client) (AuthScopeTokenInfo, error) {
	return AuthScopeTokenInfo{
		projectID: sc.GetProjectID(),
		UserID:    sc.GetUserID(),
		tokenID:   sc.GetTokenID(),
	}, nil
}

func ResponseCodeIs(err error, status int) bool {
	var codeError client.ErrUnexpectedResponseCode
	if errors.As(err, &codeError) {
		return codeError.Actual == status
	}
	return false
}
