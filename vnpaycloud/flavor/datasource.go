package flavor

import (
	"context"
	"fmt"
	"terraform-provider-vnpaycloud/vnpaycloud/config"
	"terraform-provider-vnpaycloud/vnpaycloud/dto"
	"terraform-provider-vnpaycloud/vnpaycloud/helper/client"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

func DataSourceFlavor() *schema.Resource {
	return &schema.Resource{
		ReadContext: dataSourceFlavorRead,
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
			"vcpus": {
				Type:     schema.TypeInt,
				Computed: true,
			},
			"ram_mb": {
				Type:     schema.TypeInt,
				Computed: true,
			},
			"disk_gb": {
				Type:     schema.TypeInt,
				Computed: true,
			},
			"is_public": {
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

func dataSourceFlavorRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	cfg := meta.(*config.Config)

	if id, ok := d.GetOk("id"); ok {
		resp := &dto.FlavorResponse{}
		_, err := cfg.Client.Get(ctx, client.ApiPath.FlavorWithID(id.(string)), resp, nil)
		if err != nil {
			return diag.Errorf("Error fetching vnpaycloud_flavor %s: %s", id, err)
		}
		return setFlavorData(d, &resp.Flavor)
	}

	listResp := &dto.ListFlavorsResponse{}
	_, err := cfg.Client.Get(ctx, client.ApiPath.Flavors(cfg.ZoneID), listResp, nil)
	if err != nil {
		return diag.Errorf("Error listing vnpaycloud_flavor: %s", err)
	}

	nameFilter := d.Get("name").(string)
	var matched []dto.Flavor

	for _, f := range listResp.Flavors {
		if f.Name != nameFilter {
			continue
		}
		matched = append(matched, f)
	}

	if len(matched) == 0 {
		return diag.Errorf("No vnpaycloud_flavor found matching the criteria")
	}
	if len(matched) > 1 {
		return diag.Errorf("Your vnpaycloud_flavor query returned multiple results")
	}

	return setFlavorData(d, &matched[0])
}

func setFlavorData(d *schema.ResourceData, f *dto.Flavor) diag.Diagnostics {
	d.SetId(f.ID)
	d.Set("name", f.Name)
	d.Set("vcpus", f.VCPUs)
	d.Set("ram_mb", f.RAMMB)
	d.Set("disk_gb", f.DiskGB)
	d.Set("is_public", f.IsPublic)
	d.Set("zone", f.Zone)
	return nil
}

func DataSourceFlavors() *schema.Resource {
	return &schema.Resource{
		ReadContext: dataSourceFlavorsRead,
		Schema: map[string]*schema.Schema{
			"flavors": {
				Type:     schema.TypeList,
				Computed: true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"id":        {Type: schema.TypeString, Computed: true},
						"name":      {Type: schema.TypeString, Computed: true},
						"vcpus":     {Type: schema.TypeInt, Computed: true},
						"ram_mb":    {Type: schema.TypeInt, Computed: true},
						"disk_gb":   {Type: schema.TypeInt, Computed: true},
						"is_public": {Type: schema.TypeBool, Computed: true},
						"zone":      {Type: schema.TypeString, Computed: true},
					},
				},
			},
		},
	}
}

func dataSourceFlavorsRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	cfg := meta.(*config.Config)

	listResp := &dto.ListFlavorsResponse{}
	_, err := cfg.Client.Get(ctx, client.ApiPath.Flavors(cfg.ZoneID), listResp, nil)
	if err != nil {
		return diag.Errorf("Error listing vnpaycloud_flavors: %s", err)
	}

	var flavors []map[string]interface{}
	for _, f := range listResp.Flavors {
		flavors = append(flavors, map[string]interface{}{
			"id":        f.ID,
			"name":      f.Name,
			"vcpus":     f.VCPUs,
			"ram_mb":    f.RAMMB,
			"disk_gb":   f.DiskGB,
			"is_public": f.IsPublic,
			"zone":      f.Zone,
		})
	}

	d.SetId(fmt.Sprintf("flavors-%s", cfg.ZoneID))
	d.Set("flavors", flavors)

	return nil
}
