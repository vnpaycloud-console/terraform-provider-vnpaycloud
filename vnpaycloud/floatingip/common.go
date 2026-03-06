package floatingip

import (
	"context"
	"fmt"
	"net/http"
	"terraform-provider-vnpaycloud/vnpaycloud/dto"
	"terraform-provider-vnpaycloud/vnpaycloud/helper/client"
	"terraform-provider-vnpaycloud/vnpaycloud/util"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/retry"
)

func floatingIPStateRefreshFunc(ctx context.Context, c *client.Client, projectID, floatingIPID string) retry.StateRefreshFunc {
	return func() (interface{}, string, error) {
		fipResp := &dto.FloatingIPResponse{}
		_, err := c.Get(ctx, client.ApiPath.FloatingIPWithID(projectID, floatingIPID), fipResp, nil)

		if err != nil {
			if util.ResponseCodeIs(err, http.StatusNotFound) {
				return fipResp.FloatingIP, "deleted", nil
			}
			return nil, "", err
		}

		if fipResp.FloatingIP.Status == "failed" {
			return fipResp.FloatingIP, fipResp.FloatingIP.Status, fmt.Errorf("The floating IP is in error status. " +
				"Please check with your cloud admin or check the API logs.")
		}

		return fipResp.FloatingIP, fipResp.FloatingIP.Status, nil
	}
}
