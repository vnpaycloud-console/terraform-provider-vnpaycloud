package volume

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

func ResourceVolume() *schema.Resource {
	return &schema.Resource{
		CreateContext: resourceVolumeCreate,
		ReadContext:   resourceVolumeRead,
		UpdateContext: resourceVolumeUpdate,
		DeleteContext: resourceVolumeDelete,
		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},
		Timeouts: &schema.ResourceTimeout{
			Create: schema.DefaultTimeout(10 * time.Minute),
			Update: schema.DefaultTimeout(10 * time.Minute),
			Delete: schema.DefaultTimeout(10 * time.Minute),
		},
		Schema: map[string]*schema.Schema{
			"name": {
				Type:     schema.TypeString,
				Required: true,
			},
			"description": {
				Type:     schema.TypeString,
				Optional: true,
			},
			"size": {
				Type:     schema.TypeInt,
				Required: true,
			},
			"volume_type": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},
			"zone": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"encrypt": {
				Type:     schema.TypeBool,
				Optional: true,
				ForceNew: true,
			},
			"multiattach": {
				Type:     schema.TypeBool,
				Optional: true,
				ForceNew: true,
			},
			"snapshot_id": {
				Type:     schema.TypeString,
				Optional: true,
				ForceNew: true,
			},
			"status": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"iops": {
				Type:     schema.TypeInt,
				Computed: true,
			},
			"is_encrypted": {
				Type:     schema.TypeBool,
				Computed: true,
			},
			"is_multiattach": {
				Type:     schema.TypeBool,
				Computed: true,
			},
			"is_bootable": {
				Type:     schema.TypeBool,
				Computed: true,
			},
			"attached_server_id": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"attached_server_name": {
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

func resourceVolumeCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	cfg := meta.(*config.Config)

	createOpts := dto.CreateVolumeRequest{
		Name:        d.Get("name").(string),
		Description: d.Get("description").(string),
		SizeGB:      int64(d.Get("size").(int)),
		VolumeType:  d.Get("volume_type").(string),
		Encrypt:     d.Get("encrypt").(bool),
		Multiattach: d.Get("multiattach").(bool),
		SnapshotID:  d.Get("snapshot_id").(string),
	}

	tflog.Debug(ctx, "vnpaycloud_volume create options", map[string]interface{}{"create_opts": createOpts})

	createResp := &dto.VolumeResponse{}
	_, err := cfg.Client.Post(ctx, client.ApiPath.Volumes(cfg.ProjectID), createOpts, createResp, nil)
	if err != nil {
		return diag.Errorf("Error creating vnpaycloud_volume: %s", err)
	}

	d.SetId(createResp.Volume.ID)

	stateConf := &retry.StateChangeConf{
		Pending:    []string{"initiating", "creating"},
		Target:     []string{"active", "created", "available", "in-use"},
		Refresh:    volumeStateRefreshFunc(ctx, cfg.Client, cfg.ProjectID, createResp.Volume.ID),
		Timeout:    d.Timeout(schema.TimeoutCreate),
		Delay:      5 * time.Second,
		MinTimeout: 3 * time.Second,
	}

	_, err = stateConf.WaitForStateContext(ctx)
	if err != nil {
		return diag.Errorf("Error waiting for vnpaycloud_volume %s to become ready: %s", createResp.Volume.ID, err)
	}

	return resourceVolumeRead(ctx, d, meta)
}

func resourceVolumeRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	cfg := meta.(*config.Config)

	volResp := &dto.VolumeResponse{}
	_, err := cfg.Client.Get(ctx, client.ApiPath.VolumeWithID(cfg.ProjectID, d.Id()), volResp, nil)
	if err != nil {
		return diag.FromErr(util.CheckNotFound(d, err, "Error retrieving vnpaycloud_volume"))
	}

	tflog.Debug(ctx, "Retrieved vnpaycloud_volume "+d.Id(), map[string]interface{}{"volume": volResp.Volume})

	d.Set("name", volResp.Volume.Name)
	d.Set("description", volResp.Volume.Description)
	d.Set("size", volResp.Volume.SizeGB)
	d.Set("volume_type", volResp.Volume.VolumeType)
	d.Set("zone", volResp.Volume.Zone)
	d.Set("status", volResp.Volume.Status)
	d.Set("iops", volResp.Volume.IOPS)
	d.Set("is_encrypted", volResp.Volume.IsEncrypted)
	d.Set("is_multiattach", volResp.Volume.IsMultiattach)
	d.Set("is_bootable", volResp.Volume.IsBootable)
	d.Set("attached_server_id", volResp.Volume.AttachedServerID)
	d.Set("attached_server_name", volResp.Volume.AttachedServerName)
	d.Set("created_at", volResp.Volume.CreatedAt)

	return nil
}

