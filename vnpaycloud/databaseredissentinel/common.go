package databaseredissentinel

import (
	"context"
	"terraform-provider-vnpaycloud/vnpaycloud/dto"
	"terraform-provider-vnpaycloud/vnpaycloud/helper/client"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/retry"
)

func redisSentinelInstanceStateRefreshFunc(ctx context.Context, c *client.Client, projectID, id string) retry.StateRefreshFunc {
	return func() (interface{}, string, error) {
		resp := &dto.RedisSentinelInstanceResponse{}
		httpResp, err := c.Get(ctx, client.ApiPath.DatabaseRedisSentinelInstanceWithID(projectID, id), resp, nil)
		if err != nil {
			if httpResp != nil && httpResp.StatusCode == 404 {
				return resp, "deleted", nil
			}
			return nil, "", err
		}

		status := resp.RedisSentinelInstance.Status
		if status == "" {
			status = "unknown"
		}
		return resp, status, nil
	}
}
