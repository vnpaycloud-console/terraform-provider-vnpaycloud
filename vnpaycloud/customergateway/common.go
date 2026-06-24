package customergateway

import (
	"context"
	"fmt"
	"net/http"
	"terraform-provider-vnpaycloud/vnpaycloud/dto"
	"terraform-provider-vnpaycloud/vnpaycloud/helper/client"
	"terraform-provider-vnpaycloud/vnpaycloud/util"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/retry"
)

func customerGatewayStateRefreshFunc(ctx context.Context, c *client.Client, projectID, cgID string) retry.StateRefreshFunc {
	return func() (interface{}, string, error) {
		customerGatewayResp := &dto.CustomerGatewayResponse{}
		_, err := c.Get(ctx, client.ApiPath.CustomerGatewayWithID(projectID, cgID), customerGatewayResp, nil)

		if err != nil {
			if util.ResponseCodeIs(err, http.StatusNotFound) {
				return customerGatewayResp.CustomerGateway, "deleted", nil
			}
			return nil, "", err
		}

		status := util.NormalizeStatus(customerGatewayResp.CustomerGateway.Status)
		if status == "failed" || status == "error" {
			return customerGatewayResp.CustomerGateway, status, fmt.Errorf("The customer gateway is in error status. " +
				"Please check with your cloud admin or check the API logs.")
		}

		return customerGatewayResp.CustomerGateway, status, nil
	}
}
