package networkinterface

import (
	"context"
	"terraform-provider-vnpaycloud/vnpaycloud/config"
	"terraform-provider-vnpaycloud/vnpaycloud/dto"
	"terraform-provider-vnpaycloud/vnpaycloud/helper/client"

	"github.com/hashicorp/terraform-plugin-log/tflog"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

func DataSourceNetworkInterface() *schema.Resource {
	return &schema.Resource{
		ReadContext: dataSourceNetworkInterfaceRead,
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
			"network_id": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"subnet_id": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"ip_address": {
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
				Type:     schema.TypeList,
				Computed: true,
				Elem:     &schema.Schema{Type: schema.TypeString},
			},
			"port_security_enabled": {
				Type:     schema.TypeBool,
				Computed: true,
			},
			"allowed_address_pairs": {
				Type:     schema.TypeList,
				Computed: true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"ip_address": {
							Type:     schema.TypeString,
							Computed: true,
						},
						"mac_address": {
							Type:     schema.TypeString,
							Computed: true,
						},
					},
				},
			},
			"network_type": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"description": {
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

func dataSourceNetworkInterfaceRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	cfg := meta.(*config.Config)

	// If ID is provided, fetch directly
	if id, ok := d.GetOk("id"); ok && id.(string) != "" {
		niResp := &dto.NetworkInterfaceResponse{}
		_, err := cfg.Client.Get(ctx, client.ApiPath.NetworkInterfaceWithID(cfg.ProjectID, id.(string)), niResp, nil)
		if err != nil {
			return diag.Errorf("Error retrieving vnpaycloud_network_interface %s: %s", id, err)
		}
		tflog.Debug(ctx, "Retrieved vnpaycloud_network_interface datasource", map[string]interface{}{"network_interface": niResp.NetworkInterface})
		setNetworkInterfaceDataSourceAttributes(d, niResp.NetworkInterface)
		return nil
	}

	// Otherwise, list and filter by name
	listResp := &dto.ListNetworkInterfacesResponse{}
	_, err := cfg.Client.Get(ctx, client.ApiPath.NetworkInterfaces(cfg.ProjectID), listResp, nil)
	if err != nil {
		return diag.Errorf("Unable to query vnpaycloud_network_interface: %s", err)
	}

	name := d.Get("name").(string)
	var matched []dto.NetworkInterface
	for _, ni := range listResp.NetworkInterfaces {
		if name != "" && ni.Name != name {
			continue
		}
		matched = append(matched, ni)
	}

	if len(matched) < 1 {
		return diag.Errorf("Your vnpaycloud_network_interface query returned no results")
	}

	if len(matched) > 1 {
		return diag.Errorf("Your vnpaycloud_network_interface query returned multiple results")
	}

	tflog.Debug(ctx, "Retrieved vnpaycloud_network_interface datasource", map[string]interface{}{"network_interface": matched[0]})
	setNetworkInterfaceDataSourceAttributes(d, matched[0])

	return nil
}

func setNetworkInterfaceDataSourceAttributes(d *schema.ResourceData, ni dto.NetworkInterface) {
	d.SetId(ni.ID)
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
	d.Set("created_at", ni.CreatedAt)

	allowedAddressPairs := make([]map[string]interface{}, len(ni.AllowedAddressPairs))
	for i, pair := range ni.AllowedAddressPairs {
		allowedAddressPairs[i] = map[string]interface{}{
			"ip_address":  pair.IPAddress,
			"mac_address": pair.MACAddress,
		}
	}
	d.Set("allowed_address_pairs", allowedAddressPairs)
}
