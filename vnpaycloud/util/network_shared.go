package util

import (
	"encoding/json"
	"log"
	"net/http"

	"github.com/vnpaycloud-console/gophercloud/v2"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

func NetworkingReadAttributesTags(d *schema.ResourceData, tags []string) {
	ExpandObjectReadTags(d, tags)
}

func NetworkingUpdateAttributesTags(d *schema.ResourceData) []string {
	return ExpandObjectUpdateTags(d)
}

func NetworkingAttributesTags(d *schema.ResourceData) []string {
	return ExpandObjectTags(d)
}

type NeutronErrorWrap struct {
	NeutronError NeutronError
}

type NeutronError struct {
	Message string `json:"message"`
	Type    string `json:"type"`
	Detail  string `json:"detail"`
}

func RetryOn409(err error) bool {
	e, ok := err.(gophercloud.ErrUnexpectedResponseCode)
	if !ok {
		return false
	}

	switch e.Actual {
	case http.StatusConflict: // 409
		neutronError, err := DecodeNeutronError(e.Body)
		if err != nil {
			// retry, when error type cannot be detected
			log.Printf("[DEBUG] failed to decode a neutron error: %s", err)
			return true
		}
		if neutronError.Type == "IpAddressGenerationFailure" {
			return true
		}

		// don't retry on quota or other errors
		return false
	case http.StatusBadRequest: // 400
		neutronError, err := DecodeNeutronError(e.Body)
		if err != nil {
			// retry, when error type cannot be detected
			log.Printf("[DEBUG] failed to decode a neutron error: %s", err)
			return true
		}
		if neutronError.Type == "ExternalIpAddressExhausted" {
			return true
		}

		// don't retry on quota or other errors
		return false
	case http.StatusNotFound: // this case is handled mostly for functional tests
		return true
	}

	return false
}

func DecodeNeutronError(body []byte) (*NeutronError, error) {
	e := &NeutronErrorWrap{}
	if err := json.Unmarshal(body, e); err != nil {
		return nil, err
	}

	return &e.NeutronError, nil
}
