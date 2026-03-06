package networkinterface

import (
	"context"
	"terraform-provider-vnpaycloud/vnpaycloud/config"
	"terraform-provider-vnpaycloud/vnpaycloud/dto"
	"terraform-provider-vnpaycloud/vnpaycloud/helper/client"
	"terraform-provider-vnpaycloud/vnpaycloud/util"
	"time"

	"github.com/hashicorp/terraform-plugin-log/tflog"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/retry"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

func ResourceNetworkInterface() *schema.Resource {
	return &schema.Resource{
		CreateContext: resourceNetworkInterfaceCreate,
		ReadContext:   resourceNetworkInterfaceRead,
		DeleteContext: resourceNetworkInterfaceDelete,
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
				ForceNew: true,
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
				Type:     schema.TypeList,
				Computed: true,
				Elem:     &schema.Schema{Type: schema.TypeString},
			},
			"port_security_enabled": {
				Type:     schema.TypeBool,
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

func resourceNetworkInterfaceCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	cfg := meta.(*config.Config)

	createOpts := dto.CreateNetworkInterfaceRequest{
		Name:        d.Get("name").(string),
		SubnetID:    d.Get("subnet_id").(string),
		IPAddress:   d.Get("ip_address").(string),
		Description: d.Get("description").(string),
	}

	tflog.Debug(ctx, "vnpaycloud_network_interface create options", map[string]interface{}{"create_opts": createOpts})

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
	tflog.Debug(ctx, "Retrieved vnpaycloud_network_interface "+d.Id(), map[string]interface{}{"network_interface": ni})

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

	return nil
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
