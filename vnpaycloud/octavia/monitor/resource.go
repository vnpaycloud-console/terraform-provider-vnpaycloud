package monitor

import (
	"context"
	"fmt"
	"log"
	"strings"
	"terraform-provider-vnpaycloud/vnpaycloud/config"
	"terraform-provider-vnpaycloud/vnpaycloud/shared"
	"terraform-provider-vnpaycloud/vnpaycloud/util"
	"time"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/retry"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"

	"github.com/vnpaycloud-console/gophercloud/v2/openstack/loadbalancer/v2/monitors"
	"github.com/vnpaycloud-console/gophercloud/v2/openstack/loadbalancer/v2/pools"
)

func ResourceMonitorV2() *schema.Resource {
	return &schema.Resource{
		CreateContext: resourceMonitorV2Create,
		ReadContext:   resourceMonitorV2Read,
		UpdateContext: resourceMonitorV2Update,
		DeleteContext: resourceMonitorV2Delete,
		Importer: &schema.ResourceImporter{
			StateContext: resourceMonitorV2Import,
		},

		Timeouts: &schema.ResourceTimeout{
			Create: schema.DefaultTimeout(10 * time.Minute),
			Update: schema.DefaultTimeout(10 * time.Minute),
			Delete: schema.DefaultTimeout(10 * time.Minute),
		},

		Schema: map[string]*schema.Schema{
			"region": {
				Type:     schema.TypeString,
				Optional: true,
				Computed: true,
				ForceNew: true,
			},

			"pool_id": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},

			"name": {
				Type:     schema.TypeString,
				Optional: true,
			},

			"domain_name": {
				Type:     schema.TypeString,
				Optional: true,
			},

			"tenant_id": {
				Type:     schema.TypeString,
				Optional: true,
				Computed: true,
				ForceNew: true,
			},

			"type": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
				ValidateFunc: validation.StringInSlice([]string{
					"HTTP", "HTTPS", "PING", "SCTP", "TCP",
					"TLS-HELLO", "UDP-CONNECT",
				}, false),
			},

			"delay": {
				Type:     schema.TypeInt,
				Required: true,
			},

			"timeout": {
				Type:     schema.TypeInt,
				Required: true,
			},

			"max_retries": {
				Type:         schema.TypeInt,
				Required:     true,
				ValidateFunc: validation.IntBetween(1, 10),
			},

			"max_retries_down": {
				Type:         schema.TypeInt,
				Optional:     true,
				ValidateFunc: validation.IntBetween(1, 10),
				Computed:     true,
			},

			"url_path": {
				Type:     schema.TypeString,
				Optional: true,
				Computed: true,
			},

			"http_method": {
				Type:     schema.TypeString,
				Optional: true,
				ValidateFunc: validation.StringInSlice([]string{
					"CONNECT", "DELETE", "GET", "HEAD", "OPTIONS",
					"PATCH", "POST", "PUT", "TRACE",
				}, false),
				Computed: true,
			},

			"http_version": {
				Type:     schema.TypeString,
				Optional: true,
				ValidateFunc: validation.StringInSlice([]string{
					"1.0", "1.1",
				}, false),
			},

			"expected_codes": {
				Type:     schema.TypeString,
				Optional: true,
				Computed: true,
			},

			"admin_state_up": {
				Type:     schema.TypeBool,
				Default:  true,
				Optional: true,
			},
		},
	}
}

func resourceMonitorV2Create(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	config := meta.(*config.Config)
	lbClient, err := config.LoadBalancerV2Client(ctx, util.GetRegion(d, config))
	if err != nil {
		return diag.Errorf("Error creating VNPAYCLOUD networking client: %s", err)
	}

	adminStateUp := d.Get("admin_state_up").(bool)

	createOpts := monitors.CreateOpts{
		PoolID:         d.Get("pool_id").(string),
		TenantID:       d.Get("tenant_id").(string),
		Type:           d.Get("type").(string),
		Delay:          d.Get("delay").(int),
		Timeout:        d.Get("timeout").(int),
		MaxRetries:     d.Get("max_retries").(int),
		MaxRetriesDown: d.Get("max_retries_down").(int),
		URLPath:        d.Get("url_path").(string),
		HTTPMethod:     d.Get("http_method").(string),
		HTTPVersion:    d.Get("http_version").(string),
		ExpectedCodes:  d.Get("expected_codes").(string),
		Name:           d.Get("name").(string),
		DomainName:     d.Get("domain_name").(string),
		AdminStateUp:   &adminStateUp,
	}

	// Get a clean copy of the parent pool.
	poolID := d.Get("pool_id").(string)
	parentPool, err := pools.Get(ctx, lbClient, poolID).Extract()
	if err != nil {
		return diag.Errorf("Unable to retrieve parent vnpaycloud_lb_pool %s: %s", poolID, err)
	}

	// Wait for parent pool to become active before continuing.
	timeout := d.Timeout(schema.TimeoutCreate)
	err = shared.WaitForLBV2Pool(ctx, lbClient, parentPool, "ACTIVE", shared.GetLbPendingStatuses(), timeout)
	if err != nil {
		return diag.FromErr(err)
	}

	log.Printf("[DEBUG] vnpaycloud_lb_monitor create options: %#v", createOpts)
	var monitor *monitors.Monitor
	err = retry.RetryContext(ctx, timeout, func() *retry.RetryError {
		monitor, err = monitors.Create(ctx, lbClient, createOpts).Extract()
		if err != nil {
			return util.CheckForRetryableError(err)
		}
		return nil
	})

	if err != nil {
		return diag.Errorf("Unable to create vnpaycloud_lb_monitor: %s", err)
	}

	// Wait for monitor to become active before continuing
	err = shared.WaitForLBV2Monitor(ctx, lbClient, parentPool, monitor, "ACTIVE", shared.GetLbPendingStatuses(), timeout)
	if err != nil {
		return diag.FromErr(err)
	}

	d.SetId(monitor.ID)

	return resourceMonitorV2Read(ctx, d, meta)
}

