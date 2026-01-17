package volume

import (
	"context"
	"net/http"
	"terraform-provider-vnpaycloud/vnpaycloud/config"
	"terraform-provider-vnpaycloud/vnpaycloud/dto"
	"terraform-provider-vnpaycloud/vnpaycloud/helper/client"
	"terraform-provider-vnpaycloud/vnpaycloud/util"
	"time"

	"github.com/hashicorp/terraform-plugin-log/tflog"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/retry"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
)

func ResourceBlockStorageVolume() *schema.Resource {
	return &schema.Resource{
		CreateContext: resourceBlockStorageVolumeCreate,
		ReadContext:   resourceBlockStorageVolumeRead,
		UpdateContext: resourceBlockStorageVolumeUpdate,
		DeleteContext: resourceBlockStorageVolumeDelete,
		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},

		Timeouts: &schema.ResourceTimeout{
			Create: schema.DefaultTimeout(10 * time.Minute),
			Delete: schema.DefaultTimeout(10 * time.Minute),
		},

		Schema: map[string]*schema.Schema{
			"region": {
				Type:     schema.TypeString,
				Optional: true,
				Computed: true,
				ForceNew: true,
			},

			"size": {
				Type:     schema.TypeInt,
				Required: true,
			},

			"enable_online_resize": {
				Type:     schema.TypeBool,
				Optional: true,
			},

			"name": {
				Type:     schema.TypeString,
				Optional: true,
				ForceNew: false,
			},

			"description": {
				Type:     schema.TypeString,
				Optional: true,
				ForceNew: false,
			},

			"availability_zone": {
				Type:     schema.TypeString,
				Optional: true,
				ForceNew: true,
				Computed: true,
			},

			"metadata": {
				Type:     schema.TypeMap,
				Optional: true,
				ForceNew: false,
				Computed: true,
			},

			"snapshot_id": {
				Type:          schema.TypeString,
				Optional:      true,
				ForceNew:      true,
				ConflictsWith: []string{"source_vol_id", "image_id", "backup_id"},
			},

			"source_vol_id": {
				Type:          schema.TypeString,
				Optional:      true,
				ForceNew:      true,
				ConflictsWith: []string{"snapshot_id", "image_id", "backup_id"},
			},

			"image_id": {
				Type:          schema.TypeString,
				Optional:      true,
				ForceNew:      true,
				ConflictsWith: []string{"snapshot_id", "source_vol_id", "backup_id"},
			},

			"backup_id": {
				Type:          schema.TypeString,
				Optional:      true,
				ForceNew:      true,
				ConflictsWith: []string{"snapshot_id", "source_vol_id", "image_id"},
			},

			"volume_type": {
				Type:     schema.TypeString,
				Optional: true,
				Computed: true,
			},

			"volume_retype_policy": {
				Type:     schema.TypeString,
				Optional: true,
				Default:  "never",
				ValidateFunc: validation.StringInSlice([]string{
					"never", "on-demand",
				}, true),
			},

			"consistency_group_id": {
				Type:     schema.TypeString,
				Optional: true,
				ForceNew: true,
			},

			"source_replica": {
				Type:     schema.TypeString,
				Optional: true,
				ForceNew: true,
			},

			"attachment": {
				Type:     schema.TypeSet,
				Computed: true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"id": {
							Type:     schema.TypeString,
							Computed: true,
						},
						"instance_id": {
							Type:     schema.TypeString,
							Computed: true,
						},
						"device": {
							Type:     schema.TypeString,
							Computed: true,
						},
					},
				},
				Set: blockStorageVolumeAttachmentHash,
			},

			"scheduler_hints": {
				Type:     schema.TypeSet,
				Optional: true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"different_host": {
							Type:     schema.TypeList,
							Optional: true,
							ForceNew: true,
							Elem:     &schema.Schema{Type: schema.TypeString},
						},
						"same_host": {
							Type:     schema.TypeList,
							Optional: true,
							ForceNew: true,
							Elem:     &schema.Schema{Type: schema.TypeString},
						},
						"query": {
							Type:     schema.TypeString,
							Optional: true,
							ForceNew: true,
						},
						"local_to_instance": {
							Type:     schema.TypeString,
							Optional: true,
							ForceNew: true,
						},
						"additional_properties": {
							Type:     schema.TypeMap,
							Optional: true,
							ForceNew: true,
						},
					},
				},
				Set: blockStorageVolumeSchedulerHintsHash,
			},
		},
	}
}

func resourceBlockStorageVolumeCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	config := meta.(*config.Config)
	blockStorageClient, err := client.NewClient(ctx, config.ConsoleClientConfig)
	if err != nil {
		return diag.Errorf("Error creating VNPAY Cloud block storage client: %s", err)
	}

	metadata := d.Get("metadata").(map[string]interface{})
	createOpts := &dto.CreateVolumeOpts{
		AvailabilityZone:   d.Get("availability_zone").(string),
		ConsistencyGroupID: d.Get("consistency_group_id").(string),
		Description:        d.Get("description").(string),
		ImageID:            d.Get("image_id").(string),
		Metadata:           util.ExpandToMapStringString(metadata),
		Name:               d.Get("name").(string),
		Size:               d.Get("size").(int),
		SnapshotID:         d.Get("snapshot_id").(string),
		SourceReplica:      d.Get("source_replica").(string),
		SourceVolID:        d.Get("source_vol_id").(string),
		VolumeType:         d.Get("volume_type").(string),
	}

	var schedulerHints dto.SchedulerVolumeHintOpts

	schedulerHintsRaw := d.Get("scheduler_hints").(*schema.Set).List()
	if len(schedulerHintsRaw) > 0 {
		tflog.Debug(ctx, "vnpaycloud_blockstorage_volume scheduler hints", map[string]interface{}{"scheduler_hints": schedulerHintsRaw[0]})
		schedulerHints = resourceBlockStorageVolumeSchedulerHints(schedulerHintsRaw[0].(map[string]interface{}))
	}

	createOpts.SchedulerHints, err = schedulerHints.ToSchedulerHintsMap()

	if err != nil {
		return diag.Errorf("Error creating vnpaycloud_blockstorage_volume: %s", err)
	}

	if v := d.Get("backup_id").(string); v != "" {
		createOpts.BackupID = v
	}

	tflog.Debug(ctx, "vnpaycloud_blockstorage_volume create options", map[string]interface{}{"create_opts": createOpts})

	createVolumeReq := dto.CreateVolumeRequest{
		Volume: *createOpts,
	}
	createVolumeResp := &dto.CreateVolumeResponse{}
	_, err = blockStorageClient.Post(ctx, client.ApiPath.Volume(blockStorageClient.GetProjectID()), createVolumeReq, createVolumeResp, nil)
	if err != nil {
		return diag.Errorf("Error creating vnpaycloud_blockstorage_volume: %s", err)
	}

	stateConf := &retry.StateChangeConf{
		Pending:    []string{"downloading", "creating"},
		Target:     []string{"available"},
		Refresh:    blockStorageVolumeStateRefreshFunc(ctx, blockStorageClient, createVolumeResp.Volume.ID),
		Timeout:    d.Timeout(schema.TimeoutCreate),
		Delay:      10 * time.Second,
		MinTimeout: 3 * time.Second,
	}

	_, err = stateConf.WaitForStateContext(ctx)
	if err != nil {
		return diag.Errorf(
			"Error waiting for vnpaycloud_blockstorage_volume %s to become ready: %s", createVolumeResp.Volume.ID, err)
	}

	d.SetId(createVolumeResp.Volume.ID)

	return resourceBlockStorageVolumeRead(ctx, d, meta)
}

func resourceBlockStorageVolumeRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	config := meta.(*config.Config)
	blockStorageClient, err := client.NewClient(ctx, config.ConsoleClientConfig)
	if err != nil {
		return diag.Errorf("Error creating VNPAY Cloud block storage client: %s", err)
	}

	volumeResp := &dto.GetVolumeResponse{}
	_, err = blockStorageClient.Get(ctx, client.ApiPath.VolumeWithId(blockStorageClient.GetProjectID(), d.Id()), volumeResp, nil)
	if err != nil {
		return diag.FromErr(util.CheckNotFound(d, err, "Error retrieving vnpaycloud_blockstorage_volume"))
	}

	tflog.Debug(ctx, "Retrieved vnpaycloud_blockstorage_volume "+d.Id(), map[string]interface{}{"volume": volumeResp.Volume})

	d.Set("size", volumeResp.Volume.Size)
	d.Set("description", volumeResp.Volume.Description)
	d.Set("availability_zone", volumeResp.Volume.AvailabilityZone)
	d.Set("name", volumeResp.Volume.Name)
	d.Set("snapshot_id", volumeResp.Volume.SnapshotID)
	d.Set("backup_id", volumeResp.Volume.BackupID)
	d.Set("source_vol_id", volumeResp.Volume.SourceVolID)
	d.Set("volume_type", volumeResp.Volume.VolumeType)
	d.Set("metadata", volumeResp.Volume.Metadata)
	d.Set("region", util.GetRegion(d, config))

	if _, exists := d.GetOk("volume_retype_policy"); !exists {
		d.Set("volume_retype_policy", "never")
	}

	attachments := flattenBlockStorageVolumeAttachments(volumeResp.Volume.Attachments)
	tflog.Debug(ctx, "vnpaycloud_blockstorage_volume "+d.Id()+" with attachments", map[string]interface{}{"attachments": attachments})
	if err := d.Set("attachment", attachments); err != nil {
		tflog.Error(
			ctx,
			"Unable to set vnpaycloud_blockstorage_volume "+d.Id()+" attachments",
			map[string]interface{}{
				"error":       err,
				"attachments": attachments,
			},
		)
	}

	return nil
}

