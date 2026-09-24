package backupkubernetesrestore

import (
	"context"
	"net/url"
	"terraform-provider-vnpaycloud/vnpaycloud/config"
	"terraform-provider-vnpaycloud/vnpaycloud/dto"
	"terraform-provider-vnpaycloud/vnpaycloud/helper/client"
	"terraform-provider-vnpaycloud/vnpaycloud/util"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

func DataSourceBackupKubernetesRestorePoints() *schema.Resource {
	return &schema.Resource{
		ReadContext: dataSourceRestorePointsRead,
		Schema: map[string]*schema.Schema{
			"cluster_backup_id":    {Type: schema.TypeString, Optional: true},
			"cluster_id":           {Type: schema.TypeString, Optional: true},
			"pvc_backup_policy_id": {Type: schema.TypeString, Optional: true},
			"name":                 {Type: schema.TypeString, Optional: true},
			"restore_points": {
				Type:     schema.TypeList,
				Computed: true,
				Elem:     &schema.Resource{Schema: restorePointComputedSchema()},
			},
		},
	}
}

func dataSourceRestorePointsRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	cfg := meta.(*config.Config)

	path := client.ApiPath.BackupKubernetesRestorePoints(cfg.ProjectID)
	q := url.Values{}
	if v, ok := d.GetOk("cluster_backup_id"); ok {
		q.Set("clusterBackupId", v.(string))
	}
	if v, ok := d.GetOk("cluster_id"); ok {
		q.Set("clusterId", v.(string))
	}
	if v, ok := d.GetOk("pvc_backup_policy_id"); ok {
		q.Set("pvcBackupPolicyId", v.(string))
	}
	if v, ok := d.GetOk("name"); ok {
		q.Set("name", v.(string))
	}
	if len(q) > 0 {
		path += "?" + q.Encode()
	}

	resp := &dto.ListBackupKubernetesRestorePointsResponse{}
	if _, err := cfg.Client.Get(ctx, path, resp, nil); err != nil {
		return diag.Errorf("Unable to query vnpaycloud_backup_kubernetes_restore_points: %s", err)
	}

	points := make([]map[string]interface{}, 0, len(resp.RestorePoints))
	for i := range resp.RestorePoints {
		points = append(points, restorePointAttrs(&resp.RestorePoints[i]))
	}

	util.SortNewestFirst(points, "backup_point")

	d.SetId(cfg.ProjectID)
	d.Set("restore_points", points)

	return nil
}
