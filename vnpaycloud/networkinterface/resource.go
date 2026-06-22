package networkinterface

import (
	"context"
	"errors"
	"terraform-provider-vnpaycloud/vnpaycloud/config"
	"terraform-provider-vnpaycloud/vnpaycloud/dto"
	"terraform-provider-vnpaycloud/vnpaycloud/helper/client"
	"terraform-provider-vnpaycloud/vnpaycloud/util"
	"time"

	"github.com/hashicorp/go-cty/cty"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/retry"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

func ResourceNetworkInterface() *schema.Resource {
	return &schema.Resource{
		CreateContext: resourceNetworkInterfaceCreate,
		ReadContext:   resourceNetworkInterfaceRead,
		UpdateContext: resourceNetworkInterfaceUpdate,
		DeleteContext: resourceNetworkInterfaceDelete,
		CustomizeDiff: validateNetworkInterfaceDiff,
		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},
		Timeouts: &schema.ResourceTimeout{
			Create: schema.DefaultTimeout(10 * time.Minute),
			Delete: schema.DefaultTimeout(10 * time.Minute),
		},
		Schema: map[string]*schema.Schema{
			"name": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},
			"subnet_id": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},
			"ip_address": {
				Type:     schema.TypeString,
				Optional: true,
				Computed: true,
				ForceNew: true,
			},
			"description": {
				Type:     schema.TypeString,
				Optional: true,
			},
			"reserved": {
				Type:     schema.TypeBool,
				Optional: true,
				Computed: true,
			},
			"virtual_ip": {
				Type:     schema.TypeBool,
				Optional: true,
				Computed: true,
			},
			"allowed_address_pairs": {
				Type:     schema.TypeList,
				Optional: true,
				Computed: true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"ip_address": {
							Type:     schema.TypeString,
							Required: true,
						},
						"mac_address": {
							Type:     schema.TypeString,
							Optional: true,
							Computed: true,
						},
					},
				},
			},
			"network_id": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"mac_address": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"status": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"security_groups": {
				Type:     schema.TypeSet,
				Optional: true,
				Computed: true,
				Elem:     &schema.Schema{Type: schema.TypeString},
			},
			"port_security_enabled": {
				Type:     schema.TypeBool,
				Optional: true,
				Computed: true,
			},
			"network_type": {
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

func validateNetworkInterfaceDiff(_ context.Context, d *schema.ResourceDiff, _ interface{}) error {
	raw := d.GetRawConfig()

	if emptySecurityGroupsConfig(raw) {
		return errors.New("security_groups cannot be empty; omit it to use default/system security groups")
	}

	sgs := d.Get("security_groups").(*schema.Set)
	if invalidNetworkInterfaceSecurityGroupsConfig(raw, sgs.Len()) {
		return errors.New("security_groups requires port_security_enabled to be true")
	}

	return nil
}

func emptySecurityGroupsConfig(raw cty.Value) bool {
	if raw.IsNull() || !raw.IsKnown() || !raw.Type().IsObjectType() {
		return false
	}
	if !raw.Type().HasAttribute("security_groups") {
		return false
	}

	securityGroups := raw.GetAttr("security_groups")
	if securityGroups.IsNull() || !securityGroups.IsKnown() {
		return false
	}

	return securityGroups.LengthInt() == 0
}

func invalidNetworkInterfaceSecurityGroupsConfig(raw cty.Value, securityGroupsLen int) bool {
	if raw.IsNull() || !raw.IsKnown() || !raw.Type().IsObjectType() {
		return false
	}
	if !raw.Type().HasAttribute("port_security_enabled") || !raw.Type().HasAttribute("security_groups") {
		return false
	}

	portSecurity := raw.GetAttr("port_security_enabled")
	if !portSecurity.IsKnown() || portSecurity.IsNull() || !portSecurity.False() {
		return false
	}

	securityGroups := raw.GetAttr("security_groups")
	if !securityGroups.IsKnown() || securityGroups.IsNull() {
		return false
	}

	return securityGroupsLen > 0
}

func expandAllowedAddressPairs(raw []interface{}, defaultMAC string) []dto.NetworkInterfaceAddressPair {
	pairs := make([]dto.NetworkInterfaceAddressPair, 0, len(raw))
	for _, r := range raw {
		m := r.(map[string]interface{})
		mac, _ := m["mac_address"].(string)
		if mac == "" {
			mac = defaultMAC
		}
		pairs = append(pairs, dto.NetworkInterfaceAddressPair{
			IPAddress:  m["ip_address"].(string),
			MACAddress: mac,
		})
	}
	return pairs
}

func flattenAllowedAddressPairs(pairs []dto.NetworkInterfaceAddressPair) []map[string]interface{} {
	result := make([]map[string]interface{}, 0, len(pairs))
	for _, p := range pairs {
		result = append(result, map[string]interface{}{
			"ip_address":  p.IPAddress,
			"mac_address": p.MACAddress,
		})
	}
	return result
}

func expandStringSet(s *schema.Set) []string {
	out := make([]string, 0, s.Len())
	for _, v := range s.List() {
		out = append(out, v.(string))
	}
	return out
}

func resourceNetworkInterfaceCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	cfg := meta.(*config.Config)

	createOpts := dto.CreateNetworkInterfaceRequest{
		Name:        d.Get("name").(string),
		SubnetID:    d.Get("subnet_id").(string),
		IPAddress:   d.Get("ip_address").(string),
		Description: d.Get("description").(string),
		Reserved:    d.Get("reserved").(bool),
		VirtualIP:   d.Get("virtual_ip").(bool),
	}

	createResp := &dto.NetworkInterfaceResponse{}
	_, err := cfg.Client.Post(ctx, client.ApiPath.NetworkInterfaces(cfg.ProjectID), createOpts, createResp, nil)
	if err != nil {
		return diag.Errorf("Error creating vnpaycloud_network_interface: %s", err)
	}

	d.SetId(createResp.NetworkInterface.ID)

	stateConf := &retry.StateChangeConf{
		Pending:    []string{"initiating", "creating", "build"},
		Target:     []string{"active", "created"},
		Refresh:    networkInterfaceStateRefreshFunc(ctx, cfg.Client, cfg.ProjectID, createResp.NetworkInterface.ID),
		Timeout:    d.Timeout(schema.TimeoutCreate),
		Delay:      5 * time.Second,
		MinTimeout: 3 * time.Second,
	}

	_, err = stateConf.WaitForStateContext(ctx)
	if err != nil {
		return diag.Errorf("Error waiting for vnpaycloud_network_interface %s to become ready: %s", createResp.NetworkInterface.ID, err)
	}

	current := &dto.NetworkInterfaceResponse{}
	if _, err := cfg.Client.Get(ctx, client.ApiPath.NetworkInterfaceWithID(cfg.ProjectID, d.Id()), current, nil); err != nil {
		return diag.Errorf("Error reading vnpaycloud_network_interface %s after create: %s", d.Id(), err)
	}

	if raw := d.Get("allowed_address_pairs").([]interface{}); len(raw) > 0 {
		pairsReq := dto.UpdateNetworkInterfaceAllowedAddressPairsRequest{
			AllowedAddressPairs: expandAllowedAddressPairs(raw, current.NetworkInterface.MACAddress),
		}
		if _, err := cfg.Client.Put(ctx, client.ApiPath.NetworkInterfaceAllowedAddressPairs(cfg.ProjectID, d.Id()), pairsReq, nil, nil); err != nil {
			return diag.Errorf("Error setting allowed address pairs for vnpaycloud_network_interface %s: %s", d.Id(), err)
		}
	}

	if raw := d.GetRawConfig(); !raw.IsNull() {
		if psAttr := raw.GetAttr("port_security_enabled"); !psAttr.IsNull() {
			if desired := d.Get("port_security_enabled").(bool); desired != current.NetworkInterface.PortSecurityEnabled {
				req := dto.UpdateNetworkInterfacePortSecurityRequest{PortSecurityEnabled: desired}
				if _, err := cfg.Client.Put(ctx, client.ApiPath.NetworkInterfacePortSecurity(cfg.ProjectID, d.Id()), req, nil, nil); err != nil {
					return diag.Errorf("Error setting port security for vnpaycloud_network_interface %s: %s", d.Id(), err)
				}
			}
		}
	}

	if sgs := d.Get("security_groups").(*schema.Set); sgs.Len() > 0 {
		req := dto.UpdateNetworkInterfaceSecurityGroupsRequest{SecurityGroupIDs: expandStringSet(sgs)}
		if _, err := cfg.Client.Put(ctx, client.ApiPath.NetworkInterfaceSecurityGroups(cfg.ProjectID, d.Id()), req, nil, nil); err != nil {
			return diag.Errorf("Error setting security groups for vnpaycloud_network_interface %s: %s", d.Id(), err)
		}
	}

	return resourceNetworkInterfaceRead(ctx, d, meta)
}

func resourceNetworkInterfaceRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	cfg := meta.(*config.Config)

	niResp := &dto.NetworkInterfaceResponse{}
	_, err := cfg.Client.Get(ctx, client.ApiPath.NetworkInterfaceWithID(cfg.ProjectID, d.Id()), niResp, nil)
	if err != nil {
		return diag.FromErr(util.CheckNotFound(d, err, "Error retrieving vnpaycloud_network_interface"))
	}

	ni := niResp.NetworkInterface

	d.Set("name", ni.Name)
	d.Set("network_id", ni.NetworkID)
	d.Set("subnet_id", ni.SubnetID)
	d.Set("ip_address", ni.IPAddress)
	d.Set("mac_address", ni.MACAddress)
	d.Set("status", ni.Status)
	d.Set("security_groups", ni.SecurityGroups)
	d.Set("port_security_enabled", ni.PortSecurityEnabled)
	d.Set("network_type", ni.NetworkType)
	d.Set("description", ni.Description)
	d.Set("reserved", ni.Reserved)
	d.Set("virtual_ip", ni.VirtualIP)
	d.Set("allowed_address_pairs", flattenAllowedAddressPairs(ni.AllowedAddressPairs))
	d.Set("created_at", ni.CreatedAt)

	return nil
}

func resourceNetworkInterfaceUpdate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	cfg := meta.(*config.Config)

	if d.HasChange("reserved") || d.HasChange("description") {
		req := dto.UpdateNetworkInterfaceReservedRequest{
			Reserved:    d.Get("reserved").(bool),
			Description: d.Get("description").(string),
		}
		if _, err := cfg.Client.Put(ctx, client.ApiPath.NetworkInterfaceReserved(cfg.ProjectID, d.Id()), req, nil, nil); err != nil {
			return diag.Errorf("Error updating reserved status for vnpaycloud_network_interface %s: %s", d.Id(), err)
		}
	}

	if d.HasChange("virtual_ip") {
		req := dto.UpdateNetworkInterfaceVirtualIpRequest{
			VirtualIP: d.Get("virtual_ip").(bool),
		}
		if _, err := cfg.Client.Put(ctx, client.ApiPath.NetworkInterfaceVirtualIP(cfg.ProjectID, d.Id()), req, nil, nil); err != nil {
			return diag.Errorf("Error updating virtual IP status for vnpaycloud_network_interface %s: %s", d.Id(), err)
		}
	}

	if d.HasChange("allowed_address_pairs") {
		pairsReq := dto.UpdateNetworkInterfaceAllowedAddressPairsRequest{
			AllowedAddressPairs: expandAllowedAddressPairs(d.Get("allowed_address_pairs").([]interface{}), d.Get("mac_address").(string)),
		}
		if _, err := cfg.Client.Put(ctx, client.ApiPath.NetworkInterfaceAllowedAddressPairs(cfg.ProjectID, d.Id()), pairsReq, nil, nil); err != nil {
			return diag.Errorf("Error updating allowed address pairs for vnpaycloud_network_interface %s: %s", d.Id(), err)
		}
	}

	if d.HasChange("port_security_enabled") {
		req := dto.UpdateNetworkInterfacePortSecurityRequest{PortSecurityEnabled: d.Get("port_security_enabled").(bool)}
		if _, err := cfg.Client.Put(ctx, client.ApiPath.NetworkInterfacePortSecurity(cfg.ProjectID, d.Id()), req, nil, nil); err != nil {
			return diag.Errorf("Error updating port security for vnpaycloud_network_interface %s: %s", d.Id(), err)
		}
	}

	if d.HasChange("security_groups") {
		req := dto.UpdateNetworkInterfaceSecurityGroupsRequest{SecurityGroupIDs: expandStringSet(d.Get("security_groups").(*schema.Set))}
		if _, err := cfg.Client.Put(ctx, client.ApiPath.NetworkInterfaceSecurityGroups(cfg.ProjectID, d.Id()), req, nil, nil); err != nil {
			return diag.Errorf("Error updating security groups for vnpaycloud_network_interface %s: %s", d.Id(), err)
		}
	}

	return resourceNetworkInterfaceRead(ctx, d, meta)
}

func resourceNetworkInterfaceDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	cfg := meta.(*config.Config)

	niResp := &dto.NetworkInterfaceResponse{}
	_, err := cfg.Client.Get(ctx, client.ApiPath.NetworkInterfaceWithID(cfg.ProjectID, d.Id()), niResp, nil)
	if err != nil {
		return diag.FromErr(util.CheckDeleted(d, err, "Error retrieving vnpaycloud_network_interface"))
	}

	if niResp.NetworkInterface.Status != "deleting" {
		if _, err := cfg.Client.Delete(ctx, client.ApiPath.NetworkInterfaceWithID(cfg.ProjectID, d.Id()), nil); err != nil {
			return diag.FromErr(util.CheckDeleted(d, err, "Error deleting vnpaycloud_network_interface"))
		}
	}

	stateConf := &retry.StateChangeConf{
		Pending:    []string{"deleting", "active", "created"},
		Target:     []string{"deleted"},
		Refresh:    networkInterfaceStateRefreshFunc(ctx, cfg.Client, cfg.ProjectID, d.Id()),
		Timeout:    d.Timeout(schema.TimeoutDelete),
		Delay:      5 * time.Second,
		MinTimeout: 3 * time.Second,
	}

	_, err = stateConf.WaitForStateContext(ctx)
	if err != nil {
		return diag.Errorf("Error waiting for vnpaycloud_network_interface %s to delete: %s", d.Id(), err)
	}

	return nil
}
