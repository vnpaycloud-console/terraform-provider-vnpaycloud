package shared

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"terraform-provider-vnpaycloud/vnpaycloud/config"

	"terraform-provider-vnpaycloud/vnpaycloud/util"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/vnpaycloud-console/gophercloud/v2"
	"github.com/vnpaycloud-console/gophercloud/v2/openstack/networking/v2/networks"
	"github.com/vnpaycloud-console/gophercloud/v2/pagination"
)

func NetworkingReadAttributesTags(d *schema.ResourceData, tags []string) {
	util.ExpandObjectReadTags(d, tags)
}

func NetworkingUpdateAttributesTags(d *schema.ResourceData) []string {
	return util.ExpandObjectUpdateTags(d)
}

func NetworkingAttributesTags(d *schema.ResourceData) []string {
	return util.ExpandObjectTags(d)
}

type neutronErrorWrap struct {
	NeutronError neutronError
}

type neutronError struct {
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
		neutronError, err := decodeNeutronError(e.Body)
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
		neutronError, err := decodeNeutronError(e.Body)
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

func decodeNeutronError(body []byte) (*neutronError, error) {
	e := &neutronErrorWrap{}
	if err := json.Unmarshal(body, e); err != nil {
		return nil, err
	}

	return &e.NeutronError, nil
}

// NetworkingNetworkV2ID retrieves network ID by the provided name.
func NetworkingNetworkV2ID(ctx context.Context, d *schema.ResourceData, meta interface{}, networkName string) (string, error) {
	config := meta.(*config.Config)
	networkingClient, err := config.NetworkingV2Client(ctx, util.GetRegion(d, config))
	if err != nil {
		return "", fmt.Errorf("Error creating OpenStack network client: %s", err)
	}

	opts := networks.ListOpts{Name: networkName}
	pager := networks.List(networkingClient, opts)
	networkID := ""

	err = pager.EachPage(ctx, func(ctx context.Context, page pagination.Page) (bool, error) {
		networkList, err := networks.ExtractNetworks(page)
		if err != nil {
			return false, err
		}

		for _, n := range networkList {
			if n.Name == networkName {
				networkID = n.ID
				return false, nil
			}
		}

		return true, nil
	})

	return networkID, err
}

// NetworkingNetworkV2Name retrieves network name by the provided ID.
func NetworkingNetworkV2Name(ctx context.Context, d *schema.ResourceData, meta interface{}, networkID string) (string, error) {
	config := meta.(*config.Config)
	networkingClient, err := config.NetworkingV2Client(ctx, util.GetRegion(d, config))
	if err != nil {
		return "", fmt.Errorf("Error creating OpenStack network client: %s", err)
	}

	opts := networks.ListOpts{ID: networkID}
	pager := networks.List(networkingClient, opts)
	networkName := ""

	err = pager.EachPage(ctx, func(ctx context.Context, page pagination.Page) (bool, error) {
		networkList, err := networks.ExtractNetworks(page)
		if err != nil {
			return false, err
		}

		for _, n := range networkList {
			if n.ID == networkID {
				networkName = n.Name
				return false, nil
			}
		}

		return true, nil
	})

	return networkName, err
}
