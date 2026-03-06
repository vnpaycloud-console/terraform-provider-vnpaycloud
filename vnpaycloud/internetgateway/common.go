package internetgateway

import (
	"context"
	"fmt"
	"net/http"
	"terraform-provider-vnpaycloud/vnpaycloud/dto"
	"terraform-provider-vnpaycloud/vnpaycloud/helper/client"
	"terraform-provider-vnpaycloud/vnpaycloud/util"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/retry"
)

func internetGatewayStateRefreshFunc(ctx context.Context, c *client.Client, projectID, igwID string) retry.StateRefreshFunc {
	return func() (interface{}, string, error) {
		igwResp := &dto.InternetGatewayResponse{}
		_, err := c.Get(ctx, client.ApiPath.InternetGatewayWithID(projectID, igwID), igwResp, nil)

		if err != nil {
			if util.ResponseCodeIs(err, http.StatusNotFound) {
				return igwResp.InternetGateway, "deleted", nil
			}
			return nil, "", err
		}

		if igwResp.InternetGateway.Status == "failed" || igwResp.InternetGateway.Status == "error" {
			return igwResp.InternetGateway, igwResp.InternetGateway.Status, fmt.Errorf("The internet gateway is in error status. " +
				"Please check with your cloud admin or check the API logs.")
		}

		return igwResp.InternetGateway, igwResp.InternetGateway.Status, nil
	}
}
