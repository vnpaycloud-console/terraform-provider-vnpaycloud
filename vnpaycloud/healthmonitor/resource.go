package healthmonitor

import (
	"context"
	"fmt"
	"regexp"
	"strings"
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
		UpdateContext: resourceHealthMonitorUpdate,
		DeleteContext: resourceHealthMonitorDelete,
		CustomizeDiff: func(_ context.Context, d *schema.ResourceDiff, _ interface{}) error {
			delay := d.Get("delay").(int)
			timeout := d.Get("timeout").(int)
			if delay > 0 && timeout > 0 && timeout > delay {
				return fmt.Errorf("timeout (%d) must be <= delay (%d)", timeout, delay)
			}
			return nil
		},
		Importer: &schema.ResourceImporter{
			StateContext: func(ctx context.Context, d *schema.ResourceData, meta interface{}) ([]*schema.ResourceData, error) {
				cfg := meta.(*config.Config)
				resp := &dto.HealthMonitorResponse{}
				if _, err := cfg.Client.Get(ctx, client.ApiPath.HealthMonitorWithID(cfg.ProjectID, d.Id()), resp, nil); err != nil {
					return nil, fmt.Errorf("vnpaycloud_lb_health_monitor %q not found: %w", d.Id(), err)
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
				Optional: true,
				Computed: true,
				ValidateFunc: validation.All(
					validation.StringLenBetween(0, 250),
					validation.StringMatch(regexp.MustCompile(`^([^ ].*[^ ])?$`), "name must not start or end with whitespace"),
				),
			},
			"pool_id": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},
			"type": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
				ValidateFunc: validation.StringInSlice(
					[]string{"HTTP", "HTTPS", "TCP", "PING", "TLS-HELLO", "UDP-CONNECT", "SCTP"}, false,
				),
			},
			"delay": {
				Type:         schema.TypeInt,
				Required:     true,
				ValidateFunc: validation.IntAtLeast(1),
			},
			"timeout": {
				Type:         schema.TypeInt,
				Required:     true,
				ValidateFunc: validation.IntAtLeast(1),
			},
			"max_retries": {
				Type:         schema.TypeInt,
				Required:     true,
				ValidateFunc: validation.IntBetween(1, 10),
			},
			"max_retries_down": {
				Type:         schema.TypeInt,
				Optional:     true,
				Computed:     true,
				ValidateFunc: validation.IntBetween(1, 10),
			},
			"http_method": {
				Type:     schema.TypeString,
				Optional: true,
				Computed: true,
				ValidateFunc: validation.StringInSlice(
					[]string{"GET", "POST", "PUT", "DELETE", "HEAD", "OPTIONS", "PATCH", "CONNECT", "TRACE"}, false,
				),
			},
			"url_path": {
				Type:     schema.TypeString,
				Optional: true,
				Computed: true,
				ValidateFunc: validation.StringMatch(
					regexp.MustCompile(`^/`), "url_path must start with /",
				),
			},
			"expected_codes": {
				Type:     schema.TypeString,
				Optional: true,
				Computed: true,
				ValidateFunc: validation.StringMatch(
					regexp.MustCompile(`^(\d{3}(,\d{3})*|\d{3}-\d{3})$`),
					"expected_codes must be a single code (200), a list (200,201,302), or a range (200-299)",
				),
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
		Name:           d.Get("name").(string),
		PoolID:         d.Get("pool_id").(string),
		Type:           d.Get("type").(string),
		Delay:          d.Get("delay").(int),
		Timeout:        d.Get("timeout").(int),
		MaxRetries:     d.Get("max_retries").(int),
		MaxRetriesDown: d.Get("max_retries_down").(int),
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

	d.Set("name", resp.HealthMonitor.Name)
	d.Set("pool_id", resp.HealthMonitor.PoolID)
	d.Set("type", resp.HealthMonitor.Type)
	d.Set("delay", resp.HealthMonitor.Delay)
	d.Set("timeout", resp.HealthMonitor.Timeout)
	d.Set("max_retries", resp.HealthMonitor.MaxRetries)
	d.Set("max_retries_down", resp.HealthMonitor.MaxRetriesDown)
	d.Set("http_method", resp.HealthMonitor.HTTPMethod)
	d.Set("url_path", resp.HealthMonitor.URLPath)
	d.Set("expected_codes", resp.HealthMonitor.ExpectedCodes)
	d.Set("status", resp.HealthMonitor.Status)

	return nil
}

func resourceHealthMonitorUpdate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	cfg := meta.(*config.Config)

	if d.HasChanges("name", "delay", "timeout", "max_retries", "max_retries_down", "http_method", "url_path", "expected_codes") {
		waitBefore := &retry.StateChangeConf{
			Pending:    []string{"initiating", "creating", "pending_create", "pending_update"},
			Target:     []string{"active", "created"},
			Refresh:    healthMonitorStateRefreshFunc(ctx, cfg.Client, cfg.ProjectID, d.Id()),
			Timeout:    d.Timeout(schema.TimeoutUpdate),
			Delay:      3 * time.Second,
			MinTimeout: 3 * time.Second,
		}
		if _, err := waitBefore.WaitForStateContext(ctx); err != nil {
			return diag.Errorf("Error waiting for vnpaycloud_lb_health_monitor %s to become ready before update: %s", d.Id(), err)
		}

		updateOpts := dto.UpdateHealthMonitorRequest{
			Name:           d.Get("name").(string),
			Delay:          d.Get("delay").(int),
			Timeout:        d.Get("timeout").(int),
			MaxRetries:     d.Get("max_retries").(int),
			MaxRetriesDown: d.Get("max_retries_down").(int),
			HTTPMethod:     d.Get("http_method").(string),
			URLPath:        d.Get("url_path").(string),
			ExpectedCodes: d.Get("expected_codes").(string),
		}

		tflog.Debug(ctx, "vnpaycloud_lb_health_monitor update options", map[string]interface{}{"update_opts": updateOpts})

		err := util.RetryLBPendingPut(ctx, d.Timeout(schema.TimeoutUpdate), func() error {
			_, putErr := cfg.Client.Put(ctx, client.ApiPath.HealthMonitorWithID(cfg.ProjectID, d.Id()), updateOpts, nil, nil)
			return putErr
		})
		if err != nil {
			return diag.Errorf("Error updating vnpaycloud_lb_health_monitor %s: %s", d.Id(), err)
		}

		waitAfter := &retry.StateChangeConf{
			Pending:    []string{"pending_update", "creating"},
			Target:     []string{"active", "created"},
			Refresh:    healthMonitorStateRefreshFunc(ctx, cfg.Client, cfg.ProjectID, d.Id()),
			Timeout:    d.Timeout(schema.TimeoutUpdate),
			Delay:      3 * time.Second,
			MinTimeout: 3 * time.Second,
		}
		if _, err := waitAfter.WaitForStateContext(ctx); err != nil {
			return diag.Errorf("Error waiting for vnpaycloud_lb_health_monitor %s to converge after update: %s", d.Id(), err)
		}
	}

	return resourceHealthMonitorRead(ctx, d, meta)
}

func resourceHealthMonitorDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	cfg := meta.(*config.Config)

	deleteErr := retry.RetryContext(ctx, d.Timeout(schema.TimeoutDelete), func() *retry.RetryError {
		_, err := cfg.Client.Delete(ctx, client.ApiPath.HealthMonitorWithID(cfg.ProjectID, d.Id()), nil)
		if err != nil && strings.Contains(err.Error(), "not active") {
			return retry.RetryableError(err)
		}
		if err != nil {
			return retry.NonRetryableError(err)
		}
		return nil
	})
	if deleteErr != nil {
		return diag.FromErr(util.CheckDeleted(d, deleteErr, "Error deleting vnpaycloud_lb_health_monitor"))
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
