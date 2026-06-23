package vpnconnection

import (
	"context"
	"fmt"
	"net/http"
	"terraform-provider-vnpaycloud/vnpaycloud/dto"
	"terraform-provider-vnpaycloud/vnpaycloud/helper/client"
	"terraform-provider-vnpaycloud/vnpaycloud/util"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/retry"
)

func vpnConnectionStateRefreshFunc(ctx context.Context, c *client.Client, projectID, vpnConnectionID string) retry.StateRefreshFunc {
	return func() (interface{}, string, error) {
		resp := &dto.VPNConnectionResponse{}
		_, err := c.Get(ctx, client.ApiPath.VPNConnectionWithID(projectID, vpnConnectionID), resp, nil)
		if err != nil {
			if util.ResponseCodeIs(err, http.StatusNotFound) {
				return resp, "deleted", nil
			}

			return nil, "", err
		}

		status := util.NormalizeStatus(resp.VPNConnection.Status)
		if status == "failed" || status == "error" {
			return resp.VPNConnection, status, fmt.Errorf("The VPN Connection is in error status. " +
				"Please check with your cloud admin or check the API logs.")
		}

		return resp.VPNConnection, status, nil
	}
}
