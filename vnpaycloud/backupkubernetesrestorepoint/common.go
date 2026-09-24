package backupkubernetesrestorepoint

import (
	"context"

	"terraform-provider-vnpaycloud/vnpaycloud/dto"
	"terraform-provider-vnpaycloud/vnpaycloud/helper/client"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/retry"
)

const maxTransientRestorePointErrors = 20

// restorePointDeleteStateRefreshFunc polls the restore point until the backend
// finishes tearing it down: the delete is async (the record moves to "deleting"
// while Velero backup data and storage are rotated), so a 404 is the synthetic
// terminal "deleted" state the Delete wait is looking for.
func restorePointDeleteStateRefreshFunc(ctx context.Context, c *client.Client, projectID, id string) retry.StateRefreshFunc {
	transientErrors := 0
	return func() (interface{}, string, error) {
		resp := &dto.BackupKubernetesRestorePointResponse{}
		httpResp, err := c.Get(ctx, client.ApiPath.BackupKubernetesRestorePointWithID(projectID, id), resp, nil)
		if err != nil {
			if httpResp != nil && httpResp.StatusCode == 404 {
				return resp, "deleted", nil
			}
			transientErrors++
			if transientErrors > maxTransientRestorePointErrors {
				return nil, "", err
			}
			return resp, "unknown", nil
		}

		transientErrors = 0
		status := resp.RestorePoint.Status
		if status == "" {
			status = "unknown"
		}
		return resp, status, nil
	}
}
