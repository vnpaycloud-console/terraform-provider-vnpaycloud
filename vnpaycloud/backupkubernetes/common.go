package backupkubernetes

import (
	"context"

	"terraform-provider-vnpaycloud/vnpaycloud/dto"
	"terraform-provider-vnpaycloud/vnpaycloud/helper/client"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/retry"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

const maxTransientBackupKubernetesErrors = 20

func setBackupKubernetesData(d *schema.ResourceData, b *dto.BackupKubernetes) {
	d.Set("name", b.Name)
	d.Set("description", b.Description)
	d.Set("protected_cluster_id", b.ProtectedClusterID)
	d.Set("pvc_backup_policy_id", b.PVCBackupPolicyID)
	d.Set("cluster_backup_type", b.ClusterBackupType)
	d.Set("zone_id", b.ZoneID)
	d.Set("project_id", b.ProjectID)
	d.Set("customer_username", b.CustomerUsername)
	d.Set("status", b.Status)
	d.Set("created_at", b.CreatedAt)
}

func backupKubernetesStateRefreshFunc(ctx context.Context, c *client.Client, projectID, id string) retry.StateRefreshFunc {
	transientErrors := 0
	return func() (interface{}, string, error) {
		resp := &dto.BackupKubernetesResponse{}
		httpResp, err := c.Get(ctx, client.ApiPath.BackupKubernetesWithID(projectID, id), resp, nil)
		if err != nil {
			if httpResp != nil && httpResp.StatusCode == 404 {
				return resp, "deleted", nil
			}
			transientErrors++
			if transientErrors > maxTransientBackupKubernetesErrors {
				return nil, "", err
			}
			return resp, "unknown", nil
		}

		transientErrors = 0
		status := resp.BackupKubernetes.Status
		if status == "" {
			status = "unknown"
		}
		return resp, status, nil
	}
}