func resourceMonitorV2Read(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	config := meta.(*config.Config)
	lbClient, err := config.LoadBalancerV2Client(ctx, util.GetRegion(d, config))
	if err != nil {
		return diag.Errorf("Error creating VNPAYCLOUD networking client: %s", err)
	}

	monitor, err := monitors.Get(ctx, lbClient, d.Id()).Extract()
	if err != nil {
		return diag.FromErr(util.CheckDeleted(d, err, "monitor"))
	}

	log.Printf("[DEBUG] Retrieved vnpaycloud_lb_monitor %s: %#v", d.Id(), monitor)

	d.Set("tenant_id", monitor.ProjectID)
	d.Set("type", monitor.Type)
	d.Set("delay", monitor.Delay)
	d.Set("timeout", monitor.Timeout)
	d.Set("max_retries", monitor.MaxRetries)
	d.Set("max_retries_down", monitor.MaxRetriesDown)
	d.Set("url_path", monitor.URLPath)
	d.Set("http_method", monitor.HTTPMethod)
	d.Set("http_version", monitor.HTTPVersion)
	d.Set("expected_codes", monitor.ExpectedCodes)
	d.Set("admin_state_up", monitor.AdminStateUp)
	d.Set("name", monitor.Name)
	d.Set("domain_name", monitor.DomainName)
	d.Set("region", util.GetRegion(d, config))

	if len(monitor.Pools) > 0 && monitor.Pools[0].ID != "" {
		d.Set("pool_id", monitor.Pools[0].ID)
	}

	return nil
}

