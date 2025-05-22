package peeringconnection

import (
	"context"
	"fmt"
	"net/http"
	"terraform-provider-vnpaycloud/vnpaycloud/helper/client"
	"terraform-provider-vnpaycloud/vnpaycloud/util"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/retry"
)

func peeringConnectionRequestStateRefreshFunc(ctx context.Context, consoleClient *client.Client, id string) retry.StateRefreshFunc {
	return func() (interface{}, string, error) {
		getResp := &GetPeeringConnectionRequestResponse{}
		_, err := consoleClient.Get(ctx, client.ApiPath.PeeringConnectionRequestWithId(id), getResp, nil)

		if err != nil {
			if util.ResponseCodeIs(err, http.StatusNotFound) {
				return getResp.PeeringConnectionRequest, "OS_DELETED", nil
			}

			return nil, "", err
		}

		if getResp.PeeringConnectionRequest.Status == "OS_FAILED" {
			return getResp.PeeringConnectionRequest, getResp.PeeringConnectionRequest.Status, fmt.Errorf("The Peering Connection Request is in error status. " +
				"Please check with your cloud admin or check the Peering Connection Request " +
				"API logs to see why this error occurred.")
		}

		return getResp.PeeringConnectionRequest, getResp.PeeringConnectionRequest.Status, nil
	}
}

func peeringConnectionStateRefreshFunc(ctx context.Context, consoleClient *client.Client, id string) retry.StateRefreshFunc {
	return func() (interface{}, string, error) {
		resp := &GetPeeringConnectionResponse{}
		_, err := consoleClient.Get(ctx, client.ApiPath.PeeringConnectionWithId(id), resp, nil)

		if err != nil {
			if util.ResponseCodeIs(err, http.StatusNotFound) {
				return resp.PeeringConnection, "OS_DELETED", nil
			}

			return nil, "", err
		}

		if resp.PeeringConnection.Status == "OS_FAILED" {
			return resp.PeeringConnection, resp.PeeringConnection.Status, fmt.Errorf("The Peering Connection is in error status. " +
				"Please check with your cloud admin or check the Peering Connection " +
				"API logs to see why this error occurred.")
		}

		return resp.PeeringConnection, resp.PeeringConnection.Status, nil
	}
}
