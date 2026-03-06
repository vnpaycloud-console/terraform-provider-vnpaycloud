package privategateway

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

func ResourcePrivateGateway() *schema.Resource {
	return &schema.Resource{
		CreateContext: resourcePrivateGatewayCreate,
		ReadContext:   resourcePrivateGatewayRead,
		UpdateContext: resourcePrivateGatewayUpdate,
		DeleteContext: resourcePrivateGatewayDelete,
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
			},
			"description": {
				Type:     schema.TypeString,
				Optional: true,
			},
			"load_balancer_id": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"subnet_id": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"flavor_id": {
				Type:     schema.TypeString,
				Computed: true,
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

func resourcePrivateGatewayCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	cfg := meta.(*config.Config)

	createOpts := dto.CreatePrivateGatewayRequest{
		Name:        d.Get("name").(string),
		Description: d.Get("description").(string),
	}

	tflog.Debug(ctx, "vnpaycloud_private_gateway create options", map[string]interface{}{"create_opts": createOpts})

	createResp := &dto.PrivateGatewayResponse{}
	_, err := cfg.Client.Post(ctx, client.ApiPath.PrivateGateways(cfg.ProjectID), createOpts, createResp, nil)
	if err != nil {
		return diag.Errorf("Error creating vnpaycloud_private_gateway: %s", err)
	}

	d.SetId(createResp.PrivateGateway.ID)

	stateConf := &retry.StateChangeConf{
		Pending:    []string{"initiating", "creating"},
		Target:     []string{"active", "created"},
		Refresh:    privateGatewayStateRefreshFunc(ctx, cfg.Client, cfg.ProjectID, createResp.PrivateGateway.ID),
		Timeout:    d.Timeout(schema.TimeoutCreate),
		Delay:      5 * time.Second,
		MinTimeout: 3 * time.Second,
	}

	_, err = stateConf.WaitForStateContext(ctx)
	if err != nil {
		return diag.Errorf("Error waiting for vnpaycloud_private_gateway %s to become ready: %s", createResp.PrivateGateway.ID, err)
	}

	return resourcePrivateGatewayRead(ctx, d, meta)
}

func resourcePrivateGatewayRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	cfg := meta.(*config.Config)

	pgwResp := &dto.PrivateGatewayResponse{}
	_, err := cfg.Client.Get(ctx, client.ApiPath.PrivateGatewayWithID(cfg.ProjectID, d.Id()), pgwResp, nil)
	if err != nil {
		return diag.FromErr(util.CheckNotFound(d, err, "Error retrieving vnpaycloud_private_gateway"))
	}

	tflog.Debug(ctx, "Retrieved vnpaycloud_private_gateway "+d.Id(), map[string]interface{}{"private_gateway": pgwResp.PrivateGateway})

	d.Set("name", pgwResp.PrivateGateway.Name)
	d.Set("description", pgwResp.PrivateGateway.Description)
	d.Set("load_balancer_id", pgwResp.PrivateGateway.LoadBalancerID)
	d.Set("subnet_id", pgwResp.PrivateGateway.SubnetID)
	d.Set("flavor_id", pgwResp.PrivateGateway.FlavorID)
	d.Set("status", pgwResp.PrivateGateway.Status)
	d.Set("created_at", pgwResp.PrivateGateway.CreatedAt)

	return nil
}

func resourcePrivateGatewayUpdate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	cfg := meta.(*config.Config)

	if d.HasChanges("name", "description") {
		updateOpts := dto.UpdatePrivateGatewayRequest{
			Name:        d.Get("name").(string),
			Description: d.Get("description").(string),
		}

		tflog.Debug(ctx, "Updating vnpaycloud_private_gateway", map[string]interface{}{
			"id":          d.Id(),
			"update_opts": updateOpts,
		})

		_, err := cfg.Client.Put(ctx, client.ApiPath.PrivateGatewayWithID(cfg.ProjectID, d.Id()), updateOpts, nil, nil)
		if err != nil {
			return diag.Errorf("Error updating vnpaycloud_private_gateway %s: %s", d.Id(), err)
		}
	}

	return resourcePrivateGatewayRead(ctx, d, meta)
}

func resourcePrivateGatewayDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	cfg := meta.(*config.Config)

	pgwResp := &dto.PrivateGatewayResponse{}
	_, err := cfg.Client.Get(ctx, client.ApiPath.PrivateGatewayWithID(cfg.ProjectID, d.Id()), pgwResp, nil)
	if err != nil {
		return diag.FromErr(util.CheckDeleted(d, err, "Error retrieving vnpaycloud_private_gateway"))
	}

	if pgwResp.PrivateGateway.Status != "deleting" {
		if _, err := cfg.Client.Delete(ctx, client.ApiPath.PrivateGatewayWithID(cfg.ProjectID, d.Id()), nil); err != nil {
			return diag.FromErr(util.CheckDeleted(d, err, "Error deleting vnpaycloud_private_gateway"))
		}
	}

	stateConf := &retry.StateChangeConf{
		Pending:    []string{"deleting", "active", "created"},
		Target:     []string{"deleted"},
		Refresh:    privateGatewayStateRefreshFunc(ctx, cfg.Client, cfg.ProjectID, d.Id()),
		Timeout:    d.Timeout(schema.TimeoutDelete),
		Delay:      5 * time.Second,
		MinTimeout: 3 * time.Second,
	}

	_, err = stateConf.WaitForStateContext(ctx)
	if err != nil {
		return diag.Errorf("Error waiting for vnpaycloud_private_gateway %s to delete: %s", d.Id(), err)
	}

	return nil
}
