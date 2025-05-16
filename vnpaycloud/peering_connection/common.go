package peeringconnection

import (
	"context"
	"fmt"
	"net/http"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/retry"
	"github.com/vnpaycloud-console/gophercloud/v2"
	"github.com/vnpaycloud-console/gophercloud/v2/openstack/networking/v2/peeringconnectionrequests"
	"github.com/vnpaycloud-console/gophercloud/v2/openstack/networking/v2/peeringconnections"
)

func peeringConnectionRequestStateRefreshFunc(ctx context.Context, client *gophercloud.ServiceClient, requestId string) retry.StateRefreshFunc {
	return func() (interface{}, string, error) {
		peeringConnectionRequest, err := peeringconnectionrequests.Get(ctx, client, requestId).Extract()
		fmt.Println("peeringConnectionRequest: ", peeringConnectionRequest)

		if err != nil {
			if gophercloud.ResponseCodeIs(err, http.StatusNotFound) {
				return peeringConnectionRequest, "OS_DELETED", nil
			}

			return nil, "", err
		}

		if peeringConnectionRequest.Status == "OS_FAILED" {
			return peeringConnectionRequest, peeringConnectionRequest.Status, fmt.Errorf("The Peering Connection Request is in error status. " +
				"Please check with your cloud admin or check the Peering Connection Request " +
				"API logs to see why this error occurred.")
		}

		return peeringConnectionRequest, peeringConnectionRequest.Status, nil
	}
}

func peeringConnectionStateRefreshFunc(ctx context.Context, client *gophercloud.ServiceClient, id string) retry.StateRefreshFunc {
	return func() (interface{}, string, error) {
		peeringConnection, err := peeringconnections.Get(ctx, client, id).Extract()

		if err != nil {
			if gophercloud.ResponseCodeIs(err, http.StatusNotFound) {
				return peeringConnection, "OS_DELETED", nil
			}

			return nil, "", err
		}

		if peeringConnection.Status == "OS_FAILED" {
			return peeringConnection, peeringConnection.Status, fmt.Errorf("The Peering Connection is in error status. " +
				"Please check with your cloud admin or check the Peering Connection " +
				"API logs to see why this error occurred.")
		}

		return peeringConnection, peeringConnection.Status, nil
	}
}
