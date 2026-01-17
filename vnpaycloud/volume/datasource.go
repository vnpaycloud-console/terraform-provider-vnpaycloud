package volume

import (
	"context"
	"terraform-provider-vnpaycloud/vnpaycloud/config"
	"terraform-provider-vnpaycloud/vnpaycloud/dto"
	"terraform-provider-vnpaycloud/vnpaycloud/helper/client"
	"terraform-provider-vnpaycloud/vnpaycloud/util"

	"github.com/hashicorp/terraform-plugin-log/tflog"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

func DataSourceBlockStorageVolume() *schema.Resource {
	return &schema.Resource{
		ReadContext: dataSourceBlockStorageVolumeRead,

		Schema: map[string]*schema.Schema{
			"region": {
				Type:     schema.TypeString,
				Optional: true,
				Computed: true,
			},

			"name": {
				Type:     schema.TypeString,
				Optional: true,
				Computed: true,
			},

			"status": {
				Type:     schema.TypeString,
				Optional: true,
				Computed: true,
			},

			"metadata": {
				Type:     schema.TypeMap,
				Optional: true,
				Computed: true,
			},

			// Computed values
			"bootable": {
				Type:     schema.TypeString,
				Optional: true,
				Computed: true,
			},

			"volume_type": {
				Type:     schema.TypeString,
				Optional: true,
				Computed: true,
			},

			"size": {
				Type:     schema.TypeInt,
				Computed: true,
			},

			"source_volume_id": {
				Type:     schema.TypeString,
				Computed: true,
			},

			"host": {
				Type:     schema.TypeString,
				Optional: true,
				Computed: true,
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
		},
	}
}

func dataSourceBlockStorageVolumeRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	config := meta.(*config.Config)
	c, err := client.NewClient(ctx, config.ConsoleClientConfig)
	if err != nil {
		return diag.Errorf("Error creating VNPAY Cloud block storage client: %s", err)
	}

	listOpts := dto.ListVolumeParams{
		Metadata: util.ExpandToMapStringString(d.Get("metadata").(map[string]interface{})),
		Name:     d.Get("name").(string),
		Status:   d.Get("status").(string),
	}

	listResp := dto.ListVolumeResponse{}

	_, err = c.Get(ctx, client.ApiPath.VolumeWithParams(c.GetProjectID(), listOpts), &listResp, nil)
	if err != nil {
		return diag.Errorf("Unable to query vnpaycloud_blockstorage_volume: %s", err)
	}

	allVolumes := listResp.Volumes

	if len(allVolumes) > 1 {
		return diag.Errorf("Your vnpaycloud_blockstorage_volume query returned multiple results")
	}

	if len(allVolumes) < 1 {
		return diag.Errorf("Your vnpaycloud_blockstorage_volume query returned no results")
	}

	dataSourceBlockStorageVolumeAttributes(ctx, d, allVolumes[0])

	return nil
}

func dataSourceBlockStorageVolumeAttributes(ctx context.Context, d *schema.ResourceData, volume dto.Volume) {
	d.SetId(volume.ID)
	d.Set("name", volume.Name)
	d.Set("status", volume.Status)
	d.Set("bootable", volume.Bootable)
	d.Set("volume_type", volume.VolumeType)
	d.Set("size", volume.Size)
	d.Set("source_volume_id", volume.SourceVolID)
	d.Set("host", volume.Host)

	if err := d.Set("metadata", volume.Metadata); err != nil {
		tflog.Error(ctx, "Unable to set metadata for vnpaycloud_blockstorage_volume "+volume.ID, map[string]interface{}{"error": err})
	}

	attachments := flattenBlockStorageVolumeAttachments(volume.Attachments)
	tflog.Debug(ctx, "vnpaycloud_blockstorage_volume %"+d.Id()+" attachments", map[string]interface{}{"attachments": attachments})
	if err := d.Set("attachment", attachments); err != nil {
		tflog.Error(
			ctx,
			"unable to set vnpaycloud_blockstorage_volume "+d.Id()+" attachments",
			map[string]interface{}{"error": err},
		)
	}
}
