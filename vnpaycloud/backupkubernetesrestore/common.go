package backupkubernetesrestore

import (
	"context"

	"terraform-provider-vnpaycloud/vnpaycloud/dto"
	"terraform-provider-vnpaycloud/vnpaycloud/helper/client"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/retry"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

const maxTransientRestoreErrors = 20

// clusterRestoreStateRefreshFunc polls the destination cluster's restore_status
// (none/restoring/success/failed) so the restore Create can wait for completion.
func clusterRestoreStateRefreshFunc(ctx context.Context, c *client.Client, projectID, clusterID string) retry.StateRefreshFunc {
	transientErrors := 0
	return func() (interface{}, string, error) {
		resp := &dto.K8sClusterResponse{}
		if _, err := c.Get(ctx, client.ApiPath.ClusterWithID(projectID, clusterID), resp, nil); err != nil {
			transientErrors++
			if transientErrors > maxTransientRestoreErrors {
				return nil, "", err
			}
			return resp, "unknown", nil
		}

		transientErrors = 0
		status := resp.Cluster.RestoreStatus
		if status == "" {
			status = "none"
		}
		return resp, status, nil
	}
}

func restorePointComputedSchema() map[string]*schema.Schema {
	return map[string]*schema.Schema{
		"id":                   {Type: schema.TypeString, Computed: true},
		"name":                 {Type: schema.TypeString, Computed: true},
		"cluster_id":           {Type: schema.TypeString, Computed: true},
		"cluster_name":         {Type: schema.TypeString, Computed: true},
		"cluster_backup_id":    {Type: schema.TypeString, Computed: true},
		"pvc_backup_policy_id": {Type: schema.TypeString, Computed: true},
		"zone_id":              {Type: schema.TypeString, Computed: true},
		"backup_vault_ids":     {Type: schema.TypeList, Computed: true, Elem: &schema.Schema{Type: schema.TypeString}},
		"velero_backup_name":   {Type: schema.TypeString, Computed: true},
		"backup_point":         {Type: schema.TypeString, Computed: true},
		"kubernetes_version":   {Type: schema.TypeString, Computed: true},
		"status":               {Type: schema.TypeString, Computed: true},
	}
}

func restorePointAttrs(p *dto.BackupKubernetesRestorePoint) map[string]interface{} {
	return map[string]interface{}{
		"id":                   p.ID,
		"name":                 p.Name,
		"cluster_id":           p.ClusterID,
		"cluster_name":         p.ClusterName,
		"cluster_backup_id":    p.ClusterBackupID,
		"pvc_backup_policy_id": p.PVCBackupPolicyID,
		"zone_id":              p.ZoneID,
		"backup_vault_ids":     p.BackupVaultIDs,
		"velero_backup_name":   p.VeleroBackupName,
		"backup_point":         p.BackupPoint,
		"kubernetes_version":   p.KubernetesVersion,
		"status":               p.Status,
	}
}
