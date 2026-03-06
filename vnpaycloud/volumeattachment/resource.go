package volumeattachment

import (
	"context"
	"fmt"
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
)

func ResourceVolumeAttachment() *schema.Resource {
	return &schema.Resource{
		CreateContext: resourceVolumeAttachmentCreate,
		ReadContext:   resourceVolumeAttachmentRead,
		DeleteContext: resourceVolumeAttachmentDelete,
		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},
		Timeouts: &schema.ResourceTimeout{
			Create: schema.DefaultTimeout(10 * time.Minute),
			Delete: schema.DefaultTimeout(10 * time.Minute),
		},
		Schema: map[string]*schema.Schema{
			"volume_id": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},
			"server_id": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},
			"device": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"status": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"attached_at": {
				Type:     schema.TypeString,
				Computed: true,
			},
		},
	}
}

func resourceVolumeAttachmentCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	cfg := meta.(*config.Config)

	volumeID := d.Get("volume_id").(string)
	serverID := d.Get("server_id").(string)

	attachOpts := dto.AttachVolumeRequest{
		ServerID: serverID,
	}

	tflog.Debug(ctx, "vnpaycloud_volume_attachment create options", map[string]interface{}{
		"volume_id": volumeID,
		"server_id": serverID,
	})

	attachResp := &dto.VolumeAttachmentResponse{}
	_, err := cfg.Client.Post(ctx, client.ApiPath.VolumeAttach(cfg.ProjectID, volumeID), attachOpts, attachResp, nil)
	if err != nil {
		return diag.Errorf("Error attaching volume %s to server %s: %s", volumeID, serverID, err)
	}

	d.SetId(attachResp.Attachment.ID)

	stateConf := &retry.StateChangeConf{
		Pending:    []string{"initiating", "attaching"},
		Target:     []string{"active", "attached", "in-use"},
		Refresh:    volumeAttachmentStateRefreshFunc(ctx, cfg.Client, cfg.ProjectID, attachResp.Attachment.ID),
		Timeout:    d.Timeout(schema.TimeoutCreate),
		Delay:      5 * time.Second,
		MinTimeout: 3 * time.Second,
	}

	_, err = stateConf.WaitForStateContext(ctx)
	if err != nil {
		return diag.Errorf("Error waiting for volume attachment %s to become ready: %s", attachResp.Attachment.ID, err)
	}

	return resourceVolumeAttachmentRead(ctx, d, meta)
}

func resourceVolumeAttachmentRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	cfg := meta.(*config.Config)

	attachResp := &dto.VolumeAttachmentResponse{}
	_, err := cfg.Client.Get(ctx, client.ApiPath.VolumeAttachmentWithID(cfg.ProjectID, d.Id()), attachResp, nil)
	if err != nil {
		return diag.FromErr(util.CheckNotFound(d, err, "Error retrieving vnpaycloud_volume_attachment"))
	}

	tflog.Debug(ctx, "Retrieved vnpaycloud_volume_attachment "+d.Id(), map[string]interface{}{"attachment": attachResp.Attachment})

	d.Set("volume_id", attachResp.Attachment.VolumeID)
	d.Set("server_id", attachResp.Attachment.ServerID)
	d.Set("device", attachResp.Attachment.Device)
	d.Set("status", attachResp.Attachment.Status)
	d.Set("attached_at", attachResp.Attachment.AttachedAt)

	return nil
}

func resourceVolumeAttachmentDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	cfg := meta.(*config.Config)

	volumeID := d.Get("volume_id").(string)
	serverID := d.Get("server_id").(string)

	detachOpts := dto.DetachVolumeRequest{
		ServerID: serverID,
	}

	tflog.Debug(ctx, "vnpaycloud_volume_attachment delete", map[string]interface{}{
		"volume_id": volumeID,
		"server_id": serverID,
	})

	_, err := cfg.Client.Post(ctx, client.ApiPath.VolumeDetach(cfg.ProjectID, volumeID), detachOpts, nil, nil)
	if err != nil {
		if util.ResponseCodeIs(err, http.StatusNotFound) {
			return nil
		}
		return diag.Errorf("Error detaching volume %s from server %s: %s", volumeID, serverID, err)
	}

	stateConf := &retry.StateChangeConf{
		Pending:    []string{"detaching", "active", "attached", "in-use"},
		Target:     []string{"detached", "deleted"},
		Refresh:    volumeAttachmentStateRefreshFunc(ctx, cfg.Client, cfg.ProjectID, d.Id()),
		Timeout:    d.Timeout(schema.TimeoutDelete),
		Delay:      5 * time.Second,
		MinTimeout: 3 * time.Second,
	}

	_, err = stateConf.WaitForStateContext(ctx)
	if err != nil {
		return diag.Errorf("Error waiting for volume attachment %s to be detached: %s", d.Id(), err)
	}

	return nil
}

func volumeAttachmentStateRefreshFunc(ctx context.Context, c *client.Client, projectID, attachmentID string) retry.StateRefreshFunc {
	return func() (interface{}, string, error) {
		attachResp := &dto.VolumeAttachmentResponse{}
		_, err := c.Get(ctx, client.ApiPath.VolumeAttachmentWithID(projectID, attachmentID), attachResp, nil)

		if err != nil {
			if util.ResponseCodeIs(err, http.StatusNotFound) {
				return attachResp.Attachment, "deleted", nil
			}
			return nil, "", err
		}

		if attachResp.Attachment.Status == "failed" {
			return attachResp.Attachment, attachResp.Attachment.Status, fmt.Errorf("The volume attachment is in error status. " +
				"Please check with your cloud admin or check the API logs.")
		}

		return attachResp.Attachment, attachResp.Attachment.Status, nil
	}
}
