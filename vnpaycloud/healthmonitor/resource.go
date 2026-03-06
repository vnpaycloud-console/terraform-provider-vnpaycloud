package healthmonitor

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
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
)

func ResourceHealthMonitor() *schema.Resource {
	return &schema.Resource{
		CreateContext: resourceHealthMonitorCreate,
		ReadContext:   resourceHealthMonitorRead,
		DeleteContext: resourceHealthMonitorDelete,
		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},
		Timeouts: &schema.ResourceTimeout{
			Create: schema.DefaultTimeout(10 * time.Minute),
			Delete: schema.DefaultTimeout(10 * time.Minute),
		},
		Schema: map[string]*schema.Schema{
			"pool_id": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},
			"type": {
				Type:         schema.TypeString,
				Required:     true,
				ForceNew:     true,
				ValidateFunc: validation.StringInSlice([]string{"HTTP", "HTTPS", "TCP", "PING"}, false),
			},
			"delay": {
				Type:         schema.TypeInt,
				Required:     true,
				ForceNew:     true,
				ValidateFunc: validation.IntAtLeast(1),
			},
			"timeout": {
				Type:         schema.TypeInt,
				Required:     true,
				ForceNew:     true,
				ValidateFunc: validation.IntAtLeast(1),
			},
			"max_retries": {
				Type:         schema.TypeInt,
				Required:     true,
				ForceNew:     true,
				ValidateFunc: validation.IntBetween(1, 10),
			},
			"http_method": {
				Type:     schema.TypeString,
				Optional: true,
				ForceNew: true,
				Computed: true,
			},
			"url_path": {
				Type:     schema.TypeString,
				Optional: true,
				ForceNew: true,
				Computed: true,
			},
			"expected_codes": {
				Type:     schema.TypeString,
				Optional: true,
				ForceNew: true,
				Computed: true,
			},
			"status": {
				Type:     schema.TypeString,
				Computed: true,
			},
		},
	}
}

func resourceHealthMonitorCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	cfg := meta.(*config.Config)

	createOpts := dto.CreateHealthMonitorRequest{
		PoolID:     d.Get("pool_id").(string),
		Type:       d.Get("type").(string),
		Delay:      d.Get("delay").(int),
		Timeout:    d.Get("timeout").(int),
		MaxRetries: d.Get("max_retries").(int),
	}

	if v, ok := d.GetOk("http_method"); ok {
		createOpts.HTTPMethod = v.(string)
	}
	if v, ok := d.GetOk("url_path"); ok {
		createOpts.URLPath = v.(string)
	}
	if v, ok := d.GetOk("expected_codes"); ok {
		createOpts.ExpectedCodes = v.(string)
	}

	tflog.Debug(ctx, "vnpaycloud_lb_health_monitor create options", map[string]interface{}{"create_opts": createOpts})

	createResp := &dto.HealthMonitorResponse{}
	_, err := cfg.Client.Post(ctx, client.ApiPath.HealthMonitors(cfg.ProjectID), createOpts, createResp, nil)
	if err != nil {
		return diag.Errorf("Error creating vnpaycloud_lb_health_monitor: %s", err)
	}

	d.SetId(createResp.HealthMonitor.ID)

	stateConf := &retry.StateChangeConf{
		Pending:    []string{"initiating", "creating", "pending_create"},
		Target:     []string{"active", "created"},
		Refresh:    healthMonitorStateRefreshFunc(ctx, cfg.Client, cfg.ProjectID, createResp.HealthMonitor.ID),
		Timeout:    d.Timeout(schema.TimeoutCreate),
		Delay:      5 * time.Second,
		MinTimeout: 3 * time.Second,
	}

	_, err = stateConf.WaitForStateContext(ctx)
	if err != nil {
		return diag.Errorf("Error waiting for vnpaycloud_lb_health_monitor %s to become ready: %s", createResp.HealthMonitor.ID, err)
	}

	return resourceHealthMonitorRead(ctx, d, meta)
}

func resourceHealthMonitorRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	cfg := meta.(*config.Config)

	resp := &dto.HealthMonitorResponse{}
	_, err := cfg.Client.Get(ctx, client.ApiPath.HealthMonitorWithID(cfg.ProjectID, d.Id()), resp, nil)
	if err != nil {
		return diag.FromErr(util.CheckNotFound(d, err, "Error retrieving vnpaycloud_lb_health_monitor"))
	}

	tflog.Debug(ctx, "Retrieved vnpaycloud_lb_health_monitor "+d.Id(), map[string]interface{}{"health_monitor": resp.HealthMonitor})

	d.Set("pool_id", resp.HealthMonitor.PoolID)
	d.Set("type", resp.HealthMonitor.Type)
	d.Set("delay", resp.HealthMonitor.Delay)
	d.Set("timeout", resp.HealthMonitor.Timeout)
	d.Set("max_retries", resp.HealthMonitor.MaxRetries)
	d.Set("http_method", resp.HealthMonitor.HTTPMethod)
	d.Set("url_path", resp.HealthMonitor.URLPath)
	d.Set("expected_codes", resp.HealthMonitor.ExpectedCodes)
	d.Set("status", resp.HealthMonitor.Status)

	return nil
}

func resourceHealthMonitorDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	cfg := meta.(*config.Config)

	if _, err := cfg.Client.Delete(ctx, client.ApiPath.HealthMonitorWithID(cfg.ProjectID, d.Id()), nil); err != nil {
		return diag.FromErr(util.CheckDeleted(d, err, "Error deleting vnpaycloud_lb_health_monitor"))
	}

	stateConf := &retry.StateChangeConf{
		Pending:    []string{"deleting", "active", "created", "pending_delete"},
		Target:     []string{"deleted"},
		Refresh:    healthMonitorStateRefreshFunc(ctx, cfg.Client, cfg.ProjectID, d.Id()),
		Timeout:    d.Timeout(schema.TimeoutDelete),
		Delay:      5 * time.Second,
		MinTimeout: 3 * time.Second,
	}

	_, err := stateConf.WaitForStateContext(ctx)
	if err != nil {
		return diag.Errorf("Error waiting for vnpaycloud_lb_health_monitor %s to delete: %s", d.Id(), err)
	}

	return nil
}
