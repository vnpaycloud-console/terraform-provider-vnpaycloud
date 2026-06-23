package servicegateway

import (
	"context"
	"regexp"
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

func ResourceServiceGateway() *schema.Resource {
	return &schema.Resource{
		CreateContext: resourceServiceGatewayCreate,
		ReadContext:   resourceServiceGatewayRead,
		UpdateContext: resourceServiceGatewayUpdate,
		DeleteContext: resourceServiceGatewayDelete,
		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
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
					validation.StringLenBetween(3, 250),
					validation.StringMatch(regexp.MustCompile(`^[a-zA-Z0-9-_. ]*$`), "name may only contain ASCII letters, digits, spaces, and the characters - _ ."),
				),
			},
			"description": {
				Type:     schema.TypeString,
				Optional: true,
			},
			"subnet_id": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},
			"vpc_id": {
				Type:     schema.TypeString,
				Optional: true,
				ForceNew: true,
			},
			"flavor_id": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},
			"allowed_icmp": {
				Type:     schema.TypeBool,
				Optional: true,
				Computed: true,
			},
			"vip_address": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"load_balancer_id": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"port_id": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"operating_status": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"provisioning_status": {
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

func resourceServiceGatewayCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	cfg := meta.(*config.Config)

	createOpts := dto.CreateServiceGatewayRequest{
		Name:        d.Get("name").(string),
		Description: d.Get("description").(string),
		VPCID:       d.Get("vpc_id").(string),
		SubnetID:    d.Get("subnet_id").(string),
		FlavorID:    d.Get("flavor_id").(string),
	}

	tflog.Debug(ctx, "vnpaycloud_service_gateway create options", map[string]interface{}{"create_opts": createOpts})

	createResp := &dto.ServiceGatewayResponse{}
	_, err := cfg.Client.Post(ctx, client.ApiPath.ServiceGateways(cfg.ProjectID), createOpts, createResp, nil)
	if err != nil {
		return diag.Errorf("Error creating vnpaycloud_service_gateway: %s", err)
	}

	d.SetId(createResp.ServiceGateway.ID)

	stateConf := &retry.StateChangeConf{
		Pending:    []string{"creating", "unknown"},
		Target:     []string{"active"},
		Refresh:    serviceGatewayStateRefreshFunc(ctx, cfg.Client, cfg.ProjectID, d.Id()),
		Timeout:    d.Timeout(schema.TimeoutCreate),
		Delay:      5 * time.Second,
		MinTimeout: 3 * time.Second,
	}
	if _, err := stateConf.WaitForStateContext(ctx); err != nil {
		return diag.Errorf("Error waiting for vnpaycloud_service_gateway %s to become ready: %s", d.Id(), err)
	}

	// allowed_icmp is not accepted at create; apply it post-create via the dedicated action.
	if d.Get("allowed_icmp").(bool) {
		icmpOpts := dto.SetServiceGatewayICMPRequest{AllowedICMP: true}
		if _, err := cfg.Client.Post(ctx, client.ApiPath.ServiceGatewayICMP(cfg.ProjectID, d.Id()), icmpOpts, nil, nil); err != nil {
			return diag.Errorf("Error setting ICMP on vnpaycloud_service_gateway %s: %s", d.Id(), err)
		}
		if err := waitServiceGatewayActive(ctx, cfg, d.Id(), d.Timeout(schema.TimeoutCreate)); err != nil {
			return diag.Errorf("Error waiting for vnpaycloud_service_gateway %s to converge after ICMP change: %s", d.Id(), err)
		}
	}

	return resourceServiceGatewayRead(ctx, d, meta)
}

func resourceServiceGatewayRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	cfg := meta.(*config.Config)

	resp := &dto.ServiceGatewayResponse{}
	_, err := cfg.Client.Get(ctx, client.ApiPath.ServiceGatewayWithID(cfg.ProjectID, d.Id()), resp, nil)
	if err != nil {
		return diag.FromErr(util.CheckNotFound(d, err, "Error retrieving vnpaycloud_service_gateway"))
	}

	sg := resp.ServiceGateway
	tflog.Debug(ctx, "Retrieved vnpaycloud_service_gateway "+d.Id(), map[string]interface{}{"service_gateway": sg})

	d.Set("name", sg.Name)
	d.Set("description", sg.Description)
	d.Set("subnet_id", sg.SubnetID)
	d.Set("vpc_id", sg.VPCID)
	d.Set("flavor_id", sg.FlavorID)
	d.Set("allowed_icmp", sg.AllowedICMP)
	d.Set("vip_address", sg.VipAddress)
	d.Set("load_balancer_id", sg.LoadBalancerID)
	d.Set("port_id", sg.PortID)
	d.Set("operating_status", sg.OperatingStatus)
	d.Set("provisioning_status", sg.ProvisioningStatus)
	d.Set("status", sg.Status)
	d.Set("created_at", sg.CreatedAt)

	return nil
}

func resourceServiceGatewayUpdate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	cfg := meta.(*config.Config)

	if d.HasChanges("name", "description") {
		updateOpts := dto.UpdateServiceGatewayRequest{
			Name:        d.Get("name").(string),
			Description: d.Get("description").(string),
		}
		if _, err := cfg.Client.Put(ctx, client.ApiPath.ServiceGatewayWithID(cfg.ProjectID, d.Id()), updateOpts, nil, nil); err != nil {
			return diag.Errorf("Error updating vnpaycloud_service_gateway %s: %s", d.Id(), err)
		}
		if err := waitServiceGatewayActive(ctx, cfg, d.Id(), d.Timeout(schema.TimeoutUpdate)); err != nil {
			return diag.Errorf("Error waiting for vnpaycloud_service_gateway %s to converge after update: %s", d.Id(), err)
		}
	}

	if d.HasChange("allowed_icmp") {
		icmpOpts := dto.SetServiceGatewayICMPRequest{AllowedICMP: d.Get("allowed_icmp").(bool)}
		if _, err := cfg.Client.Post(ctx, client.ApiPath.ServiceGatewayICMP(cfg.ProjectID, d.Id()), icmpOpts, nil, nil); err != nil {
			return diag.Errorf("Error setting ICMP on vnpaycloud_service_gateway %s: %s", d.Id(), err)
		}
		if err := waitServiceGatewayActive(ctx, cfg, d.Id(), d.Timeout(schema.TimeoutUpdate)); err != nil {
			return diag.Errorf("Error waiting for vnpaycloud_service_gateway %s to converge after ICMP change: %s", d.Id(), err)
		}
	}

	return resourceServiceGatewayRead(ctx, d, meta)
}

func resourceServiceGatewayDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	cfg := meta.(*config.Config)

	if _, err := cfg.Client.Delete(ctx, client.ApiPath.ServiceGatewayWithID(cfg.ProjectID, d.Id()), nil); err != nil {
		return diag.FromErr(util.CheckDeleted(d, err, "Error deleting vnpaycloud_service_gateway"))
	}

	stateConf := &retry.StateChangeConf{
		Pending:    []string{"deleting", "active", "creating", "unknown"},
		Target:     []string{"deleted"},
		Refresh:    serviceGatewayStateRefreshFunc(ctx, cfg.Client, cfg.ProjectID, d.Id()),
		Timeout:    d.Timeout(schema.TimeoutDelete),
		Delay:      5 * time.Second,
		MinTimeout: 3 * time.Second,
	}
	if _, err := stateConf.WaitForStateContext(ctx); err != nil {
		return diag.Errorf("Error waiting for vnpaycloud_service_gateway %s to delete: %s", d.Id(), err)
	}

	return nil
}

func waitServiceGatewayActive(ctx context.Context, cfg *config.Config, id string, timeout time.Duration) error {
	stateConf := &retry.StateChangeConf{
		Pending:    []string{"creating", "unknown"},
		Target:     []string{"active"},
		Refresh:    serviceGatewayStateRefreshFunc(ctx, cfg.Client, cfg.ProjectID, id),
		Timeout:    timeout,
		Delay:      3 * time.Second,
		MinTimeout: 3 * time.Second,
	}
	_, err := stateConf.WaitForStateContext(ctx)
	return err
}
