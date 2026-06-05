package lbflavor

import (
	"context"
	"terraform-provider-vnpaycloud/vnpaycloud/config"
	"terraform-provider-vnpaycloud/vnpaycloud/dto"
	"terraform-provider-vnpaycloud/vnpaycloud/helper/client"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

func DataSourceLBFlavors() *schema.Resource {
	return &schema.Resource{
		ReadContext: dataSourceLBFlavorsRead,
		Schema: map[string]*schema.Schema{
			"flavors": {
				Type:     schema.TypeList,
				Computed: true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"id":          {Type: schema.TypeString, Computed: true},
						"name":        {Type: schema.TypeString, Computed: true},
						"description": {Type: schema.TypeString, Computed: true},
						"zone_id":     {Type: schema.TypeString, Computed: true},
					},
				},
			},
		},
	}
}

func dataSourceLBFlavorsRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	cfg := meta.(*config.Config)

	resp := &dto.ListLoadBalancerFlavorsResponse{}
	_, err := cfg.Client.Get(ctx, client.ApiPath.LBFlavors(cfg.ProjectID), resp, nil)
	if err != nil {
		return diag.Errorf("Error listing vnpaycloud_lb_flavors: %s", err)
	}

	flavors := make([]map[string]interface{}, 0, len(resp.Flavors))
	for _, f := range resp.Flavors {
		flavors = append(flavors, map[string]interface{}{
			"id":          f.ID,
			"name":        f.Name,
			"description": f.Description,
			"zone_id":     f.ZoneID,
		})
	}

	d.SetId("lb-flavors")
	d.Set("flavors", flavors)

	return nil
}
