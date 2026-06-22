package serviceendpoint

import (
	"context"
	"fmt"
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

func ResourceServiceEndpoint() *schema.Resource {
	return &schema.Resource{
		CreateContext: resourceServiceEndpointCreate,
		ReadContext:   resourceServiceEndpointRead,
		UpdateContext: resourceServiceEndpointUpdate,
		DeleteContext: resourceServiceEndpointDelete,
		CustomizeDiff: resourceServiceEndpointCustomizeDiff,
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
			"provider_id": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},
			"service_id": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},
			"service_gateway_id": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},
			"port": {
				Type:         schema.TypeInt,
				Required:     true,
				ForceNew:     true,
				ValidateFunc: validation.IsPortNumber,
			},
			"allowed_cidrs": {
				Type:     schema.TypeList,
				Required: true,
				Elem:     &schema.Schema{Type: schema.TypeString},
			},
			"listener_id": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"pool_id": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"health_monitor_id": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"pool_member_ids": {
				Type:     schema.TypeList,
				Computed: true,
				Elem:     &schema.Schema{Type: schema.TypeString},
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

func expandAllowedCIDRs(raw []interface{}) []string {
	cidrs := make([]string, 0, len(raw))
	for _, v := range raw {
		cidrs = append(cidrs, v.(string))
	}
	return cidrs
}

// resourceServiceEndpointCustomizeDiff enforces that allowed_cidrs is never empty.
// The backend silently ignores an empty allowed_cidrs on update (it only applies the
// list when len > 0) and substitutes "0.0.0.0/0" on create, so an empty list can never
// be represented faithfully and would leave Terraform perpetually out of sync. Requiring
// a non-empty list — and pointing users at ["0.0.0.0/0"] for allow-all — keeps plan and
// state consistent.
func resourceServiceEndpointCustomizeDiff(ctx context.Context, d *schema.ResourceDiff, meta interface{}) error {
	if cidrs := d.Get("allowed_cidrs").([]interface{}); len(cidrs) == 0 {
		return fmt.Errorf("allowed_cidrs must contain at least one CIDR; " +
			"to allow all sources set allowed_cidrs = [\"0.0.0.0/0\"]")
	}

	return nil
}

func resourceServiceEndpointCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	cfg := meta.(*config.Config)

	createOpts := dto.CreateServiceEndpointRequest{
		Name:             d.Get("name").(string),
		Description:      d.Get("description").(string),
		ProviderID:       d.Get("provider_id").(string),
		ServiceID:        d.Get("service_id").(string),
		ServiceGatewayID: d.Get("service_gateway_id").(string),
		Port:             d.Get("port").(int),
		AllowedCIDRs:     expandAllowedCIDRs(d.Get("allowed_cidrs").([]interface{})),
	}

	tflog.Debug(ctx, "vnpaycloud_service_endpoint create options", map[string]interface{}{"create_opts": createOpts})

	createResp := &dto.ServiceEndpointResponse{}
	_, err := cfg.Client.Post(ctx, client.ApiPath.ServiceEndpoints(cfg.ProjectID), createOpts, createResp, nil)
	if err != nil {
		return diag.Errorf("Error creating vnpaycloud_service_endpoint: %s", err)
	}

	d.SetId(createResp.ServiceEndpoint.ID)

	stateConf := &retry.StateChangeConf{
		Pending:    []string{"creating", "unknown"},
		Target:     []string{"active"},
		Refresh:    serviceEndpointStateRefreshFunc(ctx, cfg.Client, cfg.ProjectID, d.Id()),
		Timeout:    d.Timeout(schema.TimeoutCreate),
		Delay:      5 * time.Second,
		MinTimeout: 3 * time.Second,
	}
	if _, err := stateConf.WaitForStateContext(ctx); err != nil {
		return diag.Errorf("Error waiting for vnpaycloud_service_endpoint %s to become ready: %s", d.Id(), err)
	}

	return resourceServiceEndpointRead(ctx, d, meta)
}

func resourceServiceEndpointRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	cfg := meta.(*config.Config)

	resp := &dto.ServiceEndpointResponse{}
	_, err := cfg.Client.Get(ctx, client.ApiPath.ServiceEndpointWithID(cfg.ProjectID, d.Id()), resp, nil)
	if err != nil {
		return diag.FromErr(util.CheckNotFound(d, err, "Error retrieving vnpaycloud_service_endpoint"))
	}

	se := resp.ServiceEndpoint
	tflog.Debug(ctx, "Retrieved vnpaycloud_service_endpoint "+d.Id(), map[string]interface{}{"service_endpoint": se})

	d.Set("name", se.Name)
	d.Set("description", se.Description)
	d.Set("provider_id", se.ProviderID)
	d.Set("service_id", se.ServiceID)
	d.Set("service_gateway_id", se.ServiceGatewayID)
	d.Set("port", se.Port)
	d.Set("allowed_cidrs", se.AllowedCIDRs)
	d.Set("listener_id", se.ListenerID)
	d.Set("pool_id", se.PoolID)
	d.Set("health_monitor_id", se.HealthMonitorID)
	d.Set("pool_member_ids", se.PoolMemberIDs)
	d.Set("operating_status", se.OperatingStatus)
	d.Set("provisioning_status", se.ProvisioningStatus)
	d.Set("status", se.Status)
	d.Set("created_at", se.CreatedAt)

	return nil
}

func resourceServiceEndpointUpdate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	cfg := meta.(*config.Config)

	if d.HasChanges("name", "description", "allowed_cidrs") {
		updateOpts := dto.UpdateServiceEndpointRequest{
			Name:         d.Get("name").(string),
			Description:  d.Get("description").(string),
			AllowedCIDRs: expandAllowedCIDRs(d.Get("allowed_cidrs").([]interface{})),
		}
		if _, err := cfg.Client.Put(ctx, client.ApiPath.ServiceEndpointWithID(cfg.ProjectID, d.Id()), updateOpts, nil, nil); err != nil {
			return diag.Errorf("Error updating vnpaycloud_service_endpoint %s: %s", d.Id(), err)
		}

		stateConf := &retry.StateChangeConf{
			Pending:    []string{"creating", "unknown"},
			Target:     []string{"active"},
			Refresh:    serviceEndpointStateRefreshFunc(ctx, cfg.Client, cfg.ProjectID, d.Id()),
			Timeout:    d.Timeout(schema.TimeoutUpdate),
			Delay:      3 * time.Second,
			MinTimeout: 3 * time.Second,
		}
		if _, err := stateConf.WaitForStateContext(ctx); err != nil {
			return diag.Errorf("Error waiting for vnpaycloud_service_endpoint %s to converge after update: %s", d.Id(), err)
		}
	}

	return resourceServiceEndpointRead(ctx, d, meta)
}

func resourceServiceEndpointDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	cfg := meta.(*config.Config)

	if _, err := cfg.Client.Delete(ctx, client.ApiPath.ServiceEndpointWithID(cfg.ProjectID, d.Id()), nil); err != nil {
		return diag.FromErr(util.CheckDeleted(d, err, "Error deleting vnpaycloud_service_endpoint"))
	}

	stateConf := &retry.StateChangeConf{
		Pending:    []string{"deleting", "active", "creating", "unknown"},
		Target:     []string{"deleted"},
		Refresh:    serviceEndpointStateRefreshFunc(ctx, cfg.Client, cfg.ProjectID, d.Id()),
		Timeout:    d.Timeout(schema.TimeoutDelete),
		Delay:      5 * time.Second,
		MinTimeout: 3 * time.Second,
	}
	if _, err := stateConf.WaitForStateContext(ctx); err != nil {
		return diag.Errorf("Error waiting for vnpaycloud_service_endpoint %s to delete: %s", d.Id(), err)
	}

	return nil
}
