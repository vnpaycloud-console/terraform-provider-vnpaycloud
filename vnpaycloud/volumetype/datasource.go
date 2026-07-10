package volumetype

import (
	"context"
	"fmt"
	"terraform-provider-vnpaycloud/vnpaycloud/config"
	"terraform-provider-vnpaycloud/vnpaycloud/dto"
	"terraform-provider-vnpaycloud/vnpaycloud/helper/client"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

func DataSourceVolumeType() *schema.Resource {
	return &schema.Resource{
		ReadContext: dataSourceVolumeTypeRead,
		Schema: map[string]*schema.Schema{
			"id": {
				Type:         schema.TypeString,
				Optional:     true,
				Computed:     true,
				ExactlyOneOf: []string{"id", "name"},
			},
			"name": {
				Type:         schema.TypeString,
				Optional:     true,
				Computed:     true,
				ExactlyOneOf: []string{"id", "name"},
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
			"zone": {
				Type:     schema.TypeString,
				Computed: true,
			},
		},
	}
}

func dataSourceVolumeTypeRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	cfg := meta.(*config.Config)

	if id, ok := d.GetOk("id"); ok {
		resp := &dto.VolumeTypeResponse{}
		_, err := cfg.Client.Get(ctx, client.ApiPath.VolumeTypeWithID(id.(string)), resp, nil)
		if err != nil {
			return diag.Errorf("Error fetching vnpaycloud_volume_type %s: %s", id, err)
		}
		return setVolumeTypeData(d, &resp.VolumeType)
	}

	listResp := &dto.ListVolumeTypesResponse{}
	_, err := cfg.Client.Get(ctx, client.ApiPath.VolumeTypes(cfg.ZoneID), listResp, nil)
	if err != nil {
		return diag.Errorf("Error listing vnpaycloud_volume_type: %s", err)
	}

	nameFilter := d.Get("name").(string)
	var matched []dto.VolumeType

	for _, vt := range listResp.VolumeTypes {
		if vt.Name != nameFilter {
			continue
		}
		matched = append(matched, vt)
	}

	if len(matched) == 0 {
		return diag.Errorf("No vnpaycloud_volume_type found matching the criteria")
	}
	if len(matched) > 1 {
		return diag.Errorf("Your vnpaycloud_volume_type query returned multiple results")
	}

	return setVolumeTypeData(d, &matched[0])
}

func setVolumeTypeData(d *schema.ResourceData, vt *dto.VolumeType) diag.Diagnostics {
	d.SetId(vt.ID)
	d.Set("name", vt.Name)
	d.Set("iops", vt.IOPS)
	d.Set("is_encrypted", vt.IsEncrypted)
	d.Set("is_multiattach", vt.IsMultiattach)
	d.Set("zone", vt.Zone)
	return nil
}

func DataSourceVolumeTypes() *schema.Resource {
	return &schema.Resource{
		ReadContext: dataSourceVolumeTypesRead,
		Schema: map[string]*schema.Schema{
			"volume_types": {
				Type:     schema.TypeList,
				Computed: true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"id":             {Type: schema.TypeString, Computed: true},
						"name":           {Type: schema.TypeString, Computed: true},
						"iops":           {Type: schema.TypeInt, Computed: true},
						"is_encrypted":   {Type: schema.TypeBool, Computed: true},
						"is_multiattach": {Type: schema.TypeBool, Computed: true},
						"zone":           {Type: schema.TypeString, Computed: true},
					},
				},
			},
		},
	}
}

func dataSourceVolumeTypesRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	cfg := meta.(*config.Config)

	listResp := &dto.ListVolumeTypesResponse{}
	_, err := cfg.Client.Get(ctx, client.ApiPath.VolumeTypes(cfg.ZoneID), listResp, nil)
	if err != nil {
		return diag.Errorf("Error listing vnpaycloud_volume_types: %s", err)
	}

	var volumeTypes []map[string]interface{}
	for _, vt := range listResp.VolumeTypes {
		volumeTypes = append(volumeTypes, map[string]interface{}{
			"id":             vt.ID,
			"name":           vt.Name,
			"iops":           vt.IOPS,
			"is_encrypted":   vt.IsEncrypted,
			"is_multiattach": vt.IsMultiattach,
			"zone":           vt.Zone,
		})
	}

	d.SetId(fmt.Sprintf("volume-types-%s", cfg.ZoneID))
	d.Set("volume_types", volumeTypes)

	return nil
}
