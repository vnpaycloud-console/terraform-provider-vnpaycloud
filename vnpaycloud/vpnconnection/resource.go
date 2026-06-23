package vpnconnection

import (
	"context"
	"fmt"
	"regexp"
	"strings"
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

// preSharedKeyRegex matches the backend's allowed pre-shared key character set:
// letters, digits, '-', '_' and '.'. The backend rejects any other character.
var preSharedKeyRegex = regexp.MustCompile(`^[a-zA-Z0-9._-]+$`)

// vpnGatewayMutexKey returns the MutexKV key used to serialize mutating
// operations on a single VPN gateway. It must match the key format used by the
// vpngateway VPC-attachment resource so connection create/delete and VPC
// attach/detach never run concurrently against the same gateway.
func vpnGatewayMutexKey(vpnGatewayID string) string {
	return "vnpaycloud_vpn_gateway/" + vpnGatewayID
}

// validateIPSecLifetimeUnit enforces the backend rule that ipsec_lifetime_bytes
// and ipsec_lifetime_packets must be 0 (disabled) or at least 1024. Values
// between 1 and 1023 are accepted by a plain IntAtLeast(0) check but rejected by
// the backend, so validate them here to fail fast during plan.
func validateIPSecLifetimeUnit(i interface{}, k string) ([]string, []error) {
	v, ok := i.(int)
	if !ok {
		return nil, []error{fmt.Errorf("expected type of %q to be int", k)}
	}
	if v != 0 && v < 1024 {
		return nil, []error{fmt.Errorf("%q must be 0 (disabled) or at least 1024, got %d", k, v)}
	}
	return nil, nil
}

var allowedIKEVersions = []string{
	"IKE_V1",
	"IKE_V2",
}

var allowedVPNConnectionGroups = []string{
	"GROUP_1",
	"GROUP_2",
	"GROUP_5",
	"GROUP_14",
	"GROUP_15",
	"GROUP_16",
	"GROUP_17",
	"GROUP_18",
	"GROUP_19",
	"GROUP_20",
	"GROUP_21",
	"GROUP_22",
	"GROUP_23",
	"GROUP_24",
	"GROUP_25",
	"GROUP_26",
	"GROUP_27",
	"GROUP_28",
	"GROUP_29",
	"GROUP_30",
	"GROUP_31",
	"GROUP_32",
}

var allowedIKECloseActions = []string{
	"NONE",
	"TRAP",
	"START",
}

var allowedVPNConnectionEncryptionAlgorithms = []string{
	"AES128",
	"AES192",
	"AES256",
	"AES128_GCM96",
	"AES128_GCM128",
	"AES256_GCM96",
	"AES256_GCM128",
}

var allowedVPNConnectionHashAlgorithms = []string{
	"MD5",
	"MD5_128",
	"SHA1",
	"SHA1_160",
	"SHA256",
	"SHA256_96",
	"SHA384",
	"SHA512",
	"AES_XCBC",
	"AES_CMAC",
	"AES128_GMAC",
	"AES192_GMAC",
	"AES256_GMAC",
}

var allowedVPNConnectionPRFAlgorithms = []string{
	"MD5",
	"SHA1",
	"AES_XCBC",
	"AES_CMAC",
	"SHA256",
	"SHA384",
	"SHA512",
}

var allowedIKEDPDActions = []string{
	"TRAP",
	"CLEAR",
	"RESTART",
}

func ResourceVPNConnection() *schema.Resource {
	return &schema.Resource{
		CreateContext: resourceVPNConnectionCreate,
		ReadContext:   resourceVPNConnectionRead,
		UpdateContext: resourceVPNConnectionUpdate,
		DeleteContext: resourceVPNConnectionDelete,
		CustomizeDiff: func(_ context.Context, d *schema.ResourceDiff, _ interface{}) error {
			if d.Id() == "" {
				return nil
			}
			if d.HasChange("name") || d.HasChange("description") {
				return fmt.Errorf("vnpaycloud_vpn_connection does not currently support updating name or description in Terraform; change is rejected to avoid recreating the VPN tunnel")
			}
			return nil
		},
		Description: "Manages a VNPAY Cloud VPN connection.",
		Importer: &schema.ResourceImporter{
			StateContext: func(ctx context.Context, d *schema.ResourceData, m interface{}) ([]*schema.ResourceData, error) {
				cfg := m.(*config.Config)
				resp := &dto.VPNConnectionResponse{}
				if _, err := cfg.Client.Get(ctx, client.ApiPath.VPNConnectionWithID(cfg.ProjectID, d.Id()), resp, nil); err != nil {
					return nil, fmt.Errorf("vnpaycloud_vpn_connection %q not found: %w", d.Id(), err)
				}

				return []*schema.ResourceData{d}, nil
			},
		},
		Timeouts: &schema.ResourceTimeout{
			Create: schema.DefaultTimeout(10 * time.Minute),
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
			"vpn_gateway_id": {
				Type:        schema.TypeString,
				Required:    true,
				ForceNew:    true,
				Description: "The ID of the VPN gateway used by this VPN connection.",
			},
			"customer_gateway_id": {
				Type:        schema.TypeString,
				Required:    true,
				ForceNew:    true,
				Description: "The ID of the customer gateway used by this VPN connection.",
			},
			"vpn_public_ip_id": {
				Type:        schema.TypeString,
				Required:    true,
				ForceNew:    true,
				Description: "The ID of the VPN public IP associated with this VPN connection.",
			},
			"vpn_type": {
				Type:         schema.TypeString,
				Required:     true,
				ForceNew:     true,
				ValidateFunc: validation.StringInSlice([]string{"POLICY_BASED", "ROUTE_BASED"}, false),
			},
			"ipsec_auth_config": {
				Type:        schema.TypeList,
				Required:    true,
				Sensitive:   true,
				ForceNew:    true,
				MaxItems:    1,
				Description: "The IPSec authentication configuration for this VPN connection.",
				Elem: &schema.Resource{Schema: map[string]*schema.Schema{
					"pre_shared_key": {
						Type:        schema.TypeString,
						Required:    true,
						Sensitive:   true,
						Description: "The pre-shared key used for IPSec authentication.",
						ValidateFunc: validation.All(
							validation.StringLenBetween(8, 255),
							validation.StringMatch(
								preSharedKeyRegex,
								"pre_shared_key must contain only letters, digits, '-', '_' or '.'",
							),
						),
					},
				}},
			},
			"ike_profile_config": {
				Type:        schema.TypeList,
				Optional:    true,
				Computed:    true,
				ForceNew:    true,
				MaxItems:    1,
				Description: "The IKE profile configuration for this VPN connection.",
				Elem: &schema.Resource{Schema: map[string]*schema.Schema{
					"ike_version": {
						Type:         schema.TypeString,
						Optional:     true,
						ForceNew:     true,
						ValidateFunc: validation.StringInSlice(allowedIKEVersions, false),
						Description:  "The IKE protocol version.",
						Default:      "IKE_V2",
					},
					"ike_lifetime": {
						Type:         schema.TypeInt,
						Optional:     true,
						ForceNew:     true,
						Description:  "The IKE SA lifetime in seconds.",
						ValidateFunc: validation.IntBetween(0, 86400),
						Default:      28800,
					},
					"ike_close_action": {
						Type:         schema.TypeString,
						Optional:     true,
						ForceNew:     true,
						Description:  "The action to take when the IKE SA is closed.",
						ValidateFunc: validation.StringInSlice(allowedIKECloseActions, false),
						Default:      "START",
					},
					"ike_dh": {
						Type:         schema.TypeString,
						Optional:     true,
						ForceNew:     true,
						Description:  "The IKE Diffie-Hellman group.",
						ValidateFunc: validation.StringInSlice(allowedVPNConnectionGroups, false),
						Default:      "GROUP_14",
					},
					"ike_encryption": {
						Type:         schema.TypeString,
						Optional:     true,
						ForceNew:     true,
						Description:  "The IKE encryption algorithm.",
						ValidateFunc: validation.StringInSlice(allowedVPNConnectionEncryptionAlgorithms, false),
						Default:      "AES128_GCM96",
					},
					"ike_hash": {
						Type:         schema.TypeString,
						Optional:     true,
						ForceNew:     true,
						Description:  "The IKE integrity hash algorithm.",
						ValidateFunc: validation.StringInSlice(allowedVPNConnectionHashAlgorithms, false),
						Default:      "SHA256",
					},
					"ike_prf": {
						Type:         schema.TypeString,
						Optional:     true,
						ForceNew:     true,
						Description:  "The IKE pseudo-random function algorithm.",
						ValidateFunc: validation.StringInSlice(allowedVPNConnectionPRFAlgorithms, false),
						Default:      "SHA1",
					},
					"ike_dpd_action": {
						Type:         schema.TypeString,
						Optional:     true,
						ForceNew:     true,
						Description:  "The IKE dead peer detection action.",
						ValidateFunc: validation.StringInSlice(allowedIKEDPDActions, false),
						Default:      "CLEAR",
					},
					"ike_dpd_interval": {
						Type:         schema.TypeInt,
						Optional:     true,
						ForceNew:     true,
						Description:  "The IKE dead peer detection interval in seconds.",
						ValidateFunc: validation.IntBetween(2, 86400),
						Default:      30,
					},
					"ike_dpd_timeout": {
						Type:         schema.TypeInt,
						Optional:     true,
						ForceNew:     true,
						Description:  "The IKE dead peer detection timeout in seconds.",
						ValidateFunc: validation.IntBetween(2, 86400),
						Default:      120,
					},
					"ikev2_reauth": {
						Type:        schema.TypeBool,
						Optional:    true,
						ForceNew:    true,
						Description: "Whether IKEv2 reauthentication is enabled.",
						Default:     true,
					},
				}},
			},
			"ipsec_profile_config": {
				Type:        schema.TypeList,
				Optional:    true,
				Computed:    true,
				ForceNew:    true,
				MaxItems:    1,
				Description: "The IPSec profile configuration for this VPN connection.",
				Elem: &schema.Resource{Schema: map[string]*schema.Schema{
					"ipsec_lifetime": {
						Type:         schema.TypeInt,
						Optional:     true,
						ForceNew:     true,
						Description:  "The IPSec SA lifetime in seconds.",
						ValidateFunc: validation.IntBetween(30, 86400),
						Default:      3600,
					},
					"ipsec_pfs": {
						Type:         schema.TypeString,
						Optional:     true,
						ForceNew:     true,
						Description:  "The IPSec perfect forward secrecy group.",
						ValidateFunc: validation.StringInSlice(allowedVPNConnectionGroups, false),
						Default:      "GROUP_14",
					},
					"ipsec_encryption": {
						Type:         schema.TypeString,
						Optional:     true,
						ForceNew:     true,
						Description:  "The IPSec encryption algorithm.",
						ValidateFunc: validation.StringInSlice(allowedVPNConnectionEncryptionAlgorithms, false),
						Default:      "AES256",
					},
					"ipsec_hash": {
						Type:         schema.TypeString,
						Optional:     true,
						ForceNew:     true,
						Description:  "The IPSec integrity hash algorithm.",
						ValidateFunc: validation.StringInSlice(allowedVPNConnectionHashAlgorithms, false),
						Default:      "SHA256",
					},
					"ipsec_disable_rekey": {
						Type:        schema.TypeBool,
						Optional:    true,
						ForceNew:    true,
						Description: "Whether IPSec rekey is disabled.",
						Default:     false,
					},
					"ipsec_lifetime_bytes": {
						Type:         schema.TypeInt,
						Optional:     true,
						ForceNew:     true,
						Description:  "The IPSec SA lifetime in bytes. Must be 0 (disabled) or at least 1024.",
						ValidateFunc: validateIPSecLifetimeUnit,
						Default:      0,
					},
					"ipsec_lifetime_packets": {
						Type:         schema.TypeInt,
						Optional:     true,
						ForceNew:     true,
						Description:  "The IPSec SA lifetime in packets. Must be 0 (disabled) or at least 1024.",
						ValidateFunc: validateIPSecLifetimeUnit,
						Default:      0,
					},
				}},
			},
			"route_base_config": {
				Type:        schema.TypeList,
				Optional:    true,
				ForceNew:    true,
				MaxItems:    1,
				Description: "The route-based VPN configuration. This is only used for route-based VPN connections.",
				Elem: &schema.Resource{Schema: map[string]*schema.Schema{
					"vti_mss": {
						Type:        schema.TypeInt,
						Optional:    true,
						ForceNew:    true,
						Description: "The TCP MSS value configured on the VTI interface.",
						Default:     1350,
					},
				}},
			},
			"connection_bgp_config": {
				Type:        schema.TypeList,
				Optional:    true,
				ForceNew:    true,
				MaxItems:    1,
				Description: "The BGP timer configuration for this VPN connection.",
				Elem: &schema.Resource{Schema: map[string]*schema.Schema{
					"bgp_keepalive": {
						Type:         schema.TypeInt,
						Optional:     true,
						ForceNew:     true,
						Description:  "The BGP keepalive interval in seconds.",
						ValidateFunc: validation.IntBetween(4, 65535),
						Default:      60,
					},
					"bgp_holdtime": {
						Type:         schema.TypeInt,
						Optional:     true,
						ForceNew:     true,
						Description:  "The BGP hold time in seconds.",
						ValidateFunc: validation.IntBetween(4, 65535),
						Default:      180,
					},
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
		},
	}
}

func resourceVPNConnectionCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	cfg := meta.(*config.Config)

	// Serialize mutating operations on the parent VPN gateway (this key matches
	// the one the vpngateway VPC-attachment resource uses) so a connection
	// create does not race a VPC attach/detach or another connection on the
	// same gateway, which the backend rejects as "being modified by another
	// operation".
	gatewayMutexKey := vpnGatewayMutexKey(d.Get("vpn_gateway_id").(string))
	cfg.MutexKV.Lock(gatewayMutexKey)
	defer cfg.MutexKV.Unlock(gatewayMutexKey)

	createOpts := dto.CreateVPNConnectionRequest{
		Name:                d.Get("name").(string),
		Description:         d.Get("description").(string),
		VPNGatewayID:        d.Get("vpn_gateway_id").(string),
		CustomerGatewayID:   d.Get("customer_gateway_id").(string),
		VPNPublicIPID:       d.Get("vpn_public_ip_id").(string),
		VPNType:             d.Get("vpn_type").(string),
		IPSecAuthConfig:     expandIPSecAuthConfig(d),
		IKEProfileConfig:    expandIKEProfileConfig(d),
		IPSecProfileConfig:  expandIPSecProfileConfig(d),
		RouteBaseConfig:     expandRouteBaseConfig(d),
		ConnectionBGPConfig: expandConnectionBGPConfig(d),
	}

	tflog.Debug(ctx, "vnpaycloud_vpn_connection create options", map[string]interface{}{
		"name":                createOpts.Name,
		"vpn_gateway_id":      createOpts.VPNGatewayID,
		"customer_gateway_id": createOpts.CustomerGatewayID,
		"vpn_public_ip_id":    createOpts.VPNPublicIPID,
		"vpn_type":            createOpts.VPNType,
	})

	createResp := &dto.VPNConnectionResponse{}
	createErr := retry.RetryContext(ctx, d.Timeout(schema.TimeoutCreate), func() *retry.RetryError {
		createResp = &dto.VPNConnectionResponse{}
		_, err := cfg.Client.Post(ctx, client.ApiPath.VPNConnections(cfg.ProjectID), createOpts, createResp, nil)
		if err == nil {
			return nil
		}

		if strings.Contains(strings.ToLower(err.Error()), "currently being modified by another operation") {
			return retry.RetryableError(err)
		}

		return retry.NonRetryableError(err)
	})
	if createErr != nil {
		return diag.Errorf("Error creating vnpaycloud_vpn_connection: %s", createErr)
	}

	d.SetId(createResp.VPNConnection.ID)

	stateConf := &retry.StateChangeConf{
		Pending:    []string{"initiating", "creating"},
		Target:     []string{"active"},
		Refresh:    vpnConnectionStateRefreshFunc(ctx, cfg.Client, cfg.ProjectID, createResp.VPNConnection.ID),
		Timeout:    d.Timeout(schema.TimeoutCreate),
		Delay:      5 * time.Second,
		MinTimeout: 3 * time.Second,
	}

	_, err := stateConf.WaitForStateContext(ctx)
	if err != nil {
		return diag.Errorf("Error waiting for vnpaycloud_vpn_connection %s to become ready: %s", createResp.VPNConnection.ID, err)
	}

	return resourceVPNConnectionRead(ctx, d, meta)
}

func resourceVPNConnectionRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	cfg := meta.(*config.Config)

	resp := &dto.VPNConnectionResponse{}
	_, err := cfg.Client.Get(ctx, client.ApiPath.VPNConnectionWithID(cfg.ProjectID, d.Id()), resp, nil)
	if err != nil {
		return diag.FromErr(util.CheckNotFound(d, err, "Error retrieving vnpaycloud_vpn_connection"))
	}

	d.Set("name", resp.VPNConnection.Name)
	d.Set("description", resp.VPNConnection.Description)
	d.Set("vpn_gateway_id", resp.VPNConnection.VPNGatewayID)
	d.Set("customer_gateway_id", resp.VPNConnection.CustomerGatewayID)
	d.Set("vpn_public_ip_id", resp.VPNConnection.VPNPublicIPID)
	d.Set("vpn_type", resp.VPNConnection.VPNType)
	d.Set("status", util.NormalizeStatus(resp.VPNConnection.Status))
	d.Set("created_at", resp.VPNConnection.CreatedAt)
	d.Set("ike_profile_config", flattenIKEProfileConfig(resp.VPNConnection.IKEProfileConfig))
	d.Set("ipsec_profile_config", flattenIPSecProfileConfig(resp.VPNConnection.IPSecProfileConfig))
	d.Set("route_base_config", flattenRouteBaseConfig(resp.VPNConnection.RouteBaseConfig))
	d.Set("connection_bgp_config", flattenConnectionBGPConfig(resp.VPNConnection.ConnectionBGPConfig))
	return nil
}

func resourceVPNConnectionUpdate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	if d.HasChange("name") || d.HasChange("description") {
		return diag.Errorf("vnpaycloud_vpn_connection does not currently support updating name or description in Terraform; change is rejected to avoid recreating the VPN tunnel")
	}

	return resourceVPNConnectionRead(ctx, d, meta)
}

func resourceVPNConnectionDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	cfg := meta.(*config.Config)

	// Serialize with other mutating operations on the same VPN gateway (see
	// resourceVPNConnectionCreate).
	gatewayMutexKey := vpnGatewayMutexKey(d.Get("vpn_gateway_id").(string))
	cfg.MutexKV.Lock(gatewayMutexKey)
	defer cfg.MutexKV.Unlock(gatewayMutexKey)

	resp := &dto.VPNConnectionResponse{}
	_, err := cfg.Client.Get(ctx, client.ApiPath.VPNConnectionWithID(cfg.ProjectID, d.Id()), resp, nil)
	if err != nil {
		return diag.FromErr(util.CheckDeleted(d, err, "Error retrieving vnpaycloud_vpn_connection"))
	}

	if resp.VPNConnection.Status != "deleting" {
		deleteErr := retry.RetryContext(ctx, d.Timeout(schema.TimeoutDelete), func() *retry.RetryError {
			_, err := cfg.Client.Delete(ctx, client.ApiPath.VPNConnectionWithID(cfg.ProjectID, d.Id()), nil)
			if err == nil {
				return nil
			}

			if strings.Contains(strings.ToLower(err.Error()), "currently being modified by another operation") {
				return retry.RetryableError(err)
			}

			return retry.NonRetryableError(err)
		})

		if deleteErr != nil {
			return diag.Errorf("Error deleting vnpaycloud_vpn_connection: %s", deleteErr)
		}
	}

	stateConf := &retry.StateChangeConf{
		Pending:    []string{"deleting", "active"},
		Target:     []string{"deleted"},
		Refresh:    vpnConnectionStateRefreshFunc(ctx, cfg.Client, cfg.ProjectID, d.Id()),
		Timeout:    d.Timeout(schema.TimeoutDelete),
		Delay:      5 * time.Second,
		MinTimeout: 3 * time.Second,
	}

	_, err = stateConf.WaitForStateContext(ctx)
	if err != nil {
		return diag.Errorf("Error waiting for vnpaycloud_vpn_connection %s to delete: %s", d.Id(), err)
	}

	return nil
}

