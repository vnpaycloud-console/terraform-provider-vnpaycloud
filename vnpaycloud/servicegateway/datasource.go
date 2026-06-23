package servicegateway

import (
	"context"
	"terraform-provider-vnpaycloud/vnpaycloud/config"
	"terraform-provider-vnpaycloud/vnpaycloud/dto"
	"terraform-provider-vnpaycloud/vnpaycloud/helper/client"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

func serviceGatewayElemSchema() map[string]*schema.Schema {
	return map[string]*schema.Schema{
		"id":                  {Type: schema.TypeString, Computed: true},
		"name":                {Type: schema.TypeString, Computed: true},
		"description":         {Type: schema.TypeString, Computed: true},
		"subnet_id":           {Type: schema.TypeString, Computed: true},
		"vpc_id":              {Type: schema.TypeString, Computed: true},
		"flavor_id":           {Type: schema.TypeString, Computed: true},
		"allowed_icmp":        {Type: schema.TypeBool, Computed: true},
		"vip_address":         {Type: schema.TypeString, Computed: true},
		"load_balancer_id":    {Type: schema.TypeString, Computed: true},
		"port_id":             {Type: schema.TypeString, Computed: true},
		"operating_status":    {Type: schema.TypeString, Computed: true},
		"provisioning_status": {Type: schema.TypeString, Computed: true},
		"status":              {Type: schema.TypeString, Computed: true},
		"created_at":          {Type: schema.TypeString, Computed: true},
	}
}

func flattenServiceGateway(sg dto.ServiceGateway) map[string]interface{} {
	return map[string]interface{}{
		"id":                  sg.ID,
		"name":                sg.Name,
		"description":         sg.Description,
		"subnet_id":           sg.SubnetID,
		"vpc_id":              sg.VPCID,
		"flavor_id":           sg.FlavorID,
		"allowed_icmp":        sg.AllowedICMP,
		"vip_address":         sg.VipAddress,
		"load_balancer_id":    sg.LoadBalancerID,
		"port_id":             sg.PortID,
		"operating_status":    sg.OperatingStatus,
		"provisioning_status": sg.ProvisioningStatus,
		"status":              sg.Status,
		"created_at":          sg.CreatedAt,
	}
}

func DataSourceServiceGateway() *schema.Resource {
	elem := serviceGatewayElemSchema()
	elem["id"] = &schema.Schema{Type: schema.TypeString, Required: true}
	return &schema.Resource{
		ReadContext: dataSourceServiceGatewayRead,
		Schema:      elem,
	}
}

func dataSourceServiceGatewayRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	cfg := meta.(*config.Config)

	id := d.Get("id").(string)

	resp := &dto.ServiceGatewayResponse{}
	if _, err := cfg.Client.Get(ctx, client.ApiPath.ServiceGatewayWithID(cfg.ProjectID, id), resp, nil); err != nil {
		return diag.Errorf("Error retrieving vnpaycloud_service_gateway %s: %s", id, err)
	}

	d.SetId(resp.ServiceGateway.ID)
	for k, v := range flattenServiceGateway(resp.ServiceGateway) {
		if k == "id" {
			continue
		}
		d.Set(k, v)
	}

	return nil
}

func DataSourceServiceGateways() *schema.Resource {
	return &schema.Resource{
		ReadContext: dataSourceServiceGatewaysRead,
		Schema: map[string]*schema.Schema{
			"service_gateways": {
				Type:     schema.TypeList,
				Computed: true,
				Elem:     &schema.Resource{Schema: serviceGatewayElemSchema()},
			},
		},
	}
}

func dataSourceServiceGatewaysRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	cfg := meta.(*config.Config)

	resp := &dto.ListServiceGatewaysResponse{}
	if _, err := cfg.Client.Get(ctx, client.ApiPath.ServiceGateways(cfg.ProjectID), resp, nil); err != nil {
		return diag.Errorf("Error listing vnpaycloud_service_gateways: %s", err)
	}

	gateways := make([]map[string]interface{}, 0, len(resp.ServiceGateways))
	for _, sg := range resp.ServiceGateways {
		gateways = append(gateways, flattenServiceGateway(sg))
	}

	d.SetId("service_gateways")
	d.Set("service_gateways", gateways)

	return nil
}

func DataSourceServiceGatewayFlavors() *schema.Resource {
	return &schema.Resource{
		ReadContext: dataSourceServiceGatewayFlavorsRead,
		Schema: map[string]*schema.Schema{
			"flavors": {
				Type:     schema.TypeList,
				Computed: true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"id":          {Type: schema.TypeString, Computed: true},
						"name":        {Type: schema.TypeString, Computed: true},
						"description": {Type: schema.TypeString, Computed: true},
					},
				},
			},
		},
	}
}

func dataSourceServiceGatewayFlavorsRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	cfg := meta.(*config.Config)

	resp := &dto.ListServiceGatewayFlavorsResponse{}
	if _, err := cfg.Client.Get(ctx, client.ApiPath.ServiceGatewayFlavors(cfg.ProjectID), resp, nil); err != nil {
		return diag.Errorf("Error listing vnpaycloud_service_gateway_flavors: %s", err)
	}

	flavors := make([]map[string]interface{}, 0, len(resp.Flavors))
	for _, f := range resp.Flavors {
		flavors = append(flavors, map[string]interface{}{
			"id":          f.ID,
			"name":        f.Name,
			"description": f.Description,
		})
	}

	d.SetId("service_gateway_flavors")
	d.Set("flavors", flavors)

	return nil
}
