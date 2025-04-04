package volumev3

import (
	"context"
	"log"
	"terraform-provider-vnpaycloud/vnpaycloud/config"
	"terraform-provider-vnpaycloud/vnpaycloud/util"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"

	"github.com/vnpaycloud-console/gophercloud/v2/openstack/blockstorage/v3/volumes"
)

func DataSourceBlockStorageVolumeV3() *schema.Resource {
	return &schema.Resource{
		ReadContext: dataSourceBlockStorageVolumeV3Read,

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
				Set: blockStorageVolumeV3AttachmentHash,
			},
		},
	}
}

func dataSourceBlockStorageVolumeV3Read(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	config := meta.(*config.Config)
	client, err := config.BlockStorageV3Client(ctx, util.GetRegion(d, config))
	if err != nil {
		return diag.Errorf("Error creating VNPAY Cloud block storage client: %s", err)
	}

	listOpts := volumes.ListOpts{
		Metadata: util.ExpandToMapStringString(d.Get("metadata").(map[string]interface{})),
		Name:     d.Get("name").(string),
		Status:   d.Get("status").(string),
	}

	allPages, err := volumes.List(client, listOpts).AllPages(ctx)
	if err != nil {
		return diag.Errorf("Unable to query vnpaycloud_blockstorage_volume_v3: %s", err)
	}

	var allVolumes []volumes.Volume
	err = volumes.ExtractVolumesInto(allPages, &allVolumes)
	if err != nil {
		return diag.Errorf("Unable to retrieve vnpaycloud_blockstorage_volume_v3: %s", err)
	}

	if len(allVolumes) > 1 {
		return diag.Errorf("Your vnpaycloud_blockstorage_volume_v3 query returned multiple results")
	}

	if len(allVolumes) < 1 {
		return diag.Errorf("Your vnpaycloud_blockstorage_volume_v3 query returned no results")
	}

	dataSourceBlockStorageVolumeV3Attributes(d, allVolumes[0])

	return nil
}

func dataSourceBlockStorageVolumeV3Attributes(d *schema.ResourceData, volume volumes.Volume) {
	d.SetId(volume.ID)
	d.Set("name", volume.Name)
	d.Set("status", volume.Status)
	d.Set("bootable", volume.Bootable)
	d.Set("volume_type", volume.VolumeType)
	d.Set("size", volume.Size)
	d.Set("source_volume_id", volume.SourceVolID)
	d.Set("host", volume.Host)

	if err := d.Set("metadata", volume.Metadata); err != nil {
		log.Printf("[DEBUG] Unable to set metadata for vnpaycloud_blockstorage_volume_v3 %s: %s", volume.ID, err)
	}

	attachments := flattenBlockStorageVolumeV3Attachments(volume.Attachments)
	log.Printf("[DEBUG] vnpaycloud_blockstorage_volume_v3 %s attachments: %#v", d.Id(), attachments)
	if err := d.Set("attachment", attachments); err != nil {
		log.Printf(
			"[DEBUG] unable to set vnpaycloud_blockstorage_volume_v3 %s attachments: %s", d.Id(), err)
	}
}
