package port

import (
	"context"
	"encoding/json"
	"log"
	"terraform-provider-vnpaycloud/vnpaycloud/config"
	"terraform-provider-vnpaycloud/vnpaycloud/dto"
	"terraform-provider-vnpaycloud/vnpaycloud/helper/client"
	"terraform-provider-vnpaycloud/vnpaycloud/util"
	"time"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/retry"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/structure"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
)

func ResourceNetworkingPort() *schema.Resource {
	return &schema.Resource{
		CreateContext: resourceNetworkingPortCreate,
		ReadContext:   resourceNetworkingPortRead,
		UpdateContext: resourceNetworkingPortUpdate,
		DeleteContext: resourceNetworkingPortDelete,
		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},

		Timeouts: &schema.ResourceTimeout{
			Create: schema.DefaultTimeout(10 * time.Minute),
			Delete: schema.DefaultTimeout(10 * time.Minute),
		},

		Schema: map[string]*schema.Schema{
			"region": {
				Type:     schema.TypeString,
				Optional: true,
				Computed: true,
				ForceNew: true,
			},

			"name": {
				Type:     schema.TypeString,
				Optional: true,
				ForceNew: false,
			},

			"description": {
				Type:     schema.TypeString,
				Optional: true,
			},

			"network_id": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},

			"admin_state_up": {
				Type:     schema.TypeBool,
				Optional: true,
				ForceNew: false,
				Computed: true,
			},

			"mac_address": {
				Type:     schema.TypeString,
				Optional: true,
				ForceNew: true,
				Computed: true,
			},

			"tenant_id": {
				Type:     schema.TypeString,
				Optional: true,
				ForceNew: true,
				Computed: true,
			},

			"device_owner": {
				Type:     schema.TypeString,
				Optional: true,
				ForceNew: true,
				Computed: true,
			},

			"security_group_ids": {
				Type:     schema.TypeSet,
				Optional: true,
				ForceNew: false,
				Elem:     &schema.Schema{Type: schema.TypeString},
				Set:      schema.HashString,
			},

			"no_security_groups": {
				Type:     schema.TypeBool,
				Optional: true,
				ForceNew: false,
			},

			"device_id": {
				Type:     schema.TypeString,
				Optional: true,
				ForceNew: true,
				Computed: true,
			},

			"fixed_ip": {
				Type:          schema.TypeList,
				Optional:      true,
				ForceNew:      false,
				ConflictsWith: []string{"no_fixed_ip"},
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"subnet_id": {
							Type:     schema.TypeString,
							Optional: true,
						},
						"ip_address": {
							Type:     schema.TypeString,
							Optional: true,
						},
					},
				},
			},

			"no_fixed_ip": {
				Type:          schema.TypeBool,
				Optional:      true,
				ForceNew:      false,
				ConflictsWith: []string{"fixed_ip"},
			},

			"allowed_address_pairs": {
				Type:     schema.TypeSet,
				Optional: true,
				ForceNew: false,
				Set:      resourceNetworkingPortAllowedAddressPairsHash,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"ip_address": {
							Type:     schema.TypeString,
							Required: true,
						},
						"mac_address": {
							Type:     schema.TypeString,
							Optional: true,
						},
					},
				},
			},

			"extra_dhcp_option": {
				Type:     schema.TypeSet,
				Optional: true,
				ForceNew: false,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"name": {
							Type:     schema.TypeString,
							Required: true,
						},
						"value": {
							Type:     schema.TypeString,
							Required: true,
						},
						"ip_version": {
							Type:     schema.TypeInt,
							Default:  4,
							Optional: true,
						},
					},
				},
			},

			"value_specs": {
				Type:     schema.TypeMap,
				Optional: true,
				ForceNew: true,
			},

			"all_fixed_ips": {
				Type:     schema.TypeList,
				Computed: true,
				Elem:     &schema.Schema{Type: schema.TypeString},
			},

			"all_security_group_ids": {
				Type:     schema.TypeSet,
				Computed: true,
				Elem:     &schema.Schema{Type: schema.TypeString},
				Set:      schema.HashString,
			},

			"tags": {
				Type:     schema.TypeSet,
				Optional: true,
				Elem:     &schema.Schema{Type: schema.TypeString},
			},

			"all_tags": {
				Type:     schema.TypeSet,
				Computed: true,
				Elem:     &schema.Schema{Type: schema.TypeString},
			},

			"port_security_enabled": {
				Type:     schema.TypeBool,
				Optional: true,
				Computed: true,
			},

			"binding": {
				Type:     schema.TypeList,
				Optional: true,
				Computed: true,
				MaxItems: 1,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"host_id": {
							Type:     schema.TypeString,
							Optional: true,
							Computed: true,
						},
						"profile": {
							Type:             schema.TypeString,
							Optional:         true,
							ValidateFunc:     util.ValidateJSONObject,
							DiffSuppressFunc: util.DiffSuppressJSONObject,
							StateFunc: func(v interface{}) string {
								json, _ := structure.NormalizeJsonString(v)
								return json
							},
						},
						"vif_details": {
							Type:     schema.TypeMap,
							Computed: true,
						},
						"vif_type": {
							Type:     schema.TypeString,
							Computed: true,
						},
						"vnic_type": {
							Type:     schema.TypeString,
							Optional: true,
							Default:  "normal",
							ValidateFunc: validation.StringInSlice([]string{
								"direct", "direct-physical", "macvtap", "normal", "baremetal", "virtio-forwarder",
							}, true),
						},
					},
				},
			},

			"dns_name": {
				Type:     schema.TypeString,
				Optional: true,
				Computed: true,
			},

			"dns_assignment": {
				Type:     schema.TypeList,
				Computed: true,
				Elem:     &schema.Schema{Type: schema.TypeMap},
			},

			"qos_policy_id": {
				Type:     schema.TypeString,
				Optional: true,
				ForceNew: false,
				Computed: true,
			},

			"virtual_ip": {
				Type:     schema.TypeBool,
				Optional: true,
				ForceNew: false,
			},
		},
	}
}

func resourceNetworkingPortCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	config := meta.(*config.Config)
	networkingClient, err := client.NewClient(ctx, config.ConsoleClientConfig)
	if err != nil {
		return diag.Errorf("Error creating VNPAYCLOUD networking client: %s", err)
	}

	securityGroups := util.ExpandToStringSlice(d.Get("security_group_ids").(*schema.Set).List())
	noSecurityGroups := d.Get("no_security_groups").(bool)

	// Check and make sure an invalid security group configuration wasn't given.
	if noSecurityGroups && len(securityGroups) > 0 {
		return diag.Errorf("Cannot have both no_security_groups and security_group_ids set for vnpaycloud_networking_port")
	}

	allowedAddressPairs := d.Get("allowed_address_pairs").(*schema.Set)
	createReq := dto.CreatePortRequest{
		Port: dto.CreatePortOpts{
			Name:                d.Get("name").(string),
			Description:         d.Get("description").(string),
			NetworkID:           d.Get("network_id").(string),
			MACAddress:          d.Get("mac_address").(string),
			TenantID:            d.Get("tenant_id").(string),
			DeviceOwner:         d.Get("device_owner").(string),
			DeviceID:            d.Get("device_id").(string),
			FixedIPs:            expandNetworkingPortFixedIP(d),
			AllowedAddressPairs: expandNetworkingPortAllowedAddressPairs(allowedAddressPairs),
			ValueSpecs:          util.MapValueSpecs(d),
		},
	}

	if v, ok := util.GetOkExists(d, "admin_state_up"); ok {
		asu := v.(bool)
		createReq.Port.AdminStateUp = &asu
	}

	if v, ok := util.GetOkExists(d, "virtual_ip"); ok {
		asu := v.(bool)
		createReq.Port.VirtualIp = &asu
	}

	if noSecurityGroups {
		securityGroups = []string{}
		createReq.Port.SecurityGroups = &securityGroups
	}

	// Only set SecurityGroups if one was specified.
	// Otherwise this would mimic the no_security_groups action.
	if len(securityGroups) > 0 {
		createReq.Port.SecurityGroups = &securityGroups
	}

	dhcpOpts := d.Get("extra_dhcp_option").(*schema.Set)
	if dhcpOpts.Len() > 0 {
		createReq.Port.ExtraDHCPOpts = expandNetworkingPortDHCPOptsCreate(dhcpOpts)
	}

	// Add the port security attribute if specified.
	if v, ok := util.GetOkExists(d, "port_security_enabled"); ok {
		portSecurityEnabled := v.(bool)
		createReq.Port.PortSecurityEnabled = &portSecurityEnabled
	}

	// Add the port binding parameters if specified.
	if v, ok := util.GetOkExists(d, "binding"); ok {
		for _, raw := range v.([]interface{}) {
			binding := raw.(map[string]interface{})
			var profile map[string]interface{}

			// Convert raw string into the map
			rawProfile := binding["profile"].(string)
			if len(rawProfile) > 0 {
				err := json.Unmarshal([]byte(rawProfile), &profile)
				if err != nil {
					return diag.Errorf("Failed to unmarshal the JSON: %s", err)
				}
			}

			createReq.Port.HostID = binding["host_id"].(string)
			createReq.Port.Profile = profile
			createReq.Port.VNICType = binding["vnic_type"].(string)
		}
	}

	if dnsName := d.Get("dns_name").(string); dnsName != "" {
		createReq.Port.DNSName = dnsName
	}

	if qosPolicyID := d.Get("qos_policy_id").(string); qosPolicyID != "" {
		createReq.Port.QoSPolicyID = qosPolicyID
	}

	log.Printf("[DEBUG] vnpaycloud_networking_port create options: %#v", createReq)

	// Create a Neutron port and set extra options if they're specified.
	createResp := dto.CreatePortResponse{}

	_, err = networkingClient.Post(ctx, client.ApiPath.Port, createReq, &createResp, nil)
	if err != nil {
		return diag.Errorf("Error creating vnpaycloud_networking_port: %s", err)
	}

	port := createResp.Port

	log.Printf("[DEBUG] Waiting for vnpaycloud_networking_port %s to become available.", port.ID)

	stateConf := &retry.StateChangeConf{
		Target:     []string{"ACTIVE", "DOWN"},
		Refresh:    resourceNetworkingPortStateRefreshFunc(ctx, networkingClient, port.ID),
		Timeout:    d.Timeout(schema.TimeoutCreate),
		Delay:      5 * time.Second,
		MinTimeout: 3 * time.Second,
	}

	_, err = stateConf.WaitForStateContext(ctx)
	if err != nil {
		return diag.Errorf("Error waiting for vnpaycloud_networking_port %s to become available: %s", port.ID, err)
	}

	d.SetId(port.ID)

	// tags := util.NetworkingAttributesTags(d)
	// if len(tags) > 0 {
	// 	tagOpts := attributestags.ReplaceAllOpts{Tags: tags}
	// 	tags, err := attributestags.ReplaceAll(ctx, networkingClient, "ports", port.ID, tagOpts).Extract()
	// 	if err != nil {
	// 		return diag.Errorf("Error setting tags on vnpaycloud_networking_port %s: %s", port.ID, err)
	// 	}
	// 	log.Printf("[DEBUG] Set tags %s on vnpaycloud_networking_port %s", tags, port.ID)
	// }

	log.Printf("[DEBUG] Created vnpaycloud_networking_port %s: %#v", port.ID, port)
	return resourceNetworkingPortRead(ctx, d, meta)
}

func resourceNetworkingPortRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	config := meta.(*config.Config)
	networkingClient, err := client.NewClient(ctx, config.ConsoleClientConfig)
	if err != nil {
		return diag.Errorf("Error creating VNPAYCLOUD networking client: %s", err)
	}

	getResp := dto.GetPortResponse{}
	_, err = networkingClient.Get(ctx, client.ApiPath.PortWithId(d.Id()), &getResp, nil)
	if err != nil {
		return diag.FromErr(util.CheckDeleted(d, err, "Error getting vnpaycloud_networking_port"))
	}

	port := getResp.Port

	log.Printf("[DEBUG] Retrieved vnpaycloud_networking_port %s: %#v", d.Id(), port)

	d.Set("name", port.Name)
	d.Set("description", port.Description)
	d.Set("admin_state_up", port.AdminStateUp)
	d.Set("network_id", port.NetworkID)
	d.Set("mac_address", port.MACAddress)
	d.Set("tenant_id", port.TenantID)
	d.Set("device_owner", port.DeviceOwner)
	d.Set("device_id", port.DeviceID)

	util.NetworkingReadAttributesTags(d, port.Tags)

	// Set a slice of all returned Fixed IPs.
	// This will be in the order returned by the API,
	// which is usually alpha-numeric.
	d.Set("all_fixed_ips", expandNetworkingPortFixedIPToStringSlice(port.FixedIPs))

	// Set all security groups.
	// This can be different from what the user specified since
	// the port can have the "default" group automatically applied.
	d.Set("all_security_group_ids", port.SecurityGroups)

	d.Set("allowed_address_pairs", flattenNetworkingPortAllowedAddressPairs(port.MACAddress, port.AllowedAddressPairs))
	d.Set("extra_dhcp_option", flattenNetworkingPortDHCPOpts(port.ExtraDHCPOptsExt))
	d.Set("port_security_enabled", port.PortSecurityEnabled)
	d.Set("binding", flattenNetworkingPortBinding(port))
	d.Set("dns_name", port.DNSName)
	d.Set("dns_assignment", port.DNSAssignment)
	d.Set("qos_policy_id", port.QoSPolicyID)

	d.Set("region", util.GetRegion(d, config))

	return nil
}

