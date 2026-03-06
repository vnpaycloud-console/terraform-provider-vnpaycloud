package loadbalancer

import (
	"context"
	"terraform-provider-vnpaycloud/vnpaycloud/config"
	"terraform-provider-vnpaycloud/vnpaycloud/dto"
	"terraform-provider-vnpaycloud/vnpaycloud/helper/client"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

func DataSourceLoadBalancer() *schema.Resource {
	return &schema.Resource{
		ReadContext: dataSourceLoadBalancerRead,
		Schema: map[string]*schema.Schema{
			"id": {
				Type:     schema.TypeString,
				Required: true,
			},
			"name": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"description": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"vip_address": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"vip_subnet_id": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"status": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"listener_ids": {
				Type:     schema.TypeList,
				Computed: true,
				Elem:     &schema.Schema{Type: schema.TypeString},
			},
			"created_at": {
				Type:     schema.TypeString,
				Computed: true,
			},
		},
	}
}

func dataSourceLoadBalancerRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	cfg := meta.(*config.Config)

	id := d.Get("id").(string)

	lbResp := &dto.LoadBalancerResponse{}
	_, err := cfg.Client.Get(ctx, client.ApiPath.LoadBalancerWithID(cfg.ProjectID, id), lbResp, nil)
	if err != nil {
		return diag.Errorf("Error retrieving vnpaycloud_lb_loadbalancer %s: %s", id, err)
	}

	d.SetId(lbResp.LoadBalancer.ID)
	d.Set("name", lbResp.LoadBalancer.Name)
	d.Set("description", lbResp.LoadBalancer.Description)
	d.Set("vip_address", lbResp.LoadBalancer.VipAddress)
	d.Set("vip_subnet_id", lbResp.LoadBalancer.VipSubnetID)
	d.Set("status", lbResp.LoadBalancer.Status)
	d.Set("listener_ids", lbResp.LoadBalancer.ListenerIDs)
	d.Set("created_at", lbResp.LoadBalancer.CreatedAt)

	return nil
}

func DataSourceLoadBalancers() *schema.Resource {
	return &schema.Resource{
		ReadContext: dataSourceLoadBalancersRead,
		Schema: map[string]*schema.Schema{
			"load_balancers": {
				Type:     schema.TypeList,
				Computed: true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"id": {
							Type:     schema.TypeString,
							Computed: true,
						},
						"name": {
							Type:     schema.TypeString,
							Computed: true,
						},
						"description": {
							Type:     schema.TypeString,
							Computed: true,
						},
						"vip_address": {
							Type:     schema.TypeString,
							Computed: true,
						},
						"vip_subnet_id": {
							Type:     schema.TypeString,
							Computed: true,
						},
						"status": {
							Type:     schema.TypeString,
							Computed: true,
						},
						"created_at": {
							Type:     schema.TypeString,
							Computed: true,
						},
					},
				},
			},
		},
	}
}

func dataSourceLoadBalancersRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	cfg := meta.(*config.Config)

	lbsResp := &dto.ListLoadBalancersResponse{}
	_, err := cfg.Client.Get(ctx, client.ApiPath.LoadBalancers(cfg.ProjectID), lbsResp, nil)
	if err != nil {
		return diag.Errorf("Error listing vnpaycloud_lb_loadbalancers: %s", err)
	}

	var lbs []map[string]interface{}
	for _, lb := range lbsResp.LoadBalancers {
		lbs = append(lbs, map[string]interface{}{
			"id":            lb.ID,
			"name":          lb.Name,
			"description":   lb.Description,
			"vip_address":   lb.VipAddress,
			"vip_subnet_id": lb.VipSubnetID,
			"status":        lb.Status,
			"created_at":    lb.CreatedAt,
		})
	}

	d.SetId("load_balancers")
	d.Set("load_balancers", lbs)

	return nil
}
