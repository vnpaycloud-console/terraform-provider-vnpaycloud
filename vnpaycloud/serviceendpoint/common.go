package serviceendpoint

import (
	"context"
	"fmt"
	"net/http"
	"terraform-provider-vnpaycloud/vnpaycloud/dto"
	"terraform-provider-vnpaycloud/vnpaycloud/helper/client"
	"terraform-provider-vnpaycloud/vnpaycloud/util"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/retry"
)

func serviceEndpointStateRefreshFunc(ctx context.Context, c *client.Client, projectID, id string) retry.StateRefreshFunc {
	return func() (interface{}, string, error) {
		resp := &dto.ServiceEndpointResponse{}
		_, err := c.Get(ctx, client.ApiPath.ServiceEndpointWithID(projectID, id), resp, nil)

		if err != nil {
			if util.ResponseCodeIs(err, http.StatusNotFound) {
				return resp.ServiceEndpoint, "deleted", nil
			}
			return nil, "", err
		}

		status := resp.ServiceEndpoint.Status
		if status == "" {
			status = "unknown"
		}
		if status == "error" {
			return resp.ServiceEndpoint, status, fmt.Errorf("vnpaycloud_service_endpoint %s is in error state", id)
		}

		return resp.ServiceEndpoint, status, nil
	}
}
