package customergateway

import (
	"context"
	"fmt"
	"terraform-provider-vnpaycloud/vnpaycloud/config"
	"terraform-provider-vnpaycloud/vnpaycloud/dto"
	"terraform-provider-vnpaycloud/vnpaycloud/helper/client"
	"terraform-provider-vnpaycloud/vnpaycloud/util"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

func DataSourceCustomerGateway() *schema.Resource {
	return &schema.Resource{
		ReadContext: dataSourceCustomerGatewayRead,
		Description: "Use this data source to retrieve a VNPAY Cloud customer gateway by ID or name.",
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
			"public_ip": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "The public IPv4 address of the customer-side VPN device.",
			},
			"vpn_type": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"status": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"remote_prefixes": {
				Type:        schema.TypeSet,
				Computed:    true,
				Description: "The remote network CIDR prefixes behind the customer gateway.",
				Elem:        &schema.Schema{Type: schema.TypeString},
			},
			"remote_tunnel_ip": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "The tunnel IP address on the customer gateway side, used for route-based VPN.",
			},
			"local_tunnel_ip": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "The tunnel IP address on the VNPAY Cloud side, used for route-based VPN.",
			},
			"routing_mode": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "The routing mode for route-based VPN. Valid values are NONE, STATIC, and DYNAMIC.",
			},
			"bgp_config": {
				Type:        schema.TypeList,
				Computed:    true,
				Description: "The BGP configuration for route-based VPN with dynamic routing.",
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"local_as": {
							Type:        schema.TypeInt,
							Computed:    true,
							Description: "The local BGP autonomous system number.",
						},
						"peer_as": {
							Type:        schema.TypeInt,
							Computed:    true,
							Description: "The peer BGP autonomous system number.",
						},
						"as_path": {
							Type:        schema.TypeString,
							Computed:    true,
							Description: "The configured BGP AS path.",
						},
					},
				},
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

func dataSourceCustomerGatewayRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	cfg := meta.(*config.Config)

	nameFilter, nameOk := d.GetOk("name")
	if id, ok := d.GetOk("id"); ok {
		cgResp := &dto.CustomerGatewayResponse{}
		_, err := cfg.Client.Get(ctx, client.ApiPath.CustomerGatewayWithID(cfg.ProjectID, id.(string)), cgResp, nil)
		if err != nil {
			return diag.Errorf("Error fetching vnpaycloud_customer_gateway %s: %s", id, err)
		}
		if nameOk && cgResp.CustomerGateway.Name != nameFilter.(string) {
			return diag.Errorf("vnpaycloud_customer_gateway %s does not match name %q", id, nameFilter.(string))
		}

		return setCustomerGatewayData(d, &cgResp.CustomerGateway)
	}

	if !nameOk {
		return diag.Errorf("One of id or name must be specified for vnpaycloud_customer_gateway")
	}

	listResp := &dto.ListCustomerGatewaysResponse{}
	_, err := cfg.Client.Get(ctx, client.ApiPath.CustomerGateways(cfg.ProjectID), listResp, nil)
	if err != nil {
		return diag.Errorf("Error listing vnpaycloud_customer_gateway: %s", err)
	}

	matches := make([]dto.CustomerGateway, 0, 1)
	for _, cg := range listResp.CustomerGateways {
		if cg.Name != nameFilter.(string) {
			continue
		}
		matches = append(matches, cg)
	}

	if len(matches) > 1 {
		return diag.Errorf("Multiple vnpaycloud_customer_gateway resources found with name %q", nameFilter.(string))
	}
	if len(matches) == 1 {
		return setCustomerGatewayData(d, &matches[0])
	}

	return diag.Errorf("No vnpaycloud_customer_gateway found matching the criteria")
}

func setCustomerGatewayData(d *schema.ResourceData, cg *dto.CustomerGateway) diag.Diagnostics {
	d.SetId(cg.ID)
	d.Set("name", cg.Name)
	d.Set("description", cg.Description)
	d.Set("public_ip", cg.PublicIP)
	d.Set("vpn_type", cg.VPNType)
	d.Set("status", util.NormalizeStatus(cg.Status))
	d.Set("remote_prefixes", cg.RemotePrefixes)
	d.Set("remote_tunnel_ip", cg.RemoteTunnelIP)
	d.Set("local_tunnel_ip", cg.LocalTunnelIP)
	d.Set("routing_mode", cg.RoutingMode)
	d.Set("created_at", cg.CreatedAt)
	d.Set("zone_id", cg.ZoneID)

	if cg.BGPConfig != nil {
		bgpConfig := []map[string]interface{}{
			{
				"local_as": int(cg.BGPConfig.LocalAs),
				"peer_as":  int(cg.BGPConfig.PeerAs),
				"as_path":  cg.BGPConfig.AsPath,
			},
		}
		d.Set("bgp_config", bgpConfig)
	} else {
		d.Set("bgp_config", nil)
	}

	return nil
}

