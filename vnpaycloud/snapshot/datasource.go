package snapshot

import (
	"context"
	"fmt"
	"terraform-provider-vnpaycloud/vnpaycloud/config"
	"terraform-provider-vnpaycloud/vnpaycloud/dto"
	"terraform-provider-vnpaycloud/vnpaycloud/helper/client"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

func DataSourceSnapshot() *schema.Resource {
	return &schema.Resource{
		ReadContext: dataSourceSnapshotRead,
		Schema: map[string]*schema.Schema{
			"id": {
				Type:     schema.TypeString,
				Optional: true,
				Computed: true,
			},
			"name": {
				Type:     schema.TypeString,
				Optional: true,
				Computed: true,
			},
			"description": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"volume_id": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"size": {
				Type:     schema.TypeInt,
				Computed: true,
			},
			"status": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"created_at": {
				Type:     schema.TypeString,
				Computed: true,
			},
		},
	}
}

func dataSourceSnapshotRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	cfg := meta.(*config.Config)

	if id, ok := d.GetOk("id"); ok {
		snapResp := &dto.SnapshotResponse{}
		_, err := cfg.Client.Get(ctx, client.ApiPath.SnapshotWithID(cfg.ProjectID, id.(string)), snapResp, nil)
		if err != nil {
			return diag.Errorf("Error fetching vnpaycloud_snapshot %s: %s", id, err)
		}
		return setSnapshotData(d, &snapResp.Snapshot)
	}

	listResp := &dto.ListSnapshotsResponse{}
	_, err := cfg.Client.Get(ctx, client.ApiPath.Snapshots(cfg.ProjectID), listResp, nil)
	if err != nil {
		return diag.Errorf("Error listing vnpaycloud_snapshot: %s", err)
	}

	nameFilter, nameOk := d.GetOk("name")

	for _, snap := range listResp.Snapshots {
		if nameOk && snap.Name != nameFilter.(string) {
			continue
		}
		return setSnapshotData(d, &snap)
	}

	return diag.Errorf("No vnpaycloud_snapshot found matching the criteria")
}

func setSnapshotData(d *schema.ResourceData, snap *dto.Snapshot) diag.Diagnostics {
	d.SetId(snap.ID)
	d.Set("name", snap.Name)
	d.Set("description", snap.Description)
	d.Set("volume_id", snap.VolumeID)
	d.Set("size", snap.SizeGB)
	d.Set("status", snap.Status)
	d.Set("created_at", snap.CreatedAt)
	return nil
}

func DataSourceSnapshots() *schema.Resource {
	return &schema.Resource{
		ReadContext: dataSourceSnapshotsRead,
		Schema: map[string]*schema.Schema{
			"snapshots": {
				Type:     schema.TypeList,
				Computed: true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"id":          {Type: schema.TypeString, Computed: true},
						"name":        {Type: schema.TypeString, Computed: true},
						"description": {Type: schema.TypeString, Computed: true},
						"volume_id":   {Type: schema.TypeString, Computed: true},
						"size":        {Type: schema.TypeInt, Computed: true},
						"status":      {Type: schema.TypeString, Computed: true},
						"created_at":  {Type: schema.TypeString, Computed: true},
					},
				},
			},
		},
	}
}

func dataSourceSnapshotsRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	cfg := meta.(*config.Config)

	listResp := &dto.ListSnapshotsResponse{}
	_, err := cfg.Client.Get(ctx, client.ApiPath.Snapshots(cfg.ProjectID), listResp, nil)
	if err != nil {
		return diag.Errorf("Error listing vnpaycloud_snapshots: %s", err)
	}

	var snapshots []map[string]interface{}
	for _, snap := range listResp.Snapshots {
		snapshots = append(snapshots, map[string]interface{}{
			"id":          snap.ID,
			"name":        snap.Name,
			"description": snap.Description,
			"volume_id":   snap.VolumeID,
			"size":        snap.SizeGB,
			"status":      snap.Status,
			"created_at":  snap.CreatedAt,
		})
	}

	d.SetId(fmt.Sprintf("snapshots-%s", cfg.ProjectID))
	d.Set("snapshots", snapshots)

	return nil
}
