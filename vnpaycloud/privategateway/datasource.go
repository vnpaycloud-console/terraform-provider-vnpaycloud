package privategateway

import (
	"context"
	"fmt"
	"terraform-provider-vnpaycloud/vnpaycloud/config"
	"terraform-provider-vnpaycloud/vnpaycloud/dto"
	"terraform-provider-vnpaycloud/vnpaycloud/helper/client"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

func DataSourcePrivateGateway() *schema.Resource {
	return &schema.Resource{
		ReadContext: dataSourcePrivateGatewayRead,
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
			"load_balancer_id": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"subnet_id": {
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
			"zone_id": {
				Type:     schema.TypeString,
				Computed: true,
			},
		},
	}
}

func dataSourcePrivateGatewayRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	cfg := meta.(*config.Config)

	if id, ok := d.GetOk("id"); ok {
		pgwResp := &dto.PrivateGatewayResponse{}
		_, err := cfg.Client.Get(ctx, client.ApiPath.PrivateGatewayWithID(cfg.ProjectID, id.(string)), pgwResp, nil)
		if err != nil {
			return diag.Errorf("Error fetching vnpaycloud_private_gateway %s: %s", id, err)
		}
		return setPrivateGatewayData(d, &pgwResp.PrivateGateway)
	}

	listResp := &dto.ListPrivateGatewaysResponse{}
	_, err := cfg.Client.Get(ctx, client.ApiPath.PrivateGateways(cfg.ProjectID), listResp, nil)
	if err != nil {
		return diag.Errorf("Error listing vnpaycloud_private_gateway: %s", err)
	}

	nameFilter, nameOk := d.GetOk("name")
	if !nameOk {
		return diag.Errorf("Either 'id' or 'name' must be specified to look up a vnpaycloud_private_gateway")
	}

	for _, pgw := range listResp.PrivateGateways {
		if pgw.Name != nameFilter.(string) {
			continue
		}
		return setPrivateGatewayData(d, &pgw)
	}

	return diag.Errorf("No vnpaycloud_private_gateway found matching the criteria")
}

func setPrivateGatewayData(d *schema.ResourceData, pgw *dto.PrivateGateway) diag.Diagnostics {
	d.SetId(pgw.ID)
	d.Set("name", pgw.Name)
	d.Set("description", pgw.Description)
	d.Set("load_balancer_id", pgw.LoadBalancerID)
	d.Set("subnet_id", pgw.SubnetID)
	d.Set("status", pgw.Status)
	d.Set("created_at", pgw.CreatedAt)
	d.Set("zone_id", pgw.ZoneID)
	return nil
}

func DataSourcePrivateGateways() *schema.Resource {
	return &schema.Resource{
		ReadContext: dataSourcePrivateGatewaysRead,
		Schema: map[string]*schema.Schema{
			"private_gateways": {
				Type:     schema.TypeList,
				Computed: true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"id":               {Type: schema.TypeString, Computed: true},
						"name":             {Type: schema.TypeString, Computed: true},
						"description":      {Type: schema.TypeString, Computed: true},
						"load_balancer_id": {Type: schema.TypeString, Computed: true},
						"subnet_id":        {Type: schema.TypeString, Computed: true},
						"status":           {Type: schema.TypeString, Computed: true},
						"created_at":       {Type: schema.TypeString, Computed: true},
						"zone_id":          {Type: schema.TypeString, Computed: true},
					},
				},
			},
		},
	}
}

func dataSourcePrivateGatewaysRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	cfg := meta.(*config.Config)

	listResp := &dto.ListPrivateGatewaysResponse{}
	_, err := cfg.Client.Get(ctx, client.ApiPath.PrivateGateways(cfg.ProjectID), listResp, nil)
	if err != nil {
		return diag.Errorf("Error listing vnpaycloud_private_gateways: %s", err)
	}

	var privateGateways []map[string]interface{}
	for _, pgw := range listResp.PrivateGateways {
		privateGateways = append(privateGateways, map[string]interface{}{
			"id":               pgw.ID,
			"name":             pgw.Name,
			"description":      pgw.Description,
			"load_balancer_id": pgw.LoadBalancerID,
			"subnet_id":        pgw.SubnetID,
			"status":           pgw.Status,
			"created_at":       pgw.CreatedAt,
			"zone_id":          pgw.ZoneID,
		})
	}

	d.SetId(fmt.Sprintf("private-gateways-%s", cfg.ProjectID))
	d.Set("private_gateways", privateGateways)

	return nil
}
