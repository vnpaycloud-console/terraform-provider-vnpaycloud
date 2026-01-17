package shared

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"terraform-provider-vnpaycloud/vnpaycloud/config"
	"terraform-provider-vnpaycloud/vnpaycloud/dto"
	"terraform-provider-vnpaycloud/vnpaycloud/helper/client"

	"terraform-provider-vnpaycloud/vnpaycloud/util"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
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
	e, ok := err.(client.ErrUnexpectedResponseCode)
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

// NetworkingNetworkID retrieves network ID by the provided name.
func NetworkingNetworkID(ctx context.Context, d *schema.ResourceData, meta interface{}, networkName string) (string, error) {
	config := meta.(*config.Config)
	networkingClient, err := client.NewClient(ctx, config.ConsoleClientConfig)
	if err != nil {
		return "", fmt.Errorf("Error creating VNPAYCLOUD network client: %s", err)
	}

	opts := dto.ListNetworkParams{Name: networkName}
	listResp := dto.ListNetworksResponse{}
	_, err = networkingClient.Get(ctx, client.ApiPath.NetworkWithParams(opts), &listResp, nil)
	if err != nil {
		return "", fmt.Errorf("Error getting VNPAYCLOUD network: %s", err)
	}

	networkID := ""

	for _, n := range listResp.Networks {
		if n.Name == networkName {
			networkID = n.ID
			break
		}
	}

	return networkID, err
}

// NetworkingNetworkName retrieves network name by the provided ID.
func NetworkingNetworkName(ctx context.Context, d *schema.ResourceData, meta interface{}, networkID string) (string, error) {
	config := meta.(*config.Config)
	networkingClient, err := client.NewClient(ctx, config.ConsoleClientConfig)
	if err != nil {
		return "", fmt.Errorf("Error creating VNPAYCLOUD network client: %s", err)
	}

	opts := dto.ListNetworkParams{ID: networkID}
	listResp := dto.ListNetworksResponse{}
	_, err = networkingClient.Get(ctx, client.ApiPath.NetworkWithParams(opts), &listResp, nil)
	if err != nil {
		return "", fmt.Errorf("Error getting VNPAYCLOUD network: %s", err)
	}

	networkName := ""

	for _, n := range listResp.Networks {
		if n.ID == networkID {
			networkName = n.Name
			break
		}
	}

	return networkName, err
}
