package vpnconnection

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

func DataSourceVPNConnection() *schema.Resource {
	return &schema.Resource{
		ReadContext: dataSourceVPNConnectionRead,
		Description: "Use this data source to retrieve a VNPAY Cloud VPN connection by ID or name.",
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
			"vpn_gateway_id": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "The ID of the VPN gateway used by this VPN connection.",
			},
			"customer_gateway_id": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "The ID of the customer gateway used by this VPN connection.",
			},
			"vpn_public_ip_id": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "The ID of the VPN public IP associated with this VPN connection.",
			},
			"vpn_type": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"ike_profile_config": {
				Type:     schema.TypeList,
				Computed: true,
				Elem: &schema.Resource{Schema: map[string]*schema.Schema{
					"ike_version":      {Type: schema.TypeString, Computed: true},
					"ike_lifetime":     {Type: schema.TypeInt, Computed: true},
					"ike_close_action": {Type: schema.TypeString, Computed: true},
					"ike_dh":           {Type: schema.TypeString, Computed: true},
					"ike_encryption":   {Type: schema.TypeString, Computed: true},
					"ike_hash":         {Type: schema.TypeString, Computed: true},
					"ike_prf":          {Type: schema.TypeString, Computed: true},
					"ike_dpd_action":   {Type: schema.TypeString, Computed: true},
					"ike_dpd_interval": {Type: schema.TypeInt, Computed: true},
					"ike_dpd_timeout":  {Type: schema.TypeInt, Computed: true},
					"ikev2_reauth":     {Type: schema.TypeBool, Computed: true},
				}},
			},
			"ipsec_profile_config": {
				Type:     schema.TypeList,
				Computed: true,
				Elem: &schema.Resource{Schema: map[string]*schema.Schema{
					"ipsec_lifetime":         {Type: schema.TypeInt, Computed: true},
					"ipsec_pfs":              {Type: schema.TypeString, Computed: true},
					"ipsec_encryption":       {Type: schema.TypeString, Computed: true},
					"ipsec_hash":             {Type: schema.TypeString, Computed: true},
					"ipsec_disable_rekey":    {Type: schema.TypeBool, Computed: true},
					"ipsec_lifetime_bytes":   {Type: schema.TypeInt, Computed: true},
					"ipsec_lifetime_packets": {Type: schema.TypeInt, Computed: true},
				}},
			},
			"route_base_config": {
				Type:     schema.TypeList,
				Computed: true,
				Elem: &schema.Resource{Schema: map[string]*schema.Schema{
					"vti_mss": {Type: schema.TypeInt, Computed: true},
				}},
			},
			"connection_bgp_config": {
				Type:     schema.TypeList,
				Computed: true,
				Elem: &schema.Resource{Schema: map[string]*schema.Schema{
					"bgp_keepalive": {Type: schema.TypeInt, Computed: true},
					"bgp_holdtime":  {Type: schema.TypeInt, Computed: true},
				}},
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

func dataSourceVPNConnectionRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	cfg := meta.(*config.Config)

	nameFilter, nameOk := d.GetOk("name")
	if id, ok := d.GetOk("id"); ok {
		resp := &dto.VPNConnectionResponse{}
		_, err := cfg.Client.Get(ctx, client.ApiPath.VPNConnectionWithID(cfg.ProjectID, id.(string)), resp, nil)
		if err != nil {
			return diag.Errorf("Error fetching vnpaycloud_vpn_connection %s: %s", id, err)
		}
		if nameOk && resp.VPNConnection.Name != nameFilter.(string) {
			return diag.Errorf("vnpaycloud_vpn_connection %s does not match name %q", id, nameFilter.(string))
		}
		return setVPNConnectionData(d, &resp.VPNConnection)
	}

	if !nameOk {
		return diag.Errorf("One of id or name must be specified for vnpaycloud_vpn_connection")
	}

	listResp := &dto.ListVPNConnectionsResponse{}
	_, err := cfg.Client.Get(ctx, client.ApiPath.VPNConnections(cfg.ProjectID), listResp, nil)
	if err != nil {
		return diag.Errorf("Error listing vnpaycloud_vpn_connection: %s", err)
	}

	matches := make([]dto.VPNConnection, 0, 1)
	for _, vpnConnection := range listResp.VPNConnections {
		if vpnConnection.Name != nameFilter.(string) {
			continue
		}
		matches = append(matches, vpnConnection)
	}

	if len(matches) > 1 {
		return diag.Errorf("Multiple vnpaycloud_vpn_connection resources found with name %q", nameFilter.(string))
	}
	if len(matches) == 1 {
		return setVPNConnectionData(d, &matches[0])
	}

	return diag.Errorf("No vnpaycloud_vpn_connection found matching the criteria")
}

func setVPNConnectionData(d *schema.ResourceData, vpnConnection *dto.VPNConnection) diag.Diagnostics {
	d.SetId(vpnConnection.ID)
	d.Set("name", vpnConnection.Name)
	d.Set("description", vpnConnection.Description)
	d.Set("vpn_gateway_id", vpnConnection.VPNGatewayID)
	d.Set("customer_gateway_id", vpnConnection.CustomerGatewayID)
	d.Set("vpn_public_ip_id", vpnConnection.VPNPublicIPID)
	d.Set("vpn_type", vpnConnection.VPNType)
	d.Set("status", util.NormalizeStatus(vpnConnection.Status))
	d.Set("created_at", vpnConnection.CreatedAt)
	d.Set("zone_id", vpnConnection.ZoneID)
	d.Set("ike_profile_config", flattenIKEProfileConfig(vpnConnection.IKEProfileConfig))
	d.Set("ipsec_profile_config", flattenIPSecProfileConfig(vpnConnection.IPSecProfileConfig))
	d.Set("route_base_config", flattenRouteBaseConfig(vpnConnection.RouteBaseConfig))
	d.Set("connection_bgp_config", flattenConnectionBGPConfig(vpnConnection.ConnectionBGPConfig))
	return nil
}

