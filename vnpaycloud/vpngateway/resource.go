package vpngateway

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

func ResourceVPNGateway() *schema.Resource {
	return &schema.Resource{
		CreateContext: resourceVPNGatewayCreate,
		ReadContext:   resourceVPNGatewayRead,
		UpdateContext: resourceVPNGatewayUpdate,
		DeleteContext: resourceVPNGatewayDelete,
		Description:   "Manages a VNPAY Cloud VPN gateway.",
		Importer: &schema.ResourceImporter{
			StateContext: func(ctx context.Context, d *schema.ResourceData, m interface{}) ([]*schema.ResourceData, error) {
				cfg := m.(*config.Config)
				resp := &dto.VPNGatewayResponse{}
				if _, err := cfg.Client.Get(ctx, client.ApiPath.VPNGatewayWithID(cfg.ProjectID, d.Id()), resp, nil); err != nil {
					return nil, fmt.Errorf("vnpaycloud_vpn_gateway %q not found: %w", d.Id(), err)
				}

				return []*schema.ResourceData{d}, nil
			},
		},
		Timeouts: &schema.ResourceTimeout{
			Create: schema.DefaultTimeout(40 * time.Minute),
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
			"vpn_type": {
				Type:         schema.TypeString,
				Required:     true,
				ForceNew:     true,
				ValidateFunc: validation.StringInSlice([]string{"POLICY_BASED", "ROUTE_BASED"}, false),
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
		},
	}
}

func resourceVPNGatewayCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	cfg := meta.(*config.Config)

	createOpts := dto.CreateVPNGatewayRequest{
		Name:        d.Get("name").(string),
		Description: d.Get("description").(string),
		VPNType:     d.Get("vpn_type").(string),
	}

	tflog.Debug(ctx, "vnpaycloud_vpn_gateway create options", map[string]interface{}{"create_opts": createOpts})

	createResp := &dto.VPNGatewayResponse{}
	_, err := cfg.Client.Post(ctx, client.ApiPath.VPNGateways(cfg.ProjectID), createOpts, createResp, nil)
	if err != nil {
		return diag.Errorf("Error creating vnpaycloud_vpn_gateway: %s", err)
	}

	d.SetId(createResp.VPNGateway.ID)

	stateConf := &retry.StateChangeConf{
		Pending:    []string{"initiating", "creating"},
		Target:     []string{"active"},
		Refresh:    vpnGatewayStateRefreshFunc(ctx, cfg.Client, cfg.ProjectID, createResp.VPNGateway.ID),
		Timeout:    d.Timeout(schema.TimeoutCreate),
		Delay:      10 * time.Second,
		MinTimeout: 5 * time.Second,
	}

	_, err = stateConf.WaitForStateContext(ctx)
	if err != nil {
		return diag.Errorf("Error waiting for vnpaycloud_vpn_gateway %s to become ready: %s", createResp.VPNGateway.ID, err)
	}

	return resourceVPNGatewayRead(ctx, d, meta)
}

func resourceVPNGatewayRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	cfg := meta.(*config.Config)

	gwResp := &dto.VPNGatewayResponse{}
	_, err := cfg.Client.Get(ctx, client.ApiPath.VPNGatewayWithID(cfg.ProjectID, d.Id()), gwResp, nil)
	if err != nil {
		return diag.FromErr(util.CheckNotFound(d, err, "Error retrieving vnpaycloud_vpn_gateway"))
	}

	tflog.Debug(ctx, "Retrieved vnpaycloud_vpn_gateway "+d.Id(), map[string]interface{}{"vpn_gateway": gwResp.VPNGateway})

	d.Set("name", gwResp.VPNGateway.Name)
	d.Set("description", gwResp.VPNGateway.Description)
	d.Set("vpn_type", gwResp.VPNGateway.VPNType)
	d.Set("status", util.NormalizeStatus(gwResp.VPNGateway.Status))
	d.Set("attached_vpc_ids", gwResp.VPNGateway.AttachedVPCIDs)
	d.Set("created_at", gwResp.VPNGateway.CreatedAt)

	return nil
}

func resourceVPNGatewayUpdate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	cfg := meta.(*config.Config)

	if d.HasChange("name") || d.HasChange("description") {
		updateOpts := dto.UpdateVPNGatewayRequest{
			Name:        d.Get("name").(string),
			Description: d.Get("description").(string),
		}

		tflog.Debug(ctx, "vnpaycloud_vpn_gateway update options", map[string]interface{}{"update_opts": updateOpts})

		_, err := cfg.Client.Put(ctx, client.ApiPath.VPNGatewayWithID(cfg.ProjectID, d.Id()), updateOpts, nil, nil)
		if err != nil {
			return diag.Errorf("Error updating vnpaycloud_vpn_gateway %s: %s", d.Id(), err)
		}
	}

	return resourceVPNGatewayRead(ctx, d, meta)
}

func resourceVPNGatewayDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	cfg := meta.(*config.Config)

	gwResp := &dto.VPNGatewayResponse{}
	_, err := cfg.Client.Get(ctx, client.ApiPath.VPNGatewayWithID(cfg.ProjectID, d.Id()), gwResp, nil)
	if err != nil {
		return diag.FromErr(util.CheckDeleted(d, err, "Error retrieving vnpaycloud_vpn_gateway"))
	}

	if gwResp.VPNGateway.Status != "deleting" {
		if _, err := cfg.Client.Delete(ctx, client.ApiPath.VPNGatewayWithID(cfg.ProjectID, d.Id()), nil); err != nil {
			return diag.FromErr(util.CheckDeleted(d, err, "Error deleting vnpaycloud_vpn_gateway"))
		}
	}

	stateConf := &retry.StateChangeConf{
		Pending:    []string{"deleting", "active"},
		Target:     []string{"deleted"},
		Refresh:    vpnGatewayStateRefreshFunc(ctx, cfg.Client, cfg.ProjectID, d.Id()),
		Timeout:    d.Timeout(schema.TimeoutDelete),
		Delay:      10 * time.Second,
		MinTimeout: 5 * time.Second,
	}

	_, err = stateConf.WaitForStateContext(ctx)
	if err != nil {
		return diag.Errorf("Error waiting for vnpaycloud_vpn_gateway %s to delete: %s", d.Id(), err)
	}

	return nil
}