func DataSourceCustomerGateways() *schema.Resource {
	return &schema.Resource{
		ReadContext: dataSourceCustomerGatewaysRead,
		Description: "Use this data source to retrieve all VNPAY Cloud customer gateways in the current project.",
		Schema: map[string]*schema.Schema{
			"customer_gateways": {
				Type:        schema.TypeList,
				Computed:    true,
				Description: "The list of customer gateways in the current project.",
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
						"public_ip": {
							Type:        schema.TypeString,
							Computed:    true,
							Description: "The public IPv4 address of the customer-side VPN device.",
						},
						"vpn_type": {
							Type:     schema.TypeString,
							Computed: true,
						},
						"remote_prefixes": {
							Type:        schema.TypeSet,
							Computed:    true,
							Description: "The remote network CIDR prefixes behind the customer gateway.",
							Elem:        &schema.Schema{Type: schema.TypeString},
						},
						"remote_tunnel_ip": {
							Type:        schema.TypeString,
							Computed:    true,
							Description: "The tunnel IP address on the customer gateway side, used for route-based VPN.",
						},
						"local_tunnel_ip": {
							Type:        schema.TypeString,
							Computed:    true,
							Description: "The tunnel IP address on the VNPAY Cloud side, used for route-based VPN.",
						},
						"routing_mode": {
							Type:        schema.TypeString,
							Computed:    true,
							Description: "The routing mode for route-based VPN.",
						},
						"bgp_config": {
							Type:        schema.TypeList,
							Computed:    true,
							Description: "The BGP configuration for route-based VPN with dynamic routing.",
							Elem: &schema.Resource{
								Schema: map[string]*schema.Schema{
									"local_as": {
										Type:        schema.TypeInt,
										Computed:    true,
										Description: "The local BGP autonomous system number.",
									},
									"peer_as": {
										Type:        schema.TypeInt,
										Computed:    true,
										Description: "The peer BGP autonomous system number.",
									},
									"as_path": {
										Type:        schema.TypeString,
										Computed:    true,
										Description: "The configured BGP AS path.",
									},
								},
							},
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
				},
			},
		},
	}
}

func dataSourceCustomerGatewaysRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	cfg := meta.(*config.Config)

	listResp := &dto.ListCustomerGatewaysResponse{}
	_, err := cfg.Client.Get(ctx, client.ApiPath.CustomerGateways(cfg.ProjectID), listResp, nil)
	if err != nil {
		return diag.Errorf("Error listing vnpaycloud_customer_gateways: %s", err)
	}

	var customerGateways []map[string]interface{}
	for _, cg := range listResp.CustomerGateways {
		customerGateways = append(customerGateways, map[string]interface{}{
			"id":               cg.ID,
			"name":             cg.Name,
			"description":      cg.Description,
			"public_ip":        cg.PublicIP,
			"vpn_type":         cg.VPNType,
			"status":           util.NormalizeStatus(cg.Status),
			"remote_prefixes":  cg.RemotePrefixes,
			"remote_tunnel_ip": cg.RemoteTunnelIP,
			"local_tunnel_ip":  cg.LocalTunnelIP,
			"routing_mode":     cg.RoutingMode,
			"bgp_config":       flattenCustomerGatewayBGPConfig(cg.BGPConfig),
			"created_at":       cg.CreatedAt,
			"zone_id":          cg.ZoneID,
		})
	}

	d.SetId(fmt.Sprintf("customer-gateways-%s", cfg.ProjectID))
	d.Set("customer_gateways", customerGateways)

	return nil
}

func flattenCustomerGatewayBGPConfig(bgpConfig *dto.BGPConfig) []map[string]interface{} {
	if bgpConfig == nil {
		return nil
	}

	return []map[string]interface{}{
		{
			"local_as": int(bgpConfig.LocalAs),
			"peer_as":  int(bgpConfig.PeerAs),
			"as_path":  bgpConfig.AsPath,
		},
	}
}