func DataSourceVPNConnections() *schema.Resource {
	return &schema.Resource{
		ReadContext: dataSourceVPNConnectionsRead,
		Description: "Use this data source to retrieve all VNPAY Cloud VPN connections in the current project.",
		Schema: map[string]*schema.Schema{
			"vpn_connections": {
				Type:        schema.TypeList,
				Computed:    true,
				Description: "The list of VPN connections in the current project.",
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
						"vpn_gateway_id": {
							Type:     schema.TypeString,
							Computed: true,
						},
						"customer_gateway_id": {
							Type:     schema.TypeString,
							Computed: true,
						},
						"vpn_public_ip_id": {
							Type:     schema.TypeString,
							Computed: true,
						},
						"vpn_type": {
							Type:     schema.TypeString,
							Computed: true,
						},
						"ike_profile_config": {
							Type:     schema.TypeList,
							Computed: true,
							Elem: &schema.Resource{Schema: map[string]*schema.Schema{
								"ike_version":      {Type: schema.TypeString, Computed: true},
								"ike_lifetime":     {Type: schema.TypeInt, Computed: true},
								"ike_close_action": {Type: schema.TypeString, Computed: true},
								"ike_dh":           {Type: schema.TypeString, Computed: true},
								"ike_encryption":   {Type: schema.TypeString, Computed: true},
								"ike_hash":         {Type: schema.TypeString, Computed: true},
								"ike_prf":          {Type: schema.TypeString, Computed: true},
								"ike_dpd_action":   {Type: schema.TypeString, Computed: true},
								"ike_dpd_interval": {Type: schema.TypeInt, Computed: true},
								"ike_dpd_timeout":  {Type: schema.TypeInt, Computed: true},
								"ikev2_reauth":     {Type: schema.TypeBool, Computed: true},
							}},
						},
						"ipsec_profile_config": {
							Type:     schema.TypeList,
							Computed: true,
							Elem: &schema.Resource{Schema: map[string]*schema.Schema{
								"ipsec_lifetime":         {Type: schema.TypeInt, Computed: true},
								"ipsec_pfs":              {Type: schema.TypeString, Computed: true},
								"ipsec_encryption":       {Type: schema.TypeString, Computed: true},
								"ipsec_hash":             {Type: schema.TypeString, Computed: true},
								"ipsec_disable_rekey":    {Type: schema.TypeBool, Computed: true},
								"ipsec_lifetime_bytes":   {Type: schema.TypeInt, Computed: true},
								"ipsec_lifetime_packets": {Type: schema.TypeInt, Computed: true},
							}},
						},
						"route_base_config": {
							Type:     schema.TypeList,
							Computed: true,
							Elem: &schema.Resource{Schema: map[string]*schema.Schema{
								"vti_mss": {Type: schema.TypeInt, Computed: true},
							}},
						},
						"connection_bgp_config": {
							Type:     schema.TypeList,
							Computed: true,
							Elem: &schema.Resource{Schema: map[string]*schema.Schema{
								"bgp_keepalive": {Type: schema.TypeInt, Computed: true},
								"bgp_holdtime":  {Type: schema.TypeInt, Computed: true},
							}},
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

func dataSourceVPNConnectionsRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	cfg := meta.(*config.Config)

	listResp := &dto.ListVPNConnectionsResponse{}
	_, err := cfg.Client.Get(ctx, client.ApiPath.VPNConnections(cfg.ProjectID), listResp, nil)
	if err != nil {
		return diag.Errorf("Error listing vnpaycloud_vpn_connections: %s", err)
	}

	vpnConnections := make([]map[string]interface{}, 0, len(listResp.VPNConnections))
	for _, vpnConnection := range listResp.VPNConnections {
		vpnConnections = append(vpnConnections, map[string]interface{}{
			"id":                    vpnConnection.ID,
			"name":                  vpnConnection.Name,
			"description":           vpnConnection.Description,
			"vpn_gateway_id":        vpnConnection.VPNGatewayID,
			"customer_gateway_id":   vpnConnection.CustomerGatewayID,
			"vpn_public_ip_id":      vpnConnection.VPNPublicIPID,
			"vpn_type":              vpnConnection.VPNType,
			"status":                util.NormalizeStatus(vpnConnection.Status),
			"created_at":            vpnConnection.CreatedAt,
			"zone_id":               vpnConnection.ZoneID,
			"ike_profile_config":    flattenIKEProfileConfig(vpnConnection.IKEProfileConfig),
			"ipsec_profile_config":  flattenIPSecProfileConfig(vpnConnection.IPSecProfileConfig),
			"route_base_config":     flattenRouteBaseConfig(vpnConnection.RouteBaseConfig),
			"connection_bgp_config": flattenConnectionBGPConfig(vpnConnection.ConnectionBGPConfig),
		})
	}

	d.SetId(fmt.Sprintf("vpn-connections-%s", cfg.ProjectID))
	d.Set("vpn_connections", vpnConnections)
	return nil
}
