package snapshot

import (
	"context"
	"terraform-provider-vnpaycloud/vnpaycloud/config"
	"terraform-provider-vnpaycloud/vnpaycloud/dto"
	"terraform-provider-vnpaycloud/vnpaycloud/helper/client"
	"terraform-provider-vnpaycloud/vnpaycloud/util"
	"time"

	"github.com/hashicorp/terraform-plugin-log/tflog"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/retry"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

func ResourceSnapshot() *schema.Resource {
	return &schema.Resource{
		CreateContext: resourceSnapshotCreate,
		ReadContext:   resourceSnapshotRead,
		DeleteContext: resourceSnapshotDelete,
		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},
		Timeouts: &schema.ResourceTimeout{
			Create: schema.DefaultTimeout(10 * time.Minute),
			Delete: schema.DefaultTimeout(10 * time.Minute),
		},
		Schema: map[string]*schema.Schema{
			"name": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},
			"description": {
				Type:     schema.TypeString,
				Optional: true,
				ForceNew: true,
			},
			"volume_id": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
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

func resourceSnapshotCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	cfg := meta.(*config.Config)

	createOpts := dto.CreateSnapshotRequest{
		Name:        d.Get("name").(string),
		Description: d.Get("description").(string),
		VolumeID:    d.Get("volume_id").(string),
	}

	tflog.Debug(ctx, "vnpaycloud_snapshot create options", map[string]interface{}{"create_opts": createOpts})

	createResp := &dto.SnapshotResponse{}
	_, err := cfg.Client.Post(ctx, client.ApiPath.Snapshots(cfg.ProjectID), createOpts, createResp, nil)
	if err != nil {
		return diag.Errorf("Error creating vnpaycloud_snapshot: %s", err)
	}

	d.SetId(createResp.Snapshot.ID)

	stateConf := &retry.StateChangeConf{
		Pending:    []string{"initiating", "creating"},
		Target:     []string{"active", "created", "available"},
		Refresh:    snapshotStateRefreshFunc(ctx, cfg.Client, cfg.ProjectID, createResp.Snapshot.ID),
		Timeout:    d.Timeout(schema.TimeoutCreate),
		Delay:      5 * time.Second,
		MinTimeout: 3 * time.Second,
	}

	_, err = stateConf.WaitForStateContext(ctx)
	if err != nil {
		return diag.Errorf("Error waiting for vnpaycloud_snapshot %s to become ready: %s", createResp.Snapshot.ID, err)
	}

	return resourceSnapshotRead(ctx, d, meta)
}

func resourceSnapshotRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	cfg := meta.(*config.Config)

	snapResp := &dto.SnapshotResponse{}
	_, err := cfg.Client.Get(ctx, client.ApiPath.SnapshotWithID(cfg.ProjectID, d.Id()), snapResp, nil)
	if err != nil {
		return diag.FromErr(util.CheckNotFound(d, err, "Error retrieving vnpaycloud_snapshot"))
	}

	tflog.Debug(ctx, "Retrieved vnpaycloud_snapshot "+d.Id(), map[string]interface{}{"snapshot": snapResp.Snapshot})

	d.Set("name", snapResp.Snapshot.Name)
	d.Set("description", snapResp.Snapshot.Description)
	d.Set("volume_id", snapResp.Snapshot.VolumeID)
	d.Set("size", snapResp.Snapshot.SizeGB)
	d.Set("status", snapResp.Snapshot.Status)
	d.Set("created_at", snapResp.Snapshot.CreatedAt)

	return nil
}

func resourceSnapshotDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	cfg := meta.(*config.Config)

	snapResp := &dto.SnapshotResponse{}
	_, err := cfg.Client.Get(ctx, client.ApiPath.SnapshotWithID(cfg.ProjectID, d.Id()), snapResp, nil)
	if err != nil {
		return diag.FromErr(util.CheckDeleted(d, err, "Error retrieving vnpaycloud_snapshot"))
	}

	if snapResp.Snapshot.Status != "deleting" {
		if _, err := cfg.Client.Delete(ctx, client.ApiPath.SnapshotWithID(cfg.ProjectID, d.Id()), nil); err != nil {
			return diag.FromErr(util.CheckDeleted(d, err, "Error deleting vnpaycloud_snapshot"))
		}
	}

	stateConf := &retry.StateChangeConf{
		Pending:    []string{"deleting", "active", "available"},
		Target:     []string{"deleted"},
		Refresh:    snapshotStateRefreshFunc(ctx, cfg.Client, cfg.ProjectID, d.Id()),
		Timeout:    d.Timeout(schema.TimeoutDelete),
		Delay:      5 * time.Second,
		MinTimeout: 3 * time.Second,
	}

	_, err = stateConf.WaitForStateContext(ctx)
	if err != nil {
		return diag.Errorf("Error waiting for vnpaycloud_snapshot %s to delete: %s", d.Id(), err)
	}

	return nil
}
