package vpnpublicip

import (
	"context"
	"fmt"
	"net/http"
	"terraform-provider-vnpaycloud/vnpaycloud/dto"
	"terraform-provider-vnpaycloud/vnpaycloud/helper/client"
	"terraform-provider-vnpaycloud/vnpaycloud/util"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/retry"
)

func vpnPublicIPStateRefreshFunc(ctx context.Context, c *client.Client, projectID, vpnPublicIPID string) retry.StateRefreshFunc {
	return func() (interface{}, string, error) {
		resp := &dto.VPNPublicIPResponse{}
		_, err := c.Get(ctx, client.ApiPath.VPNPublicIPWithID(projectID, vpnPublicIPID), resp, nil)

		if err != nil {
			if util.ResponseCodeIs(err, http.StatusNotFound) {
				return resp.VPNPublicIP, "deleted", nil
			}
			return nil, "", err
		}

		status := util.NormalizeStatus(resp.VPNPublicIP.Status)
		if status == "failed" || status == "error" {
			return resp.VPNPublicIP, status, fmt.Errorf("The VPN public IP is in error status. " +
				"Please check with your cloud admin or check the API logs.")
		}

		return resp.VPNPublicIP, status, nil
	}
}
