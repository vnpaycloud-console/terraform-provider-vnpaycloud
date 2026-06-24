package vpngateway

import (
	"context"
	"fmt"
	"net/http"
	"terraform-provider-vnpaycloud/vnpaycloud/dto"
	"terraform-provider-vnpaycloud/vnpaycloud/helper/client"
	"terraform-provider-vnpaycloud/vnpaycloud/util"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/retry"
)

func vpnGatewayStateRefreshFunc(ctx context.Context, c *client.Client, projectID, vpnGatewayID string) retry.StateRefreshFunc {
	return func() (interface{}, string, error) {
		vpnGatewayResp := &dto.VPNGatewayResponse{}
		_, err := c.Get(ctx, client.ApiPath.VPNGatewayWithID(projectID, vpnGatewayID), vpnGatewayResp, nil)

		if err != nil {
			if util.ResponseCodeIs(err, http.StatusNotFound) {
				return vpnGatewayResp, "deleted", nil
			}

			return nil, "", err
		}

		status := util.NormalizeStatus(vpnGatewayResp.VPNGateway.Status)
		if status == "failed" || status == "error" {
			return vpnGatewayResp.VPNGateway, status, fmt.Errorf("The VPN Gateway is in error status. " +
				"Please check with your cloud admin or check the API logs.")
		}

		return vpnGatewayResp.VPNGateway, status, nil
	}
}
