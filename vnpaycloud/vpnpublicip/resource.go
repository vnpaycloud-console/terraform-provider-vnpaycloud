package vpnpublicip

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

func ResourceVPNPublicIP() *schema.Resource {
	return &schema.Resource{
		CreateContext: resourceVPNPublicIPCreate,
		ReadContext:   resourceVPNPublicIPRead,
		UpdateContext: resourceVPNPublicIPUpdate,
		DeleteContext: resourceVPNPublicIPDelete,
		Description:   "Manages a VNPAY Cloud VPN public IP.",
		Importer: &schema.ResourceImporter{
			StateContext: func(ctx context.Context, d *schema.ResourceData, m interface{}) ([]*schema.ResourceData, error) {
				cfg := m.(*config.Config)
				resp := &dto.VPNPublicIPResponse{}
				if _, err := cfg.Client.Get(ctx, client.ApiPath.VPNPublicIPWithID(cfg.ProjectID, d.Id()), resp, nil); err != nil {
					return nil, fmt.Errorf("vnpaycloud_vpn_public_ip %q not found: %w", d.Id(), err)
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
		},
	}
}

func resourceVPNPublicIPCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	cfg := meta.(*config.Config)

	createOpts := dto.CreateVPNPublicIPRequest{
		Name:        d.Get("name").(string),
		Description: d.Get("description").(string),
	}

	tflog.Debug(ctx, "vnpaycloud_vpn_public_ip create options", map[string]interface{}{"create_opts": createOpts})

	createResp := &dto.VPNPublicIPResponse{}
	_, err := cfg.Client.Post(ctx, client.ApiPath.VPNPublicIPs(cfg.ProjectID), createOpts, createResp, nil)
	if err != nil {
		return diag.Errorf("Error creating vnpaycloud_vpn_public_ip: %s", err)
	}

	d.SetId(createResp.VPNPublicIP.ID)

	stateConf := &retry.StateChangeConf{
		Pending:    []string{"initiating", "creating"},
		Target:     []string{"active"},
		Refresh:    vpnPublicIPStateRefreshFunc(ctx, cfg.Client, cfg.ProjectID, createResp.VPNPublicIP.ID),
		Timeout:    d.Timeout(schema.TimeoutCreate),
		Delay:      5 * time.Second,
		MinTimeout: 3 * time.Second,
	}

	_, err = stateConf.WaitForStateContext(ctx)
	if err != nil {
		return diag.Errorf("Error waiting for vnpaycloud_vpn_public_ip %s to become ready: %s", createResp.VPNPublicIP.ID, err)
	}

	return resourceVPNPublicIPRead(ctx, d, meta)
}

func resourceVPNPublicIPRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	cfg := meta.(*config.Config)

	resp := &dto.VPNPublicIPResponse{}
	_, err := cfg.Client.Get(ctx, client.ApiPath.VPNPublicIPWithID(cfg.ProjectID, d.Id()), resp, nil)
	if err != nil {
		return diag.FromErr(util.CheckNotFound(d, err, "Error retrieving vnpaycloud_vpn_public_ip"))
	}

	tflog.Debug(ctx, "Retrieved vnpaycloud_vpn_public_ip "+d.Id(), map[string]interface{}{"vpn_public_ip": resp.VPNPublicIP})

	d.Set("name", resp.VPNPublicIP.Name)
	d.Set("description", resp.VPNPublicIP.Description)
	d.Set("address", resp.VPNPublicIP.FloatingIP)
	d.Set("status", util.NormalizeStatus(resp.VPNPublicIP.Status))
	d.Set("created_at", resp.VPNPublicIP.CreatedAt)

	return nil
}

func resourceVPNPublicIPUpdate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	cfg := meta.(*config.Config)

	updateOpts := dto.UpdateVPNPublicIPRequest{
		Name:        d.Get("name").(string),
		Description: d.Get("description").(string),
	}

	tflog.Debug(ctx, "vnpaycloud_vpn_public_ip update options", map[string]interface{}{"update_opts": updateOpts})

	updateResp := &dto.VPNPublicIPResponse{}
	_, err := cfg.Client.Put(ctx, client.ApiPath.VPNPublicIPWithID(cfg.ProjectID, d.Id()), updateOpts, updateResp, nil)
	if err != nil {
		return diag.Errorf("Error updating vnpaycloud_vpn_public_ip %s: %s", d.Id(), err)
	}

	return resourceVPNPublicIPRead(ctx, d, meta)
}

func resourceVPNPublicIPDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	cfg := meta.(*config.Config)

	resp := &dto.VPNPublicIPResponse{}
	_, err := cfg.Client.Get(ctx, client.ApiPath.VPNPublicIPWithID(cfg.ProjectID, d.Id()), resp, nil)
	if err != nil {
		return diag.FromErr(util.CheckDeleted(d, err, "Error retrieving vnpaycloud_vpn_public_ip"))
	}

	if resp.VPNPublicIP.Status != "deleting" {
		if _, err := cfg.Client.Delete(ctx, client.ApiPath.VPNPublicIPWithID(cfg.ProjectID, d.Id()), nil); err != nil {
			return diag.FromErr(util.CheckDeleted(d, err, "Error deleting vnpaycloud_vpn_public_ip"))
		}
	}

	stateConf := &retry.StateChangeConf{
		Pending:    []string{"deleting", "active"},
		Target:     []string{"deleted"},
		Refresh:    vpnPublicIPStateRefreshFunc(ctx, cfg.Client, cfg.ProjectID, d.Id()),
		Timeout:    d.Timeout(schema.TimeoutDelete),
		Delay:      5 * time.Second,
		MinTimeout: 3 * time.Second,
	}

	_, err = stateConf.WaitForStateContext(ctx)
	if err != nil {
		return diag.Errorf("Error waiting for vnpaycloud_vpn_public_ip %s to delete: %s", d.Id(), err)
	}

	return nil
}