func expandIPSecAuthConfig(d *schema.ResourceData) *dto.IPSecAuthenticationConfig {
	raw := d.Get("ipsec_auth_config").([]interface{})
	if len(raw) > 0 && raw[0] != nil {
		m := raw[0].(map[string]interface{})
		return &dto.IPSecAuthenticationConfig{PSK: m["pre_shared_key"].(string)}
	}

	return nil
}

func expandIKEProfileConfig(d *schema.ResourceData) *dto.IKEProfileConfig {
	raw := d.Get("ike_profile_config").([]interface{})
	if len(raw) == 0 || raw[0] == nil {
		return nil
	}
	m := raw[0].(map[string]interface{})
	return &dto.IKEProfileConfig{
		IKEVersion:     m["ike_version"].(string),
		IKELifetime:    m["ike_lifetime"].(int),
		IKECloseAction: m["ike_close_action"].(string),
		IKEDH:          m["ike_dh"].(string),
		IKEEncryption:  m["ike_encryption"].(string),
		IKEHash:        m["ike_hash"].(string),
		IKEPRF:         m["ike_prf"].(string),
		IKEDPDAction:   m["ike_dpd_action"].(string),
		IKEDPDInterval: m["ike_dpd_interval"].(int),
		IKEDPDTimeout:  m["ike_dpd_timeout"].(int),
		IKEV2Reauth:    m["ikev2_reauth"].(bool),
	}
}

