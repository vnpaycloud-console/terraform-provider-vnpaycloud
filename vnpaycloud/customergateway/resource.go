package customergateway

import (
	"context"
	"fmt"
	"terraform-provider-vnpaycloud/vnpaycloud/config"
	"terraform-provider-vnpaycloud/vnpaycloud/dto"
	"terraform-provider-vnpaycloud/vnpaycloud/helper/client"
	"terraform-provider-vnpaycloud/vnpaycloud/util"
	"time"

	"github.com/hashicorp/terraform-plugin-log/tflog"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/retry"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
)

func ResourceCustomerGateway() *schema.Resource {
	return &schema.Resource{
		CreateContext: resourceCustomerGatewayCreate,
		ReadContext:   resourceCustomerGatewayRead,
		UpdateContext: resourceCustomerGatewayUpdate,
		DeleteContext: resourceCustomerGatewayDelete,
		Description:   "Manages a VNPAY Cloud Customer gateway.",
		Importer: &schema.ResourceImporter{
			StateContext: func(ctx context.Context, d *schema.ResourceData, m interface{}) ([]*schema.ResourceData, error) {
				cfg := m.(*config.Config)
				resp := &dto.CustomerGatewayResponse{}
				if _, err := cfg.Client.Get(ctx, client.ApiPath.CustomerGatewayWithID(cfg.ProjectID, d.Id()), resp, nil); err != nil {
					return nil, fmt.Errorf("vnpaycloud_customer_gateway %q not found: %w", d.Id(), err)
				}

				return []*schema.ResourceData{d}, nil
			},
		},
		Timeouts: &schema.ResourceTimeout{
			Create: schema.DefaultTimeout(10 * time.Minute),
			Update: schema.DefaultTimeout(10 * time.Minute),
			Delete: schema.DefaultTimeout(10 * time.Minute),
		},
		Schema: map[string]*schema.Schema{
			"name": {
				Type:     schema.TypeString,
				Required: true,
				ValidateFunc: validation.All(
					validation.StringLenBetween(3, 255),
				),
			},
			"description": {
				Type:     schema.TypeString,
				Optional: true,
			},
			"public_ip": {
				Type:         schema.TypeString,
				Required:     true,
				Description:  "The public IPv4 address of the customer-side VPN device.",
				ValidateFunc: validation.IsIPv4Address,
			},
			"vpn_type": {
				Type:         schema.TypeString,
				Required:     true,
				ForceNew:     true,
				ValidateFunc: validation.StringInSlice([]string{"POLICY_BASED", "ROUTE_BASED"}, false),
			},
			"remote_prefixes": {
				Type:        schema.TypeSet,
				Required:    true,
				MinItems:    1,
				Description: "The remote network CIDR prefixes behind the customer gateway.",
				Elem:        &schema.Schema{Type: schema.TypeString, ValidateFunc: validation.IsCIDR},
			},
			"remote_tunnel_ip": {
				Type:         schema.TypeString,
				Optional:     true,
				Description:  "The tunnel IP address on the customer gateway side, used for route-based VPN.",
				ValidateFunc: validation.IsCIDR,
			},
			"local_tunnel_ip": {
				Type:         schema.TypeString,
				Optional:     true,
				Description:  "The tunnel IP address on the VNPAY Cloud side, used for route-based VPN.",
				ValidateFunc: validation.IsCIDR,
			},
			"routing_mode": {
				Type:         schema.TypeString,
				Optional:     true,
				Description:  "The routing mode for route-based VPN. Use NONE for policy-based VPN, STATIC for static route-based VPN, or DYNAMIC for BGP.",
				ValidateFunc: validation.StringInSlice([]string{"NONE", "DYNAMIC", "STATIC"}, false),
				Default:      "NONE",
			},
			"bgp_config": {
				Type:        schema.TypeList,
				Optional:    true,
				MaxItems:    1,
				Description: "The BGP configuration. This is only valid for route-based VPN with DYNAMIC routing mode.",
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"local_as": {
							Type:        schema.TypeInt,
							Required:    true,
							Description: "The local BGP autonomous system number.",
						},
						"peer_as": {
							Type:        schema.TypeInt,
							Required:    true,
							Description: "The peer BGP autonomous system number.",
						},
						"as_path": {
							Type:        schema.TypeString,
							Required:    true,
							Description: "The BGP AS path to advertise for this customer gateway.",
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
		},
	}
}

func expandBGPConfig(d *schema.ResourceData) *dto.BGPConfig {
	bgpConfigList := d.Get("bgp_config").([]interface{})
	if len(bgpConfigList) == 0 || bgpConfigList[0] == nil {
		return nil
	}

	bgpMap := bgpConfigList[0].(map[string]interface{})

	return &dto.BGPConfig{
		LocalAs: int64(bgpMap["local_as"].(int)),
		PeerAs:  int64(bgpMap["peer_as"].(int)),
		AsPath:  bgpMap["as_path"].(string),
	}
}

func expandRemotePrefixes(d *schema.ResourceData) []string {
	raw := d.Get("remote_prefixes").(*schema.Set).List()

	result := make([]string, 0, len(raw))
	for _, v := range raw {
		result = append(result, v.(string))
	}

	return result
}

func resourceCustomerGatewayCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	cfg := meta.(*config.Config)

	createOpts := dto.CreateCustomerGatewayRequest{
		Name:           d.Get("name").(string),
		Description:    d.Get("description").(string),
		PublicIP:       d.Get("public_ip").(string),
		VPNType:        d.Get("vpn_type").(string),
		RemotePrefixes: expandRemotePrefixes(d),
		RemoteTunnelIP: d.Get("remote_tunnel_ip").(string),
		LocalTunnelIP:  d.Get("local_tunnel_ip").(string),
		RoutingMode:    d.Get("routing_mode").(string),
		BGPConfig:      expandBGPConfig(d),
	}

	tflog.Debug(ctx, "vnpaycloud_customer_gateway create options", map[string]interface{}{"create_opts": createOpts})

	createResp := &dto.CustomerGatewayResponse{}
	_, err := cfg.Client.Post(ctx, client.ApiPath.CustomerGateways(cfg.ProjectID), createOpts, createResp, nil)
	if err != nil {
		return diag.Errorf("Error creating vnpaycloud_customer_gateway: %s", err)
	}

	d.SetId(createResp.CustomerGateway.ID)

	stateConf := &retry.StateChangeConf{
		Pending:    []string{"initiating", "creating"},
		Target:     []string{"active"},
		Refresh:    customerGatewayStateRefreshFunc(ctx, cfg.Client, cfg.ProjectID, createResp.CustomerGateway.ID),
		Timeout:    d.Timeout(schema.TimeoutCreate),
		Delay:      5 * time.Second,
		MinTimeout: 3 * time.Second,
	}

	_, err = stateConf.WaitForStateContext(ctx)
	if err != nil {
		return diag.Errorf("Error waiting for vnpaycloud_customer_gateway %s to become ready: %s", createResp.CustomerGateway.ID, err)
	}

	return resourceCustomerGatewayRead(ctx, d, meta)
}

func resourceCustomerGatewayRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	cfg := meta.(*config.Config)

	cgResp := &dto.CustomerGatewayResponse{}
	_, err := cfg.Client.Get(ctx, client.ApiPath.CustomerGatewayWithID(cfg.ProjectID, d.Id()), cgResp, nil)
	if err != nil {
		return diag.FromErr(util.CheckNotFound(d, err, "Error retrieving vnpaycloud_customer_gateway"))
	}

	tflog.Debug(ctx, "Retrieved vnpaycloud_customer_gateway "+d.Id(), map[string]interface{}{"customer_gateway": cgResp.CustomerGateway})

	d.Set("name", cgResp.CustomerGateway.Name)
	d.Set("description", cgResp.CustomerGateway.Description)
	d.Set("public_ip", cgResp.CustomerGateway.PublicIP)
	d.Set("vpn_type", cgResp.CustomerGateway.VPNType)
	d.Set("status", util.NormalizeStatus(cgResp.CustomerGateway.Status))
	d.Set("remote_prefixes", cgResp.CustomerGateway.RemotePrefixes)
	d.Set("remote_tunnel_ip", cgResp.CustomerGateway.RemoteTunnelIP)
	d.Set("local_tunnel_ip", cgResp.CustomerGateway.LocalTunnelIP)
	d.Set("routing_mode", cgResp.CustomerGateway.RoutingMode)
	d.Set("created_at", cgResp.CustomerGateway.CreatedAt)
	if cgResp.CustomerGateway.BGPConfig != nil {
		bgpConfig := []map[string]interface{}{
			{
				"local_as": int(cgResp.CustomerGateway.BGPConfig.LocalAs),
				"peer_as":  int(cgResp.CustomerGateway.BGPConfig.PeerAs),
				"as_path":  cgResp.CustomerGateway.BGPConfig.AsPath,
			},
		}
		d.Set("bgp_config", bgpConfig)
	} else {
		d.Set("bgp_config", nil)
	}

	return nil
}