func resourceBlockStorageVolumeUpdate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	config := meta.(*config.Config)
	blockStorageClient, err := client.NewClient(ctx, config.ConsoleClientConfig)
	if err != nil {
		return diag.Errorf("Error creating VNPAY Cloud block storage client: %s", err)
	}

	name := d.Get("name").(string)
	description := d.Get("description").(string)
	updateOpts := dto.UpdateVolumeOpts{
		Name:        &name,
		Description: &description,
	}

	if d.HasChange("metadata") {
		metadata := d.Get("metadata").(map[string]interface{})
		updateOpts.Metadata = util.ExpandToMapStringString(metadata)
	}

	var v *dto.Volume
	if d.HasChange("size") {
		_, err = blockStorageClient.Get(ctx, client.ApiPath.VolumeWithId(blockStorageClient.GetProjectID(), d.Id()), v, nil)
		if err != nil {
			return diag.Errorf("Error extending vnpaycloud_blockstorage_volume %s: %s", d.Id(), err)
		}

		if v.Status == "in-use" {
			if v, ok := d.Get("enable_online_resize").(bool); ok && !v {
				return diag.Errorf(
					`Error extending vnpaycloud_blockstorage_volume %s,
					volume is attached to the instance and
					resizing online is disabled,
					see enable_online_resize option`, d.Id())
			}
		}

		extendOpts := dto.ExtendVolumeSizeOpts{
			NewSize: d.Get("size").(int),
		}

		extendVolumeReq := dto.ExtendVolumeRequest{
			ExtendSize: extendOpts,
		}

		_, err = blockStorageClient.Post(ctx, client.ApiPath.VolumeAction(blockStorageClient.GetProjectID(), d.Id()), extendVolumeReq, nil, nil)
		if err != nil {
			return diag.Errorf("Error extending vnpaycloud_blockstorage_volume %s size: %s", d.Id(), err)
		}

		stateConf := &retry.StateChangeConf{
			Pending:    []string{"extending"},
			Target:     []string{"available", "in-use"},
			Refresh:    blockStorageVolumeStateRefreshFunc(ctx, blockStorageClient, d.Id()),
			Timeout:    d.Timeout(schema.TimeoutCreate),
			Delay:      10 * time.Second,
			MinTimeout: 3 * time.Second,
		}

		_, err := stateConf.WaitForStateContext(ctx)
		if err != nil {
			return diag.Errorf(
				"Error waiting for vnpaycloud_blockstorage_volume %s to become ready: %s", d.Id(), err)
		}
	}

	if d.HasChange("volume_type") {
		_, err = blockStorageClient.Get(ctx, client.ApiPath.VolumeWithId(blockStorageClient.GetProjectID(), d.Id()), v, nil)
		if err != nil {
			return diag.Errorf("Error changing volume type vnpaycloud_blockstorage_volume %s: %s", d.Id(), err)
		}

		retypeOptions := &dto.ChangeVolumeTypeOpts{
			NewType:         d.Get("volume_type").(string),
			MigrationPolicy: dto.VolumeMigrationPolicy(d.Get("volume_retype_policy").(string)),
		}

		retypeVolumeReq := dto.ChangeTypeRequest{
			ChangeType: *retypeOptions,
		}

		_, err = blockStorageClient.Post(ctx, client.ApiPath.VolumeAction(blockStorageClient.GetProjectID(), d.Id()), retypeVolumeReq, nil, nil)
		if err != nil {
			return diag.Errorf("Error changing volume %s type: %s", d.Id(), err)
		}

		stateConf := &retry.StateChangeConf{
			Pending:    []string{"retyping"},
			Target:     []string{"available", "in-use"},
			Refresh:    blockStorageVolumeStateRefreshFunc(ctx, blockStorageClient, d.Id()),
			Timeout:    d.Timeout(schema.TimeoutUpdate),
			Delay:      10 * time.Second,
			MinTimeout: 3 * time.Second,
		}

		_, err = stateConf.WaitForStateContext(ctx)
		if err != nil {
			return diag.Errorf(
				"Error waiting for vnpaycloud_blockstorage_volume %s to become ready: %s", d.Id(), err)
		}
	}

	updateVolumeReq := dto.UpdateVolumeRequest{
		Volume: updateOpts,
	}

	_, err = blockStorageClient.Put(ctx, client.ApiPath.VolumeWithId(blockStorageClient.GetProjectID(), d.Id()), updateVolumeReq, nil, nil)
	if err != nil {
		return diag.Errorf("Error updating vnpaycloud_blockstorage_volume %s: %s", d.Id(), err)
	}

	return resourceBlockStorageVolumeRead(ctx, d, meta)
}

func resourceBlockStorageVolumeDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	config := meta.(*config.Config)
	blockStorageClient, err := client.NewClient(ctx, config.ConsoleClientConfig)
	if err != nil {
		return diag.Errorf("Error creating VNPAY Cloud block storage client: %s", err)
	}

	volumeResp := &dto.GetVolumeResponse{}
	_, err = blockStorageClient.Get(ctx, client.ApiPath.VolumeWithId(blockStorageClient.GetProjectID(), d.Id()), volumeResp, nil)
	if err != nil {
		return diag.FromErr(util.CheckNotFound(d, err, "Error retrieving vnpaycloud_blockstorage_volume"))
	}

	// make sure this volume is detached from all instances before deleting
	if len(volumeResp.Volume.Attachments) > 0 {
		computeClient, err := client.NewClient(ctx, config.ConsoleClientConfig)
		if err != nil {
			return diag.Errorf("Error creating VNPAY Cloud compute client: %s", err)
		}

		for _, volumeAttachment := range volumeResp.Volume.Attachments {
			tflog.Debug(ctx, "vnpaycloud_blockstorage_volume "+d.Id(), map[string]interface{}{"attachment": volumeAttachment})

			serverID := volumeAttachment.ServerID
			attachmentID := volumeAttachment.ID
			if _, err := computeClient.Delete(ctx, client.ApiPath.VolumeAttachmentWithId(blockStorageClient.GetProjectID(), attachmentID), nil); err != nil {
				// It's possible the volume was already detached by
				// vnpaycloud_compute_volume_attach_v2, so consider
				// a 404 acceptable and continue.
				if util.ResponseCodeIs(err, http.StatusNotFound) {
					continue
				}

				// A 409 is also acceptable because there's another
				// concurrent action happening.
				if util.ResponseCodeIs(err, http.StatusConflict) {
					continue
				}

				return diag.Errorf(
					"Error detaching vnpaycloud_blockstorage_volume %s from %s: %s", d.Id(), serverID, err)
			}
		}

		stateConf := &retry.StateChangeConf{
			Pending:    []string{"in-use", "attaching", "detaching"},
			Target:     []string{"available", "deleted"},
			Refresh:    blockStorageVolumeStateRefreshFunc(ctx, blockStorageClient, volumeResp.Volume.ID),
			Timeout:    10 * time.Minute,
			Delay:      10 * time.Second,
			MinTimeout: 3 * time.Second,
		}

		_, err = stateConf.WaitForStateContext(ctx)
		if err != nil {
			return diag.Errorf(
				"Error waiting for vnpaycloud_blockstorage_volume %s to become available: %s", d.Id(), err)
		}
	}

	// It's possible that this volume was used as a boot device and is currently
	// in a "deleting" state from when the instance was terminated.
	// If this is true, just move on. It'll eventually delete.
	if volumeResp.Volume.Status != "deleting" {
		if _, err := blockStorageClient.Delete(ctx, client.ApiPath.VolumeWithId(blockStorageClient.GetProjectID(), d.Id()), nil); err != nil {
			return diag.FromErr(util.CheckNotFound(d, err, "Error deleting vnpaycloud_blockstorage_volume"))
		}
	}

	stateConf := &retry.StateChangeConf{
		Pending:    []string{"deleting", "downloading", "available"},
		Target:     []string{"deleted"},
		Refresh:    blockStorageVolumeStateRefreshFunc(ctx, blockStorageClient, volumeResp.Volume.ID),
		Timeout:    d.Timeout(schema.TimeoutDelete),
		Delay:      10 * time.Second,
		MinTimeout: 3 * time.Second,
	}

	_, err = stateConf.WaitForStateContext(ctx)
	if err != nil {
		return diag.Errorf("Error waiting for vnpaycloud_blockstorage_volume %s to Delete:  %s", d.Id(), err)
	}

	return nil
}
