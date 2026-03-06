package subnet

import (
	"context"
	"fmt"
	"terraform-provider-vnpaycloud/vnpaycloud/config"
	"terraform-provider-vnpaycloud/vnpaycloud/dto"
	"terraform-provider-vnpaycloud/vnpaycloud/helper/client"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

func DataSourceSubnet() *schema.Resource {
	return &schema.Resource{
		ReadContext: dataSourceSubnetRead,
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
			"vpc_id": {
				Type:     schema.TypeString,
				Optional: true,
				Computed: true,
			},
			"cidr": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"gateway_ip": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"enable_dhcp": {
				Type:     schema.TypeBool,
				Computed: true,
			},
			"enable_snat": {
				Type:     schema.TypeBool,
				Computed: true,
			},
			"floating_ip_id": {
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
	}
}

func dataSourceSubnetRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	cfg := meta.(*config.Config)

	if id, ok := d.GetOk("id"); ok {
		subnetResp := &dto.SubnetResponse{}
		_, err := cfg.Client.Get(ctx, client.ApiPath.SubnetWithID(cfg.ProjectID, id.(string)), subnetResp, nil)
		if err != nil {
			return diag.Errorf("Error fetching vnpaycloud_subnet %s: %s", id, err)
		}
		return setSubnetData(d, &subnetResp.Subnet)
	}

	// List and filter client-side
	listResp := &dto.ListSubnetsResponse{}
	_, err := cfg.Client.Get(ctx, client.ApiPath.Subnets(cfg.ProjectID), listResp, nil)
	if err != nil {
		return diag.Errorf("Error listing vnpaycloud_subnet: %s", err)
	}

	nameFilter, nameOk := d.GetOk("name")
	vpcFilter, vpcOk := d.GetOk("vpc_id")

	for _, s := range listResp.Subnets {
		if nameOk && s.Name != nameFilter.(string) {
			continue
		}
		if vpcOk && s.VpcID != vpcFilter.(string) {
			continue
		}
		return setSubnetData(d, &s)
	}

	return diag.Errorf("No vnpaycloud_subnet found matching the criteria")
}

func setSubnetData(d *schema.ResourceData, s *dto.Subnet) diag.Diagnostics {
	d.SetId(s.ID)
	d.Set("name", s.Name)
	d.Set("vpc_id", s.VpcID)
	d.Set("cidr", s.CIDR)
	d.Set("gateway_ip", s.GatewayIP)
	d.Set("enable_dhcp", s.EnableDHCP)
	d.Set("enable_snat", s.EnableSnat)
	d.Set("floating_ip_id", s.ExternalIpID)
	d.Set("status", s.Status)
	d.Set("created_at", s.CreatedAt)
	return nil
}

func DataSourceSubnets() *schema.Resource {
	return &schema.Resource{
		ReadContext: dataSourceSubnetsRead,
		Schema: map[string]*schema.Schema{
			"vpc_id": {
				Type:     schema.TypeString,
				Optional: true,
			},
			"subnets": {
				Type:     schema.TypeList,
				Computed: true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"id":          {Type: schema.TypeString, Computed: true},
						"name":        {Type: schema.TypeString, Computed: true},
						"vpc_id":      {Type: schema.TypeString, Computed: true},
						"cidr":        {Type: schema.TypeString, Computed: true},
						"gateway_ip":  {Type: schema.TypeString, Computed: true},
						"enable_dhcp":   {Type: schema.TypeBool, Computed: true},
						"enable_snat":   {Type: schema.TypeBool, Computed: true},
						"floating_ip_id": {Type: schema.TypeString, Computed: true},
						"status":        {Type: schema.TypeString, Computed: true},
						"created_at":    {Type: schema.TypeString, Computed: true},
					},
				},
			},
		},
	}
}

func dataSourceSubnetsRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	cfg := meta.(*config.Config)

	listResp := &dto.ListSubnetsResponse{}
	_, err := cfg.Client.Get(ctx, client.ApiPath.Subnets(cfg.ProjectID), listResp, nil)
	if err != nil {
		return diag.Errorf("Error listing vnpaycloud_subnets: %s", err)
	}

	vpcFilter, vpcOk := d.GetOk("vpc_id")
	var subnets []map[string]interface{}
	for _, s := range listResp.Subnets {
		if vpcOk && s.VpcID != vpcFilter.(string) {
			continue
		}
		subnets = append(subnets, map[string]interface{}{
			"id":             s.ID,
			"name":           s.Name,
			"vpc_id":         s.VpcID,
			"cidr":           s.CIDR,
			"gateway_ip":     s.GatewayIP,
			"enable_dhcp":    s.EnableDHCP,
			"enable_snat":    s.EnableSnat,
			"floating_ip_id": s.ExternalIpID,
			"status":         s.Status,
			"created_at":     s.CreatedAt,
		})
	}

	d.SetId(fmt.Sprintf("subnets-%s", cfg.ProjectID))
	d.Set("subnets", subnets)

	return nil
}