func resourceCustomerGatewayUpdate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	cfg := meta.(*config.Config)

	updateOpts := dto.UpdateCustomerGatewayRequest{
		Name:           d.Get("name").(string),
		Description:    d.Get("description").(string),
		PublicIP:       d.Get("public_ip").(string),
		VPNType:        d.Get("vpn_type").(string),
		RemotePrefixes: expandRemotePrefixes(d),
		RemoteTunnelIP: d.Get("remote_tunnel_ip").(string),
		LocalTunnelIP:  d.Get("local_tunnel_ip").(string),
		RoutingMode:    d.Get("routing_mode").(string),
		BGPConfig:      expandBGPConfig(d),
	}

	tflog.Debug(ctx, "vnpaycloud_customer_gateway update options", map[string]interface{}{"update_opts": updateOpts})

	_, err := cfg.Client.Put(ctx, client.ApiPath.CustomerGatewayWithID(cfg.ProjectID, d.Id()), updateOpts, nil, nil)
	if err != nil {
		return diag.Errorf("Error updating vnpaycloud_customer_gateway %s: %s", d.Id(), err)
	}

	return resourceCustomerGatewayRead(ctx, d, meta)
}

func resourceCustomerGatewayDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	cfg := meta.(*config.Config)

	cgResp := &dto.CustomerGatewayResponse{}
	_, err := cfg.Client.Get(ctx, client.ApiPath.CustomerGatewayWithID(cfg.ProjectID, d.Id()), cgResp, nil)
	if err != nil {
		return diag.FromErr(util.CheckDeleted(d, err, "Error retrieving vnpaycloud_customer_gateway"))
	}

	if cgResp.CustomerGateway.Status != "deleting" {
		if _, err := cfg.Client.Delete(ctx, client.ApiPath.CustomerGatewayWithID(cfg.ProjectID, d.Id()), nil); err != nil {
			return diag.FromErr(util.CheckDeleted(d, err, "Error deleting vnpaycloud_customer_gateway"))
		}
	}

	stateConf := &retry.StateChangeConf{
		Pending:    []string{"deleting", "active"},
		Target:     []string{"deleted"},
		Refresh:    customerGatewayStateRefreshFunc(ctx, cfg.Client, cfg.ProjectID, d.Id()),
		Timeout:    d.Timeout(schema.TimeoutDelete),
		Delay:      5 * time.Second,
		MinTimeout: 3 * time.Second,
	}

	_, err = stateConf.WaitForStateContext(ctx)
	if err != nil {
		return diag.Errorf("Error waiting for vnpaycloud_customer_gateway %s to delete: %s", d.Id(), err)
	}

	return nil
}
