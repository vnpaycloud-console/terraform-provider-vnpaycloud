package vpnpublicip

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

func DataSourceVPNPublicIP() *schema.Resource {
	return &schema.Resource{
		ReadContext: dataSourceVPNPublicIPRead,
		Description: "Use this data source to retrieve a VNPAY Cloud VPN public IP by ID or name.",
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
			"address": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "The allocated public IP address.",
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

func dataSourceVPNPublicIPRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	cfg := meta.(*config.Config)

	nameFilter, nameOk := d.GetOk("name")
	if id, ok := d.GetOk("id"); ok {
		resp := &dto.VPNPublicIPResponse{}
		_, err := cfg.Client.Get(ctx, client.ApiPath.VPNPublicIPWithID(cfg.ProjectID, id.(string)), resp, nil)
		if err != nil {
			return diag.Errorf("Error fetching vnpaycloud_vpn_public_ip %s: %s", id, err)
		}
		if nameOk && resp.VPNPublicIP.Name != nameFilter.(string) {
			return diag.Errorf("vnpaycloud_vpn_public_ip %s does not match name %q", id, nameFilter.(string))
		}
		return setVPNPublicIPData(d, &resp.VPNPublicIP)
	}

	if !nameOk {
		return diag.Errorf("One of id or name must be specified for vnpaycloud_vpn_public_ip")
	}

	listResp := &dto.ListVPNPublicIPsResponse{}
	_, err := cfg.Client.Get(ctx, client.ApiPath.VPNPublicIPs(cfg.ProjectID), listResp, nil)
	if err != nil {
		return diag.Errorf("Error listing vnpaycloud_vpn_public_ip: %s", err)
	}

	matches := make([]dto.VPNPublicIP, 0, 1)
	for _, vpnPublicIP := range listResp.VPNPublicIPs {
		if vpnPublicIP.Name != nameFilter.(string) {
			continue
		}
		matches = append(matches, vpnPublicIP)
	}

	if len(matches) > 1 {
		return diag.Errorf("Multiple vnpaycloud_vpn_public_ip resources found with name %q", nameFilter.(string))
	}
	if len(matches) == 1 {
		return setVPNPublicIPData(d, &matches[0])
	}

	return diag.Errorf("No vnpaycloud_vpn_public_ip found matching the criteria")
}

func setVPNPublicIPData(d *schema.ResourceData, vpnPublicIP *dto.VPNPublicIP) diag.Diagnostics {
	d.SetId(vpnPublicIP.ID)
	d.Set("name", vpnPublicIP.Name)
	d.Set("description", vpnPublicIP.Description)
	d.Set("address", vpnPublicIP.FloatingIP)
	d.Set("status", util.NormalizeStatus(vpnPublicIP.Status))
	d.Set("created_at", vpnPublicIP.CreatedAt)
	d.Set("zone_id", vpnPublicIP.ZoneID)
	return nil
}

func DataSourceVPNPublicIPs() *schema.Resource {
	return &schema.Resource{
		ReadContext: dataSourceVPNPublicIPsRead,
		Description: "Use this data source to retrieve all VNPAY Cloud VPN public IPs in the current project.",
		Schema: map[string]*schema.Schema{
			"vpn_public_ips": {
				Type:        schema.TypeList,
				Computed:    true,
				Description: "The list of VPN public IPs in the current project.",
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
						"address": {
							Type:        schema.TypeString,
							Computed:    true,
							Description: "The allocated public IP address.",
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

func dataSourceVPNPublicIPsRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	cfg := meta.(*config.Config)

	listResp := &dto.ListVPNPublicIPsResponse{}
	_, err := cfg.Client.Get(ctx, client.ApiPath.VPNPublicIPs(cfg.ProjectID), listResp, nil)
	if err != nil {
		return diag.Errorf("Error listing vnpaycloud_vpn_public_ips: %s", err)
	}

	var vpnPublicIPs []map[string]interface{}
	for _, vpnPublicIP := range listResp.VPNPublicIPs {
		vpnPublicIPs = append(vpnPublicIPs, map[string]interface{}{
			"id":          vpnPublicIP.ID,
			"name":        vpnPublicIP.Name,
			"description": vpnPublicIP.Description,
			"address":     vpnPublicIP.FloatingIP,
			"status":      util.NormalizeStatus(vpnPublicIP.Status),
			"created_at":  vpnPublicIP.CreatedAt,
			"zone_id":     vpnPublicIP.ZoneID,
		})
	}

	d.SetId(fmt.Sprintf("vpn-public-ips-%s", cfg.ProjectID))
	d.Set("vpn_public_ips", vpnPublicIPs)

	return nil
}
