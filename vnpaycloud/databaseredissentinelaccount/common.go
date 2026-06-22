package databaseredissentinelaccount

import (
	"context"
	"fmt"
	"terraform-provider-vnpaycloud/vnpaycloud/dto"
	"terraform-provider-vnpaycloud/vnpaycloud/helper/client"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/retry"
)

func redisSentinelAccountStateRefreshFunc(ctx context.Context, c *client.Client, projectID, id string) retry.StateRefreshFunc {
	return func() (interface{}, string, error) {
		resp := &dto.RedisSentinelAccountResponse{}
		httpResp, err := c.Get(ctx, client.ApiPath.DatabaseRedisSentinelAccountWithID(projectID, id), resp, nil)
		if err != nil {
			if httpResp != nil && httpResp.StatusCode == 404 {
				return resp, "deleted", nil
			}
			return nil, "", err
		}

		status := resp.RedisSentinelAccount.Status
		if status == "" {
			status = "unknown"
		}
		if status == "error" {
			return resp, status, fmt.Errorf("vnpaycloud_database_redis_sentinel_account %s is in error state", id)
		}
		return resp, status, nil
	}
}