func resourceMonitorV2Update(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var hasChange bool
	config := meta.(*config.Config)
	lbClient, err := config.LoadBalancerV2Client(ctx, util.GetRegion(d, config))
	if err != nil {
		return diag.Errorf("Error creating VNPAYCLOUD networking client: %s", err)
	}

	var opts monitors.UpdateOpts

	if d.HasChange("url_path") {
		hasChange = true
		opts.URLPath = d.Get("url_path").(string)
	}
	if d.HasChange("expected_codes") {
		hasChange = true
		opts.ExpectedCodes = d.Get("expected_codes").(string)
	}
	if d.HasChange("delay") {
		hasChange = true
		opts.Delay = d.Get("delay").(int)
	}
	if d.HasChange("timeout") {
		hasChange = true
		opts.Timeout = d.Get("timeout").(int)
	}
	if d.HasChange("max_retries") {
		hasChange = true
		opts.MaxRetries = d.Get("max_retries").(int)
	}
	if d.HasChange("max_retries_down") {
		hasChange = true
		opts.MaxRetriesDown = d.Get("max_retries_down").(int)
	}
	if d.HasChange("admin_state_up") {
		hasChange = true
		asu := d.Get("admin_state_up").(bool)
		opts.AdminStateUp = &asu
	}
	if d.HasChange("name") {
		hasChange = true
		name := d.Get("name").(string)
		opts.Name = &name
	}
	if d.HasChange("domain_name") {
		hasChange = true
		v := d.Get("domain_name").(string)
		opts.DomainName = &v
	}
	if d.HasChange("http_method") {
		hasChange = true
		opts.HTTPMethod = d.Get("http_method").(string)
	}
	if d.HasChange("http_version") {
		hasChange = true
		v := d.Get("http_version").(string)
		opts.HTTPVersion = &v
	}

	if !hasChange {
		log.Printf("[DEBUG] vnpaycloud_lb_monitor %s: nothing to update", d.Id())
		return resourceMonitorV2Read(ctx, d, meta)
	}

	// Get a clean copy of the parent pool.
	poolID := d.Get("pool_id").(string)
	parentPool, err := pools.Get(ctx, lbClient, poolID).Extract()
	if err != nil {
		return diag.Errorf("Unable to retrieve parent vnpaycloud_lb_pool %s: %s", poolID, err)
	}

	// Get a clean copy of the monitor.
	monitor, err := monitors.Get(ctx, lbClient, d.Id()).Extract()
	if err != nil {
		return diag.Errorf("Unable to retrieve vnpaycloud_lb_monitor %s: %s", d.Id(), err)
	}

	// Wait for parent pool to become active before continuing.
	timeout := d.Timeout(schema.TimeoutUpdate)
	err = shared.WaitForLBV2Pool(ctx, lbClient, parentPool, "ACTIVE", shared.GetLbPendingStatuses(), timeout)
	if err != nil {
		return diag.FromErr(err)
	}

	// Wait for monitor to become active before continuing.
	err = shared.WaitForLBV2Monitor(ctx, lbClient, parentPool, monitor, "ACTIVE", shared.GetLbPendingStatuses(), timeout)
	if err != nil {
		return diag.FromErr(err)
	}

	log.Printf("[DEBUG] vnpaycloud_lb_monitor %s update options: %#v", d.Id(), opts)
	err = retry.RetryContext(ctx, timeout, func() *retry.RetryError {
		_, err = monitors.Update(ctx, lbClient, d.Id(), opts).Extract()
		if err != nil {
			return util.CheckForRetryableError(err)
		}
		return nil
	})

	if err != nil {
		return diag.Errorf("Unable to update vnpaycloud_lb_monitor %s: %s", d.Id(), err)
	}

	// Wait for monitor to become active before continuing
	err = shared.WaitForLBV2Monitor(ctx, lbClient, parentPool, monitor, "ACTIVE", shared.GetLbPendingStatuses(), timeout)
	if err != nil {
		return diag.FromErr(err)
	}

	return resourceMonitorV2Read(ctx, d, meta)
}

func resourceMonitorV2Delete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	config := meta.(*config.Config)
	lbClient, err := config.LoadBalancerV2Client(ctx, util.GetRegion(d, config))
	if err != nil {
		return diag.Errorf("Error creating VNPAYCLOUD networking client: %s", err)
	}

	// Get a clean copy of the parent pool.
	poolID := d.Get("pool_id").(string)
	parentPool, err := pools.Get(ctx, lbClient, poolID).Extract()
	if err != nil {
		return diag.Errorf("Unable to retrieve parent vnpaycloud_lb_pool (%s)"+
			" for the vnpaycloud_lb_monitor: %s", poolID, err)
	}

	// Get a clean copy of the monitor.
	monitor, err := monitors.Get(ctx, lbClient, d.Id()).Extract()
	if err != nil {
		return diag.FromErr(util.CheckDeleted(d, err, "Unable to retrieve vnpaycloud_lb_monitor"))
	}

	// Wait for parent pool to become active before continuing
	timeout := d.Timeout(schema.TimeoutUpdate)
	err = shared.WaitForLBV2Pool(ctx, lbClient, parentPool, "ACTIVE", shared.GetLbPendingStatuses(), timeout)
	if err != nil {
		return diag.FromErr(err)
	}

	log.Printf("[DEBUG] Deleting vnpaycloud_lb_monitor %s", d.Id())
	err = retry.RetryContext(ctx, timeout, func() *retry.RetryError {
		err = monitors.Delete(ctx, lbClient, d.Id()).ExtractErr()
		if err != nil {
			return util.CheckForRetryableError(err)
		}
		return nil
	})

	if err != nil {
		return diag.FromErr(util.CheckDeleted(d, err, "Error deleting vnpaycloud_lb_monitor"))
	}

	// Wait for monitor to become DELETED
	err = shared.WaitForLBV2Monitor(ctx, lbClient, parentPool, monitor, "DELETED", shared.GetLbPendingDeleteStatuses(), timeout)
	if err != nil {
		return diag.FromErr(err)
	}

	return nil
}

func resourceMonitorV2Import(ctx context.Context, d *schema.ResourceData, meta interface{}) ([]*schema.ResourceData, error) {
	parts := strings.SplitN(d.Id(), "/", 2)
	monitorID := parts[0]

	if len(monitorID) == 0 {
		return nil, fmt.Errorf("Invalid format specified for vnpaycloud_lb_monitor. Format must be <monitorID>[/<poolID>]")
	}

	d.SetId(monitorID)

	if len(parts) == 2 {
		d.Set("pool_id", parts[1])
	}

	return []*schema.ResourceData{d}, nil
}
