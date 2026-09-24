package backupvault

import (
	"context"
	"fmt"
	"math"
	"terraform-provider-vnpaycloud/vnpaycloud/dto"
	"terraform-provider-vnpaycloud/vnpaycloud/helper/client"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/retry"
)

func backupVaultStateRefreshFunc(ctx context.Context, c *client.Client, projectID, id string) retry.StateRefreshFunc {
	return func() (interface{}, string, error) {
		resp := &dto.BackupVaultResponse{}
		httpResp, err := c.Get(ctx, client.ApiPath.BackupVaultWithID(projectID, id), resp, nil)
		if err != nil {
			if httpResp != nil && httpResp.StatusCode == 404 {
				return resp, "deleted", nil
			}
			return nil, "", err
		}

		status := resp.BackupVault.Status
		if status == "" {
			status = "unknown"
		}
		if status == "error" {
			return resp, status, fmt.Errorf("vnpaycloud_backup_vault %s is in error state", id)
		}
		return resp, status, nil
	}
}

func committedQuotaForState(v int64) int {
	if v == math.MaxInt64 {
		return -1
	}
	return int(v)
}
