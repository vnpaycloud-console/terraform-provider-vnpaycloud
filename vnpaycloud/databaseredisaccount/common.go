package databaseredisaccount

import (
	"context"
	"fmt"
	"terraform-provider-vnpaycloud/vnpaycloud/dto"
	"terraform-provider-vnpaycloud/vnpaycloud/helper/client"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/retry"
)

func redisAccountStateRefreshFunc(ctx context.Context, c *client.Client, projectID, id string) retry.StateRefreshFunc {
	return func() (interface{}, string, error) {
		resp := &dto.RedisAccountResponse{}
		httpResp, err := c.Get(ctx, client.ApiPath.DatabaseRedisAccountWithID(projectID, id), resp, nil)
		if err != nil {
			if httpResp != nil && httpResp.StatusCode == 404 {
				return resp, "deleted", nil
			}
			return nil, "", err
		}

		status := resp.RedisAccount.Status
		if status == "" {
			status = "unknown"
		}
		if status == "error" {
			return resp, status, fmt.Errorf("vnpaycloud_database_redis_account %s is in error state", id)
		}
		return resp, status, nil
	}
}
