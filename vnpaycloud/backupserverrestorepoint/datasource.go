package backupserverrestorepoint

import (
	"context"
	"net/url"
	"strconv"
	"terraform-provider-vnpaycloud/vnpaycloud/config"
	"terraform-provider-vnpaycloud/vnpaycloud/dto"
	"terraform-provider-vnpaycloud/vnpaycloud/helper/client"
	"terraform-provider-vnpaycloud/vnpaycloud/util"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

func computedBackupServerRestorePointAttrs() map[string]*schema.Schema {
	return map[string]*schema.Schema{
		"id":                   {Type: schema.TypeString, Computed: true},
		"name":                 {Type: schema.TypeString, Computed: true},
		"backup_server_id":     {Type: schema.TypeString, Computed: true},
		"backup_policy_id":     {Type: schema.TypeString, Computed: true},
		"type":                 {Type: schema.TypeString, Computed: true},
		"purpose":              {Type: schema.TypeString, Computed: true},
		"zone_id":              {Type: schema.TypeString, Computed: true},
		"is_visible":           {Type: schema.TypeBool, Computed: true},
		"create_by_backup_now": {Type: schema.TypeBool, Computed: true},
		"backup_point":         {Type: schema.TypeString, Computed: true},
		"status":               {Type: schema.TypeString, Computed: true},
		"created_at":           {Type: schema.TypeString, Computed: true},
	}
}

func DataSourceBackupServerRestorePoints() *schema.Resource {
	return &schema.Resource{
		ReadContext: dataSourceBackupServerRestorePointsRead,
		Schema: map[string]*schema.Schema{
			"backup_server_id": {
				Type:     schema.TypeString,
				Optional: true,
			},
			"name": {
				Type:     schema.TypeString,
				Optional: true,
			},
			"restore_points": {
				Type:     schema.TypeList,
				Computed: true,
				Elem:     &schema.Resource{Schema: computedBackupServerRestorePointAttrs()},
			},
		},
	}
}

func dataSourceBackupServerRestorePointsRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	cfg := meta.(*config.Config)

	path := client.ApiPath.BackupServerRestorePoints(cfg.ProjectID)
	q := url.Values{}
	if v, ok := d.GetOk("backup_server_id"); ok {
		q.Set("backupServerId", v.(string))
	}
	if v, ok := d.GetOk("name"); ok {
		q.Set("name", v.(string))
	}
	if len(q) > 0 {
		path += "?" + q.Encode()
	}

	listResp := &dto.ListBackupServerRestorePointsResponse{}
	if _, err := cfg.Client.Get(ctx, path, listResp, nil); err != nil {
		return diag.Errorf("Error listing vnpaycloud_backup_server_restore_points: %s", err)
	}

	points := make([]map[string]interface{}, 0, len(listResp.RestorePoints))
	for _, rp := range listResp.RestorePoints {
		points = append(points, map[string]interface{}{
			"id":                   rp.ID,
			"name":                 rp.Name,
			"backup_server_id":     rp.BackupServerID,
			"backup_policy_id":     rp.BackupPolicyID,
			"type":                 rp.Type,
			"purpose":              rp.Purpose,
			"zone_id":              rp.ZoneID,
			"is_visible":           rp.IsVisible,
			"create_by_backup_now": rp.CreateByBackupNow,
			"backup_point":         rp.BackupPoint,
			"status":               rp.Status,
			"created_at":           rp.CreatedAt,
		})
	}

	util.SortNewestFirst(points, "created_at", "backup_point")

	if err := d.Set("restore_points", points); err != nil {
		return diag.FromErr(err)
	}
	d.SetId(strconv.Itoa(len(points)) + "-restore-points")
	return nil
}

func computedBackupServerDisasterRestorePointAttrs() map[string]*schema.Schema {
	return map[string]*schema.Schema{
		"id":               {Type: schema.TypeString, Computed: true},
		"name":             {Type: schema.TypeString, Computed: true},
		"backup_server_id": {Type: schema.TypeString, Computed: true},
		"backup_policy_id": {Type: schema.TypeString, Computed: true},
		"type":             {Type: schema.TypeString, Computed: true},
		"purpose":          {Type: schema.TypeString, Computed: true},
		"zone_id":          {Type: schema.TypeString, Computed: true},
		"backup_point":     {Type: schema.TypeString, Computed: true},
		"status":           {Type: schema.TypeString, Computed: true},
		"created_at":       {Type: schema.TypeString, Computed: true},
		"server_id":        {Type: schema.TypeString, Computed: true},
		"server_name":      {Type: schema.TypeString, Computed: true},
		"server_zone_id":   {Type: schema.TypeString, Computed: true},
	}
}

func DataSourceBackupServerDisasterRestorePoints() *schema.Resource {
	return &schema.Resource{
		ReadContext: dataSourceBackupServerDisasterRestorePointsRead,
		Schema: map[string]*schema.Schema{
			"backup_server_id": {
				Type:     schema.TypeString,
				Optional: true,
			},
			"name": {
				Type:     schema.TypeString,
				Optional: true,
			},
			"restore_points": {
				Type:     schema.TypeList,
				Computed: true,
				Elem:     &schema.Resource{Schema: computedBackupServerDisasterRestorePointAttrs()},
			},
		},
	}
}

func dataSourceBackupServerDisasterRestorePointsRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	cfg := meta.(*config.Config)

	path := client.ApiPath.BackupServerDisasterRestorePoints(cfg.ProjectID)
	q := url.Values{}
	if v, ok := d.GetOk("backup_server_id"); ok {
		q.Set("backupServerId", v.(string))
	}
	if v, ok := d.GetOk("name"); ok {
		q.Set("name", v.(string))
	}
	if len(q) > 0 {
		path += "?" + q.Encode()
	}

	listResp := &dto.ListBackupServerDisasterRestorePointsResponse{}
	if _, err := cfg.Client.Get(ctx, path, listResp, nil); err != nil {
		return diag.Errorf("Error listing vnpaycloud_backup_server_disaster_restore_points: %s", err)
	}

	points := make([]map[string]interface{}, 0, len(listResp.RestorePoints))
	for _, rp := range listResp.RestorePoints {
		points = append(points, map[string]interface{}{
			"id":               rp.ID,
			"name":             rp.Name,
			"backup_server_id": rp.BackupServerID,
			"backup_policy_id": rp.BackupPolicyID,
			"type":             rp.Type,
			"purpose":          rp.Purpose,
			"zone_id":          rp.ZoneID,
			"backup_point":     rp.BackupPoint,
			"status":           rp.Status,
			"created_at":       rp.CreatedAt,
			"server_id":        rp.ServerID,
			"server_name":      rp.ServerName,
			"server_zone_id":   rp.ServerZoneID,
		})
	}

	util.SortNewestFirst(points, "created_at", "backup_point")

	if err := d.Set("restore_points", points); err != nil {
		return diag.FromErr(err)
	}
	d.SetId(strconv.Itoa(len(points)) + "-disaster-restore-points")
	return nil
}
