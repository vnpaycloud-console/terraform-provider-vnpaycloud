package servicegateway

import (
	"context"
	"fmt"
	"net/http"
	"terraform-provider-vnpaycloud/vnpaycloud/dto"
	"terraform-provider-vnpaycloud/vnpaycloud/helper/client"
	"terraform-provider-vnpaycloud/vnpaycloud/util"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/retry"
)

func serviceGatewayStateRefreshFunc(ctx context.Context, c *client.Client, projectID, id string) retry.StateRefreshFunc {
	return func() (interface{}, string, error) {
		resp := &dto.ServiceGatewayResponse{}
		_, err := c.Get(ctx, client.ApiPath.ServiceGatewayWithID(projectID, id), resp, nil)

		if err != nil {
			if util.ResponseCodeIs(err, http.StatusNotFound) {
				return resp.ServiceGateway, "deleted", nil
			}
			return nil, "", err
		}

		status := resp.ServiceGateway.Status
		if status == "" {
			status = "unknown"
		}
		if status == "error" {
			return resp.ServiceGateway, status, fmt.Errorf("vnpaycloud_service_gateway %s is in error state", id)
		}

		return resp.ServiceGateway, status, nil
	}
}