func expandIPSecProfileConfig(d *schema.ResourceData) *dto.IPSecProfileConfig {
	raw := d.Get("ipsec_profile_config").([]interface{})
	if len(raw) == 0 || raw[0] == nil {
		return nil
	}
	m := raw[0].(map[string]interface{})
	return &dto.IPSecProfileConfig{
		IPSecLifetime:        m["ipsec_lifetime"].(int),
		IPSecPFS:             m["ipsec_pfs"].(string),
		IPSecEncryption:      m["ipsec_encryption"].(string),
		IPSecHash:            m["ipsec_hash"].(string),
		IPSecDisableRekey:    m["ipsec_disable_rekey"].(bool),
		IPSecLifetimeBytes:   int64(m["ipsec_lifetime_bytes"].(int)),
		IPSecLifetimePackets: int64(m["ipsec_lifetime_packets"].(int)),
	}
}

func expandRouteBaseConfig(d *schema.ResourceData) *dto.RouteBaseConfig {
	raw := d.Get("route_base_config").([]interface{})
	if len(raw) == 0 || raw[0] == nil {
		return nil
	}
	m := raw[0].(map[string]interface{})
	return &dto.RouteBaseConfig{VTIMSS: m["vti_mss"].(int)}
}

func expandConnectionBGPConfig(d *schema.ResourceData) *dto.ConnectionBGPConfig {
	raw := d.Get("connection_bgp_config").([]interface{})
	if len(raw) == 0 || raw[0] == nil {
		return nil
	}
	m := raw[0].(map[string]interface{})
	return &dto.ConnectionBGPConfig{
		BGPKeepalive: m["bgp_keepalive"].(int),
		BGPHoldtime:  m["bgp_holdtime"].(int),
	}
}

