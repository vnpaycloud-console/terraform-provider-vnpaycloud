package vpngateway

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

func DataSourceVPNGateway() *schema.Resource {
	return &schema.Resource{
		ReadContext: dataSourceVPNGatewayRead,
		Description: "Use this data source to retrieve a VNPAY Cloud VPN gateway by ID or name.",
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
			"vpn_type": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"status": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"attached_vpc_ids": {
				Type:        schema.TypeList,
				Computed:    true,
				Description: "The IDs of VPCs currently attached to the VPN gateway.",
				Elem:        &schema.Schema{Type: schema.TypeString},
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

func dataSourceVPNGatewayRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	cfg := meta.(*config.Config)

	nameFilter, nameOk := d.GetOk("name")
	if id, ok := d.GetOk("id"); ok {
		vpnGatewayResp := &dto.VPNGatewayResponse{}
		_, err := cfg.Client.Get(ctx, client.ApiPath.VPNGatewayWithID(cfg.ProjectID, id.(string)), vpnGatewayResp, nil)
		if err != nil {
			return diag.Errorf("Error fetching vnpaycloud_vpnaas_vpn_gateway %s: %s", id, err)
		}
		if nameOk && vpnGatewayResp.VPNGateway.Name != nameFilter.(string) {
			return diag.Errorf("vnpaycloud_vpn_gateway %s does not match name %q", id, nameFilter.(string))
		}

		return setVPNGatewayData(d, &vpnGatewayResp.VPNGateway)
	}

	if !nameOk {
		return diag.Errorf("One of id or name must be specified for vnpaycloud_vpn_gateway")
	}

	listResp := &dto.ListVPNGatewaysResponse{}
	_, err := cfg.Client.Get(ctx, client.ApiPath.VPNGateways(cfg.ProjectID), listResp, nil)
	if err != nil {
		return diag.Errorf("Error listing vnpaycloud_vpnaas_vpn_gateway: %s", err)
	}

	matches := make([]dto.VPNGateway, 0, 1)
	for _, vpnGateway := range listResp.VPNGateways {
		if vpnGateway.Name != nameFilter.(string) {
			continue
		}
		matches = append(matches, vpnGateway)
	}

	if len(matches) > 1 {
		return diag.Errorf("Multiple vnpaycloud_vpn_gateway resources found with name %q", nameFilter.(string))
	}
	if len(matches) == 1 {
		return setVPNGatewayData(d, &matches[0])
	}

	return diag.Errorf("No vnpaycloud_vpn_gateway found matching the criteria")
}

func setVPNGatewayData(d *schema.ResourceData, gw *dto.VPNGateway) diag.Diagnostics {
	d.SetId(gw.ID)
	d.Set("name", gw.Name)
	d.Set("description", gw.Description)
	d.Set("vpn_type", gw.VPNType)
	d.Set("status", util.NormalizeStatus(gw.Status))
	d.Set("attached_vpc_ids", gw.AttachedVPCIDs)
	d.Set("created_at", gw.CreatedAt)
	d.Set("zone_id", gw.ZoneID)
	return nil
}

func DataSourceVPNGateways() *schema.Resource {
	return &schema.Resource{
		ReadContext: dataSourceVPNGatewaysRead,
		Description: "Use this data source to retrieve all VNPAY Cloud VPN gateways in the current project.",
		Schema: map[string]*schema.Schema{
			"vpn_gateways": {
				Type:        schema.TypeList,
				Computed:    true,
				Description: "The list of VPN gateways in the current project.",
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
						"vpn_type": {
							Type:     schema.TypeString,
							Computed: true,
						},
						"status": {
							Type:     schema.TypeString,
							Computed: true,
						},
						"attached_vpc_ids": {
							Type:        schema.TypeList,
							Computed:    true,
							Description: "The IDs of VPCs currently attached to the VPN gateway.",
							Elem:        &schema.Schema{Type: schema.TypeString},
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

func dataSourceVPNGatewaysRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	cfg := meta.(*config.Config)

	listResp := &dto.ListVPNGatewaysResponse{}
	_, err := cfg.Client.Get(ctx, client.ApiPath.VPNGateways(cfg.ProjectID), listResp, nil)
	if err != nil {
		return diag.Errorf("Error listing vnpaycloud_vpn_gateways: %s", err)
	}

	var vpnGateways []map[string]interface{}
	for _, vpnGateway := range listResp.VPNGateways {
		vpnGateways = append(vpnGateways, map[string]interface{}{
			"id":               vpnGateway.ID,
			"name":             vpnGateway.Name,
			"description":      vpnGateway.Description,
			"vpn_type":         vpnGateway.VPNType,
			"status":           util.NormalizeStatus(vpnGateway.Status),
			"attached_vpc_ids": vpnGateway.AttachedVPCIDs,
			"created_at":       vpnGateway.CreatedAt,
			"zone_id":          vpnGateway.ZoneID,
		})
	}

	d.SetId(fmt.Sprintf("vpn-gateways-%s", cfg.ProjectID))
	d.Set("vpn_gateways", vpnGateways)

	return nil
}