func resourceVolumeUpdate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	cfg := meta.(*config.Config)

	if d.HasChanges("name", "description") {
		updateOpts := dto.UpdateVolumeRequest{
			Name:        d.Get("name").(string),
			Description: d.Get("description").(string),
		}

		tflog.Debug(ctx, "vnpaycloud_volume update options", map[string]interface{}{"update_opts": updateOpts})

		_, err := cfg.Client.Put(ctx, client.ApiPath.VolumeWithID(cfg.ProjectID, d.Id()), updateOpts, nil, nil)
		if err != nil {
			return diag.Errorf("Error updating vnpaycloud_volume %s: %s", d.Id(), err)
		}
	}

	if d.HasChange("size") {
		oldRaw, newRaw := d.GetChange("size")
		oldSize := oldRaw.(int)
		newSize := newRaw.(int)

		if newSize < oldSize {
			return diag.Errorf("Error resizing vnpaycloud_volume %s: cannot shrink volume from %d GB to %d GB", d.Id(), oldSize, newSize)
		}

		resizeOpts := dto.ResizeVolumeRequest{
			SizeGB: int64(newSize),
		}

		tflog.Debug(ctx, "vnpaycloud_volume resize options", map[string]interface{}{"resize_opts": resizeOpts})

		_, err := cfg.Client.Post(ctx, client.ApiPath.VolumeResize(cfg.ProjectID, d.Id()), resizeOpts, nil, nil)
		if err != nil {
			return diag.Errorf("Error resizing vnpaycloud_volume %s: %s", d.Id(), err)
		}

		stateConf := &retry.StateChangeConf{
			Pending:    []string{"resizing", "extending"},
			Target:     []string{"active", "available", "in-use"},
			Refresh:    volumeStateRefreshFunc(ctx, cfg.Client, cfg.ProjectID, d.Id()),
			Timeout:    d.Timeout(schema.TimeoutUpdate),
			Delay:      5 * time.Second,
			MinTimeout: 3 * time.Second,
		}

		_, err = stateConf.WaitForStateContext(ctx)
		if err != nil {
			return diag.Errorf("Error waiting for vnpaycloud_volume %s to finish resizing: %s", d.Id(), err)
		}
	}

	return resourceVolumeRead(ctx, d, meta)
}

func resourceVolumeDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	cfg := meta.(*config.Config)

	volResp := &dto.VolumeResponse{}
	_, err := cfg.Client.Get(ctx, client.ApiPath.VolumeWithID(cfg.ProjectID, d.Id()), volResp, nil)
	if err != nil {
		return diag.FromErr(util.CheckDeleted(d, err, "Error retrieving vnpaycloud_volume"))
	}

	if volResp.Volume.Status != "deleting" {
		if _, err := cfg.Client.Delete(ctx, client.ApiPath.VolumeWithID(cfg.ProjectID, d.Id()), nil); err != nil {
			return diag.FromErr(util.CheckDeleted(d, err, "Error deleting vnpaycloud_volume"))
		}
	}

	stateConf := &retry.StateChangeConf{
		Pending:    []string{"deleting", "active", "available", "in-use"},
		Target:     []string{"deleted"},
		Refresh:    volumeStateRefreshFunc(ctx, cfg.Client, cfg.ProjectID, d.Id()),
		Timeout:    d.Timeout(schema.TimeoutDelete),
		Delay:      5 * time.Second,
		MinTimeout: 3 * time.Second,
	}

	_, err = stateConf.WaitForStateContext(ctx)
	if err != nil {
		return diag.Errorf("Error waiting for vnpaycloud_volume %s to delete: %s", d.Id(), err)
	}

	return nil
}