func flattenIKEProfileConfig(cfg *dto.IKEProfileConfig) []map[string]interface{} {
	if cfg == nil {
		return nil
	}
	return []map[string]interface{}{
		{
			"ike_version":      cfg.IKEVersion,
			"ike_lifetime":     cfg.IKELifetime,
			"ike_close_action": cfg.IKECloseAction,
			"ike_dh":           cfg.IKEDH,
			"ike_encryption":   cfg.IKEEncryption,
			"ike_hash":         cfg.IKEHash,
			"ike_prf":          cfg.IKEPRF,
			"ike_dpd_action":   cfg.IKEDPDAction,
			"ike_dpd_interval": cfg.IKEDPDInterval,
			"ike_dpd_timeout":  cfg.IKEDPDTimeout,
			"ikev2_reauth":     cfg.IKEV2Reauth,
		},
	}
}

func flattenIPSecProfileConfig(cfg *dto.IPSecProfileConfig) []map[string]interface{} {
	if cfg == nil {
		return nil
	}
	return []map[string]interface{}{
		{
			"ipsec_lifetime":         cfg.IPSecLifetime,
			"ipsec_pfs":              cfg.IPSecPFS,
			"ipsec_encryption":       cfg.IPSecEncryption,
			"ipsec_hash":             cfg.IPSecHash,
			"ipsec_disable_rekey":    cfg.IPSecDisableRekey,
			"ipsec_lifetime_bytes":   int(cfg.IPSecLifetimeBytes),
			"ipsec_lifetime_packets": int(cfg.IPSecLifetimePackets),
		},
	}
}

func flattenRouteBaseConfig(cfg *dto.RouteBaseConfig) []map[string]interface{} {
	if cfg == nil {
		return nil
	}
	return []map[string]interface{}{
		{
			"vti_mss": cfg.VTIMSS,
		},
	}
}

func flattenConnectionBGPConfig(cfg *dto.ConnectionBGPConfig) []map[string]interface{} {
	if cfg == nil {
		return nil
	}
	return []map[string]interface{}{
		{
			"bgp_keepalive": cfg.BGPKeepalive,
			"bgp_holdtime":  cfg.BGPHoldtime,
		},
	}
}
