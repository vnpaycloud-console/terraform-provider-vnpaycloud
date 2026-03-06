package volume

import (
	"context"
	"fmt"
	"terraform-provider-vnpaycloud/vnpaycloud/config"
	"terraform-provider-vnpaycloud/vnpaycloud/dto"
	"terraform-provider-vnpaycloud/vnpaycloud/helper/client"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

func DataSourceVolume() *schema.Resource {
	return &schema.Resource{
		ReadContext: dataSourceVolumeRead,
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
			"size": {
				Type:     schema.TypeInt,
				Computed: true,
			},
			"volume_type": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"zone": {
				Type:     schema.TypeString,
				Computed: true,
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

func dataSourceVolumeRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	cfg := meta.(*config.Config)

	if id, ok := d.GetOk("id"); ok {
		volResp := &dto.VolumeResponse{}
		_, err := cfg.Client.Get(ctx, client.ApiPath.VolumeWithID(cfg.ProjectID, id.(string)), volResp, nil)
		if err != nil {
			return diag.Errorf("Error fetching vnpaycloud_volume %s: %s", id, err)
		}
		return setVolumeData(d, &volResp.Volume)
	}

	listResp := &dto.ListVolumesResponse{}
	_, err := cfg.Client.Get(ctx, client.ApiPath.Volumes(cfg.ProjectID), listResp, nil)
	if err != nil {
		return diag.Errorf("Error listing vnpaycloud_volume: %s", err)
	}

	nameFilter, nameOk := d.GetOk("name")

	for _, vol := range listResp.Volumes {
		if nameOk && vol.Name != nameFilter.(string) {
			continue
		}
		return setVolumeData(d, &vol)
	}

	return diag.Errorf("No vnpaycloud_volume found matching the criteria")
}

func setVolumeData(d *schema.ResourceData, vol *dto.Volume) diag.Diagnostics {
	d.SetId(vol.ID)
	d.Set("name", vol.Name)
	d.Set("description", vol.Description)
	d.Set("size", vol.SizeGB)
	d.Set("volume_type", vol.VolumeType)
	d.Set("zone", vol.Zone)
	d.Set("status", vol.Status)
	d.Set("iops", vol.IOPS)
	d.Set("is_encrypted", vol.IsEncrypted)
	d.Set("is_multiattach", vol.IsMultiattach)
	d.Set("is_bootable", vol.IsBootable)
	d.Set("attached_server_id", vol.AttachedServerID)
	d.Set("attached_server_name", vol.AttachedServerName)
	d.Set("created_at", vol.CreatedAt)
	return nil
}

func DataSourceVolumes() *schema.Resource {
	return &schema.Resource{
		ReadContext: dataSourceVolumesRead,
		Schema: map[string]*schema.Schema{
			"volumes": {
				Type:     schema.TypeList,
				Computed: true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"id":                   {Type: schema.TypeString, Computed: true},
						"name":                 {Type: schema.TypeString, Computed: true},
						"description":          {Type: schema.TypeString, Computed: true},
						"size":                 {Type: schema.TypeInt, Computed: true},
						"volume_type":          {Type: schema.TypeString, Computed: true},
						"zone":                 {Type: schema.TypeString, Computed: true},
						"status":               {Type: schema.TypeString, Computed: true},
						"iops":                 {Type: schema.TypeInt, Computed: true},
						"is_encrypted":         {Type: schema.TypeBool, Computed: true},
						"is_multiattach":       {Type: schema.TypeBool, Computed: true},
						"is_bootable":          {Type: schema.TypeBool, Computed: true},
						"attached_server_id":   {Type: schema.TypeString, Computed: true},
						"attached_server_name": {Type: schema.TypeString, Computed: true},
						"created_at":           {Type: schema.TypeString, Computed: true},
					},
				},
			},
		},
	}
}

func dataSourceVolumesRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	cfg := meta.(*config.Config)

	listResp := &dto.ListVolumesResponse{}
	_, err := cfg.Client.Get(ctx, client.ApiPath.Volumes(cfg.ProjectID), listResp, nil)
	if err != nil {
		return diag.Errorf("Error listing vnpaycloud_volumes: %s", err)
	}

	var volumes []map[string]interface{}
	for _, vol := range listResp.Volumes {
		volumes = append(volumes, map[string]interface{}{
			"id":                   vol.ID,
			"name":                 vol.Name,
			"description":          vol.Description,
			"size":                 vol.SizeGB,
			"volume_type":          vol.VolumeType,
			"zone":                 vol.Zone,
			"status":               vol.Status,
			"iops":                 vol.IOPS,
			"is_encrypted":         vol.IsEncrypted,
			"is_multiattach":       vol.IsMultiattach,
			"is_bootable":          vol.IsBootable,
			"attached_server_id":   vol.AttachedServerID,
			"attached_server_name": vol.AttachedServerName,
			"created_at":           vol.CreatedAt,
		})
	}

	d.SetId(fmt.Sprintf("volumes-%s", cfg.ProjectID))
	d.Set("volumes", volumes)

	return nil
}