func resourceNetworkingPortUpdate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	config := meta.(*config.Config)
	networkingClient, err := client.NewClient(ctx, config.ConsoleClientConfig)
	if err != nil {
		return diag.Errorf("Error creating VNPAYCLOUD networking client: %s", err)
	}

	securityGroups := util.ExpandToStringSlice(d.Get("security_group_ids").(*schema.Set).List())
	noSecurityGroups := d.Get("no_security_groups").(bool)

	// Check and make sure an invalid security group configuration wasn't given.
	if noSecurityGroups && len(securityGroups) > 0 {
		return diag.Errorf("Cannot have both no_security_groups and security_group_ids set for vnpaycloud_networking_port")
	}

	var hasChange bool
	var updateOpts dto.UpdatePortOpts
	updateReq := dto.UpdatePortRequest{
		UpdatePortOpts: updateOpts,
	}

	if d.HasChange("allowed_address_pairs") {
		hasChange = true
		allowedAddressPairs := d.Get("allowed_address_pairs").(*schema.Set)
		aap := expandNetworkingPortAllowedAddressPairs(allowedAddressPairs)
		updateOpts.AllowedAddressPairs = &aap
	}

	if d.HasChange("no_security_groups") {
		if noSecurityGroups {
			hasChange = true
			v := []string{}
			updateOpts.SecurityGroups = &v
		}
	}

	if d.HasChange("security_group_ids") {
		hasChange = true
		updateOpts.SecurityGroups = &securityGroups
	}

	if d.HasChange("name") {
		hasChange = true
		name := d.Get("name").(string)
		updateOpts.Name = &name
	}

	if d.HasChange("description") {
		hasChange = true
		description := d.Get("description").(string)
		updateOpts.Description = &description
	}

	if d.HasChange("admin_state_up") {
		hasChange = true
		asu := d.Get("admin_state_up").(bool)
		updateOpts.AdminStateUp = &asu
	}

	if d.HasChange("virtual_ip") {
		hasChange = true
		asu := d.Get("virtual_ip").(bool)
		updateOpts.VirtualIp = &asu
	}

	if d.HasChange("device_owner") {
		hasChange = true
		deviceOwner := d.Get("device_owner").(string)
		updateOpts.DeviceOwner = &deviceOwner
	}

	if d.HasChange("device_id") {
		hasChange = true
		deviceID := d.Get("device_id").(string)
		updateOpts.DeviceID = &deviceID
	}

	if d.HasChange("fixed_ip") || d.HasChange("no_fixed_ip") {
		fixedIPs := expandNetworkingPortFixedIP(d)
		if fixedIPs != nil {
			hasChange = true
			updateOpts.FixedIPs = fixedIPs
		}
	}

	if d.HasChange("port_security_enabled") {
		hasChange = true
		portSecurityEnabled := d.Get("port_security_enabled").(bool)
		updateReq.PortSecurityEnabled = &portSecurityEnabled
	}

	// Next, perform any dhcp option changes.
	if d.HasChange("extra_dhcp_option") {
		hasChange = true

		o, n := d.GetChange("extra_dhcp_option")
		oldDHCPOpts := o.(*schema.Set)
		newDHCPOpts := n.(*schema.Set)

		deleteDHCPOpts := oldDHCPOpts.Difference(newDHCPOpts)
		addDHCPOpts := newDHCPOpts.Difference(oldDHCPOpts)

		updateExtraDHCPOpts := expandNetworkingPortDHCPOptsUpdate(deleteDHCPOpts, addDHCPOpts)
		updateReq.ExtraDHCPOpts = updateExtraDHCPOpts
	}

	// Next, perform port binding option changes.
	if d.HasChange("binding") {
		var bindingChanged bool

		profile := map[string]interface{}{}

		for _, raw := range d.Get("binding").([]interface{}) {
			binding := raw.(map[string]interface{})

			if d.HasChange("binding.0.vnic_type") {
				bindingChanged = true
				updateReq.VNICType = binding["vnic_type"].(string)
			}

			if d.HasChange("binding.0.host_id") {
				bindingChanged = true
				hostID := binding["host_id"].(string)
				updateReq.HostID = &hostID
			}

			if d.HasChange("binding.0.profile") {
				bindingChanged = true
				// Convert raw string into the map
				rawProfile := binding["profile"].(string)
				if len(rawProfile) > 0 {
					err := json.Unmarshal([]byte(rawProfile), &profile)
					if err != nil {
						return diag.Errorf("Failed to unmarshal the JSON: %s", err)
					}
					if profile == nil {
						profile = map[string]interface{}{}
					}
				}
				updateReq.Profile = profile
			}
		}

		if bindingChanged {
			hasChange = true
		}
	}

	if d.HasChange("dns_name") {
		hasChange = true

		dnsName := d.Get("dns_name").(string)
		updateReq.DNSName = &dnsName
	}

	if d.HasChange("qos_policy_id") {
		hasChange = true

		qosPolicyID := d.Get("qos_policy_id").(string)
		updateReq.QoSPolicyID = &qosPolicyID
	}

	// At this point, perform the update for all "standard" port changes.
	if hasChange {
		log.Printf("[DEBUG] vnpaycloud_networking_port %s update options: %#v", d.Id(), updateReq)
		_, err = networkingClient.Put(ctx, client.ApiPath.PortWithId(d.Id()), updateReq, nil, nil)
		if err != nil {
			return diag.Errorf("Error updating VNPAYCLOUD Neutron Port: %s", err)
		}
	}

	// Next, perform any required updates to the tags.
	// if d.HasChange("tags") {
	// 	tags := util.NetworkingUpdateAttributesTags(d)
	// 	tagOpts := attributestags.ReplaceAllOpts{Tags: tags}
	// 	tags, err := attributestags.ReplaceAll(ctx, networkingClient, "ports", d.Id(), tagOpts).Extract()
	// 	if err != nil {
	// 		return diag.Errorf("Error setting tags on vnpaycloud_networking_port %s: %s", d.Id(), err)
	// 	}
	// 	log.Printf("[DEBUG] Set tags %s on vnpaycloud_networking_port %s", tags, d.Id())
	// }

	return resourceNetworkingPortRead(ctx, d, meta)
}

func resourceNetworkingPortDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	config := meta.(*config.Config)
	networkingClient, err := client.NewClient(ctx, config.ConsoleClientConfig)
	if err != nil {
		return diag.Errorf("Error creating VNPAYCLOUD networking client: %s", err)
	}

	_, err = networkingClient.Delete(ctx, client.ApiPath.PortWithId(d.Id()), nil)
	if err != nil {
		return diag.FromErr(util.CheckDeleted(d, err, "Error deleting vnpaycloud_networking_port"))
	}

	stateConf := &retry.StateChangeConf{
		Pending:    []string{"ACTIVE"},
		Target:     []string{"DELETED"},
		Refresh:    resourceNetworkingPortStateRefreshFunc(ctx, networkingClient, d.Id()),
		Timeout:    d.Timeout(schema.TimeoutDelete),
		Delay:      5 * time.Second,
		MinTimeout: 3 * time.Second,
	}

	_, err = stateConf.WaitForStateContext(ctx)
	if err != nil {
		return diag.Errorf("Error waiting for vnpaycloud_networking_port %s to Delete:  %s", d.Id(), err)
	}

	d.SetId("")
	return nil
}
