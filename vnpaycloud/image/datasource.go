package image

import (
	"context"
	"fmt"
	"terraform-provider-vnpaycloud/vnpaycloud/config"
	"terraform-provider-vnpaycloud/vnpaycloud/dto"
	"terraform-provider-vnpaycloud/vnpaycloud/helper/client"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

func DataSourceImage() *schema.Resource {
	return &schema.Resource{
		ReadContext: dataSourceImageRead,
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
			"os_type": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"os_version": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"min_disk_gb": {
				Type:     schema.TypeInt,
				Computed: true,
			},
			"status": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"zone": {
				Type:     schema.TypeString,
				Computed: true,
			},
		},
	}
}

func dataSourceImageRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	cfg := meta.(*config.Config)

	if id, ok := d.GetOk("id"); ok {
		resp := &dto.ImageResponse{}
		_, err := cfg.Client.Get(ctx, client.ApiPath.ImageWithID(id.(string)), resp, nil)
		if err != nil {
			return diag.Errorf("Error fetching vnpaycloud_image %s: %s", id, err)
		}
		return setImageData(d, &resp.Image)
	}

	listResp := &dto.ListImagesResponse{}
	_, err := cfg.Client.Get(ctx, client.ApiPath.Images(cfg.ZoneID), listResp, nil)
	if err != nil {
		return diag.Errorf("Error listing vnpaycloud_image: %s", err)
	}

	nameFilter, nameOk := d.GetOk("name")

	for _, img := range listResp.Images {
		if nameOk && img.Name != nameFilter.(string) {
			continue
		}
		return setImageData(d, &img)
	}

	return diag.Errorf("No vnpaycloud_image found matching the criteria")
}

func setImageData(d *schema.ResourceData, img *dto.Image) diag.Diagnostics {
	d.SetId(img.ID)
	d.Set("name", img.Name)
	d.Set("os_type", img.OsType)
	d.Set("os_version", img.OsVersion)
	d.Set("min_disk_gb", img.MinDiskGB)
	d.Set("status", img.Status)
	d.Set("zone", img.Zone)
	return nil
}

func DataSourceImages() *schema.Resource {
	return &schema.Resource{
		ReadContext: dataSourceImagesRead,
		Schema: map[string]*schema.Schema{
			"images": {
				Type:     schema.TypeList,
				Computed: true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"id":          {Type: schema.TypeString, Computed: true},
						"name":        {Type: schema.TypeString, Computed: true},
						"os_type":     {Type: schema.TypeString, Computed: true},
						"os_version":  {Type: schema.TypeString, Computed: true},
						"min_disk_gb": {Type: schema.TypeInt, Computed: true},
						"status":      {Type: schema.TypeString, Computed: true},
						"zone":        {Type: schema.TypeString, Computed: true},
					},
				},
			},
		},
	}
}

func dataSourceImagesRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	cfg := meta.(*config.Config)

	listResp := &dto.ListImagesResponse{}
	_, err := cfg.Client.Get(ctx, client.ApiPath.Images(cfg.ZoneID), listResp, nil)
	if err != nil {
		return diag.Errorf("Error listing vnpaycloud_images: %s", err)
	}

	var images []map[string]interface{}
	for _, img := range listResp.Images {
		images = append(images, map[string]interface{}{
			"id":          img.ID,
			"name":        img.Name,
			"os_type":     img.OsType,
			"os_version":  img.OsVersion,
			"min_disk_gb": img.MinDiskGB,
			"status":      img.Status,
			"zone":        img.Zone,
		})
	}

	d.SetId(fmt.Sprintf("images-%s", cfg.ZoneID))
	d.Set("images", images)

	return nil
}
