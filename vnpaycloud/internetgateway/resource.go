package internetgateway

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

func ResourceInternetGateway() *schema.Resource {
	return &schema.Resource{
		CreateContext: resourceInternetGatewayCreate,
		ReadContext:   resourceInternetGatewayRead,
		UpdateContext: resourceInternetGatewayUpdate,
		DeleteContext: resourceInternetGatewayDelete,
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
			"description": {
				Type:     schema.TypeString,
				Optional: true,
				ForceNew: true,
			},
			"vpc_id": {
				Type:     schema.TypeString,
				Optional: true,
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

func resourceInternetGatewayCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	cfg := meta.(*config.Config)

	createOpts := dto.CreateInternetGatewayRequest{
		Name:        d.Get("name").(string),
		Description: d.Get("description").(string),
	}

	tflog.Debug(ctx, "vnpaycloud_internet_gateway create options", map[string]interface{}{"create_opts": createOpts})

	createResp := &dto.InternetGatewayResponse{}
	_, err := cfg.Client.Post(ctx, client.ApiPath.InternetGateways(cfg.ProjectID), createOpts, createResp, nil)
	if err != nil {
		return diag.Errorf("Error creating vnpaycloud_internet_gateway: %s", err)
	}

	d.SetId(createResp.InternetGateway.ID)

	stateConf := &retry.StateChangeConf{
		Pending:    []string{"initiating", "creating"},
		Target:     []string{"active", "created"},
		Refresh:    internetGatewayStateRefreshFunc(ctx, cfg.Client, cfg.ProjectID, createResp.InternetGateway.ID),
		Timeout:    d.Timeout(schema.TimeoutCreate),
		Delay:      5 * time.Second,
		MinTimeout: 3 * time.Second,
	}

	_, err = stateConf.WaitForStateContext(ctx)
	if err != nil {
		return diag.Errorf("Error waiting for vnpaycloud_internet_gateway %s to become ready: %s", createResp.InternetGateway.ID, err)
	}

	// If vpc_id is set, attach to VPC
	if vpcID, ok := d.GetOk("vpc_id"); ok && vpcID.(string) != "" {
		attachReq := dto.AttachInternetGatewayToVPCRequest{
			VPCID: vpcID.(string),
		}
		_, err := cfg.Client.Post(ctx, client.ApiPath.InternetGatewayAttachVPC(cfg.ProjectID, d.Id()), attachReq, nil, nil)
		if err != nil {
			return diag.Errorf("Error attaching vnpaycloud_internet_gateway %s to VPC %s: %s", d.Id(), vpcID.(string), err)
		}
	}

	return resourceInternetGatewayRead(ctx, d, meta)
}

func resourceInternetGatewayRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	cfg := meta.(*config.Config)

	igwResp := &dto.InternetGatewayResponse{}
	_, err := cfg.Client.Get(ctx, client.ApiPath.InternetGatewayWithID(cfg.ProjectID, d.Id()), igwResp, nil)
	if err != nil {
		return diag.FromErr(util.CheckNotFound(d, err, "Error retrieving vnpaycloud_internet_gateway"))
	}

	tflog.Debug(ctx, "Retrieved vnpaycloud_internet_gateway "+d.Id(), map[string]interface{}{"internet_gateway": igwResp.InternetGateway})

	d.Set("name", igwResp.InternetGateway.Name)
	d.Set("description", igwResp.InternetGateway.Description)
	d.Set("vpc_id", igwResp.InternetGateway.VPCID)
	d.Set("status", igwResp.InternetGateway.Status)
	d.Set("created_at", igwResp.InternetGateway.CreatedAt)
	d.Set("zone_id", igwResp.InternetGateway.ZoneID)

	return nil
}

func resourceInternetGatewayUpdate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	cfg := meta.(*config.Config)

	if d.HasChange("vpc_id") {
		oldRaw, newRaw := d.GetChange("vpc_id")
		oldVPCID := oldRaw.(string)
		newVPCID := newRaw.(string)

		// Detach from old VPC if it was set
		if oldVPCID != "" {
			tflog.Debug(ctx, "Detaching vnpaycloud_internet_gateway from old VPC", map[string]interface{}{
				"internet_gateway_id": d.Id(),
				"old_vpc_id":          oldVPCID,
			})

			detachReq := dto.DetachInternetGatewayFromVPCRequest{
				VPCID: oldVPCID,
			}
			_, err := cfg.Client.Post(ctx, client.ApiPath.InternetGatewayDetachVPC(cfg.ProjectID, d.Id()), detachReq, nil, nil)
			if err != nil {
				return diag.Errorf("Error detaching vnpaycloud_internet_gateway %s from VPC %s: %s", d.Id(), oldVPCID, err)
			}
		}

		// Attach to new VPC if it is set
		if newVPCID != "" {
			tflog.Debug(ctx, "Attaching vnpaycloud_internet_gateway to new VPC", map[string]interface{}{
				"internet_gateway_id": d.Id(),
				"new_vpc_id":          newVPCID,
			})

			attachReq := dto.AttachInternetGatewayToVPCRequest{
				VPCID: newVPCID,
			}
			_, err := cfg.Client.Post(ctx, client.ApiPath.InternetGatewayAttachVPC(cfg.ProjectID, d.Id()), attachReq, nil, nil)
			if err != nil {
				return diag.Errorf("Error attaching vnpaycloud_internet_gateway %s to VPC %s: %s", d.Id(), newVPCID, err)
			}
		}
	}

	return resourceInternetGatewayRead(ctx, d, meta)
}

func resourceInternetGatewayDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	cfg := meta.(*config.Config)

	// Read current state to check if attached to VPC
	igwResp := &dto.InternetGatewayResponse{}
	_, err := cfg.Client.Get(ctx, client.ApiPath.InternetGatewayWithID(cfg.ProjectID, d.Id()), igwResp, nil)
	if err != nil {
		return diag.FromErr(util.CheckDeleted(d, err, "Error retrieving vnpaycloud_internet_gateway"))
	}

	// Detach from VPC before deleting if currently attached
	if igwResp.InternetGateway.VPCID != "" {
		detachReq := dto.DetachInternetGatewayFromVPCRequest{
			VPCID: igwResp.InternetGateway.VPCID,
		}
		_, err := cfg.Client.Post(ctx, client.ApiPath.InternetGatewayDetachVPC(cfg.ProjectID, d.Id()), detachReq, nil, nil)
		if err != nil {
			return diag.Errorf("Error detaching vnpaycloud_internet_gateway %s from VPC before deletion: %s", d.Id(), err)
		}
	}

	if igwResp.InternetGateway.Status != "deleting" {
		if _, err := cfg.Client.Delete(ctx, client.ApiPath.InternetGatewayWithID(cfg.ProjectID, d.Id()), nil); err != nil {
			return diag.FromErr(util.CheckDeleted(d, err, "Error deleting vnpaycloud_internet_gateway"))
		}
	}

	stateConf := &retry.StateChangeConf{
		Pending:    []string{"deleting", "active", "created"},
		Target:     []string{"deleted"},
		Refresh:    internetGatewayStateRefreshFunc(ctx, cfg.Client, cfg.ProjectID, d.Id()),
		Timeout:    d.Timeout(schema.TimeoutDelete),
		Delay:      5 * time.Second,
		MinTimeout: 3 * time.Second,
	}

	_, err = stateConf.WaitForStateContext(ctx)
	if err != nil {
		return diag.Errorf("Error waiting for vnpaycloud_internet_gateway %s to delete: %s", d.Id(), err)
	}

	return nil
}
