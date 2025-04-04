package util

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/vnpaycloud-console/gophercloud/v2/openstack/identity/v2/tenants"
	tokens2 "github.com/vnpaycloud-console/gophercloud/v2/openstack/identity/v2/tokens"
	tokens3 "github.com/vnpaycloud-console/gophercloud/v2/openstack/identity/v3/tokens"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/retry"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"

	"terraform-provider-vnpaycloud/vnpaycloud/config"

	"github.com/vnpaycloud-console/gophercloud/v2"
)

// BuildRequest takes an opts struct and builds a request body for
// Gophercloud to execute.
func BuildRequest(opts interface{}, parent string) (map[string]interface{}, error) {
	b, err := gophercloud.BuildRequestBody(opts, "")
	if err != nil {
		return nil, err
	}

	b = AddValueSpecs(b)

	return map[string]interface{}{parent: b}, nil
}

// CheckDeleted checks the error to see if it's a 404 (Not Found) and, if so,
// sets the resource ID to the empty string instead of throwing an error.
func CheckDeleted(d *schema.ResourceData, err error, msg string) error {
	if gophercloud.ResponseCodeIs(err, http.StatusNotFound) {
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

	return config.Region
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
	e, ok := err.(gophercloud.ErrUnexpectedResponseCode)
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

func suppressEquivalentTimeDiffs(k, old, new string, d *schema.ResourceData) bool {
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

func resourceNetworkingAvailabilityZoneHintsV2(d *schema.ResourceData) []string {
	rawAZH := d.Get("availability_zone_hints").([]interface{})
	azh := make([]string, len(rawAZH))
	for i, raw := range rawAZH {
		azh[i] = raw.(string)
	}
	return azh
}

func expandVendorOptions(vendOptsRaw []interface{}) map[string]interface{} {
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

func sliceUnion(a, b []string) []string {
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
func compatibleMicroversion(direction, required, given string) (bool, error) {
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

func validateJSONObject(v interface{}, k string) ([]string, []error) {
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

func diffSuppressJSONObject(k, old, new string, d *schema.ResourceData) bool {
	if StrSliceContains([]string{"{}", ""}, old) &&
		StrSliceContains([]string{"{}", ""}, new) {
		return true
	}
	return false
}

// Metadata in vnpaycloud are not fully replaced with a "set"
// operation, instead, it's only additive, and the existing
// metadata are only removed when set to `null` value in json.
func mapDiffWithNilValues(oldMap, newMap map[string]interface{}) (output map[string]interface{}) {
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
func parsePairedIDs(id string, res string) (string, string, error) {
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

func GetTokenInfo(ctx context.Context, sc *gophercloud.ServiceClient) (AuthScopeTokenInfo, error) {
	r := sc.ProviderClient.GetAuthResult()
	switch r := r.(type) {
	case tokens2.CreateResult:
		return GetTokenInfoV2(r)
	case tokens3.CreateResult, tokens3.GetResult:
		return GetTokenInfoV3(r)
	default:
		token := tokens3.Get(ctx, sc, sc.ProviderClient.TokenID)
		if token.Err != nil {
			return AuthScopeTokenInfo{}, token.Err
		}
		return GetTokenInfoV3(token)
	}
}

func GetTokenInfoV3(t interface{}) (AuthScopeTokenInfo, error) {
	var info AuthScopeTokenInfo
	switch r := t.(type) {
	case tokens3.CreateResult:
		user, err := r.ExtractUser()
		if err != nil {
			return info, err
		}
		project, err := r.ExtractProject()
		if err != nil {
			return info, err
		}
		info.UserID = user.ID
		if project != nil {
			info.projectID = project.ID
		}
		return info, nil
	case tokens3.GetResult:
		user, err := r.ExtractUser()
		if err != nil {
			return info, err
		}
		project, err := r.ExtractProject()
		if err != nil {
			return info, err
		}
		info.UserID = user.ID
		if project != nil {
			info.projectID = project.ID
		}
		return info, nil
	default:
		return info, fmt.Errorf("got unexpected AuthResult type %t", r)
	}
}

func GetTokenInfoV2(t tokens2.CreateResult) (AuthScopeTokenInfo, error) {
	var info AuthScopeTokenInfo
	var s struct {
		Access struct {
			Token struct {
				Expires string         `json:"expires"`
				ID      string         `json:"id"`
				Tenant  tenants.Tenant `json:"tenant"`
			} `json:"token"`
			User tokens2.User `json:"user"`
		} `json:"access"`
	}

	err := t.ExtractInto(&s)
	if err != nil {
		return info, err
	}
	info.UserID = s.Access.User.ID
	info.tokenID = s.Access.Token.ID
	return info, nil
}
