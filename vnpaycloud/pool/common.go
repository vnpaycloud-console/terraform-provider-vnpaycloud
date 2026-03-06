package pool

import (
	"context"
	"fmt"
	"net/http"
	"terraform-provider-vnpaycloud/vnpaycloud/dto"
	"terraform-provider-vnpaycloud/vnpaycloud/helper/client"
	"terraform-provider-vnpaycloud/vnpaycloud/util"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/retry"
)

func poolStateRefreshFunc(ctx context.Context, c *client.Client, projectID, poolID string) retry.StateRefreshFunc {
	return func() (interface{}, string, error) {
		resp := &dto.PoolResponse{}
		_, err := c.Get(ctx, client.ApiPath.PoolWithID(projectID, poolID), resp, nil)

		if err != nil {
			if util.ResponseCodeIs(err, http.StatusNotFound) {
				return resp.Pool, "deleted", nil
			}
			return nil, "", err
		}

		if resp.Pool.Status == "failed" || resp.Pool.Status == "error" {
			return resp.Pool, resp.Pool.Status, fmt.Errorf("The pool is in error status. " +
				"Please check with your cloud admin or check the API logs.")
		}

		return resp.Pool, resp.Pool.Status, nil
	}
}
