package privategateway

import (
	"context"
	"fmt"
	"net/http"
	"terraform-provider-vnpaycloud/vnpaycloud/dto"
	"terraform-provider-vnpaycloud/vnpaycloud/helper/client"
	"terraform-provider-vnpaycloud/vnpaycloud/util"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/retry"
)

func privateGatewayStateRefreshFunc(ctx context.Context, c *client.Client, projectID, pgwID string) retry.StateRefreshFunc {
	return func() (interface{}, string, error) {
		pgwResp := &dto.PrivateGatewayResponse{}
		_, err := c.Get(ctx, client.ApiPath.PrivateGatewayWithID(projectID, pgwID), pgwResp, nil)

		if err != nil {
			if util.ResponseCodeIs(err, http.StatusNotFound) {
				return pgwResp.PrivateGateway, "deleted", nil
			}
			return nil, "", err
		}

		if pgwResp.PrivateGateway.Status == "failed" || pgwResp.PrivateGateway.Status == "error" {
			return pgwResp.PrivateGateway, pgwResp.PrivateGateway.Status, fmt.Errorf("The private gateway is in error status. " +
				"Please check with your cloud admin or check the API logs.")
		}

		return pgwResp.PrivateGateway, pgwResp.PrivateGateway.Status, nil
	}
}
