package loadbalancer

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

func ResourceLoadBalancer() *schema.Resource {
	return &schema.Resource{
		CreateContext: resourceLoadBalancerCreate,
		ReadContext:   resourceLoadBalancerRead,
		UpdateContext: resourceLoadBalancerUpdate,
		DeleteContext: resourceLoadBalancerDelete,
		Importer: &schema.ResourceImporter{
			StateContext: func(ctx context.Context, d *schema.ResourceData, meta interface{}) ([]*schema.ResourceData, error) {
				cfg := meta.(*config.Config)
				resp := &dto.LoadBalancerResponse{}
				if _, err := cfg.Client.Get(ctx, client.ApiPath.LoadBalancerWithID(cfg.ProjectID, d.Id()), resp, nil); err != nil {
					return nil, fmt.Errorf("vnpaycloud_lb_loadbalancer %q not found: %w", d.Id(), err)
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
					validation.StringLenBetween(3, 250),
					validation.StringMatch(regexp.MustCompile(`^[^ ].*[^ ]$`), "name must not start or end with whitespace"),
				),
			},
			"description": {
				Type:     schema.TypeString,
				Optional: true,
				ValidateFunc: validation.All(
					validation.StringLenBetween(0, 255),
					validation.StringMatch(regexp.MustCompile(`^[a-zA-Z0-9-_. ]*$`), "description may only contain ASCII letters, digits, spaces, and the characters - _ ."),
				),
			},
			"subnet_id": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},
			"flavor": {
				Type:     schema.TypeString,
				Required: true,
				DiffSuppressFunc: func(_, old, new string, _ *schema.ResourceData) bool {
					return strings.EqualFold(old, new)
				},
			},
			"floating_ip_id": {
				Type:     schema.TypeString,
				Optional: true,
				Computed: true,
				ForceNew: true,
			},
			"vip_address": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"vip_subnet_id": {
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

func resourceLoadBalancerCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	cfg := meta.(*config.Config)

	createOpts := dto.CreateLoadBalancerRequest{
		Name:        d.Get("name").(string),
		Description: d.Get("description").(string),
		SubnetID:    d.Get("subnet_id").(string),
		Flavor:      d.Get("flavor").(string),
	}

	if v, ok := d.GetOk("floating_ip_id"); ok {
		createOpts.FloatingIPID = v.(string)
		createOpts.External = true
	}

	tflog.Debug(ctx, "vnpaycloud_lb_loadbalancer create options", map[string]interface{}{"create_opts": createOpts})

	createResp := &dto.LoadBalancerResponse{}
	_, err := cfg.Client.Post(ctx, client.ApiPath.LoadBalancers(cfg.ProjectID), createOpts, createResp, nil)
	if err != nil {
		return diag.Errorf("Error creating vnpaycloud_lb_loadbalancer: %s", err)
	}

	d.SetId(createResp.LoadBalancer.ID)

	stateConf := &retry.StateChangeConf{
		Pending:    []string{"initiating", "creating", "pending_create", "unknown"},
		Target:     []string{"active", "created"},
		Refresh:    loadBalancerStateRefreshFunc(ctx, cfg.Client, cfg.ProjectID, createResp.LoadBalancer.ID),
		Timeout:    d.Timeout(schema.TimeoutCreate),
		Delay:      5 * time.Second,
		MinTimeout: 3 * time.Second,
	}

	_, err = stateConf.WaitForStateContext(ctx)
	if err != nil {
		return diag.Errorf("Error waiting for vnpaycloud_lb_loadbalancer %s to become ready: %s", createResp.LoadBalancer.ID, err)
	}

	return resourceLoadBalancerRead(ctx, d, meta)
}

func resourceLoadBalancerRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	cfg := meta.(*config.Config)

	lbResp := &dto.LoadBalancerResponse{}
	_, err := cfg.Client.Get(ctx, client.ApiPath.LoadBalancerWithID(cfg.ProjectID, d.Id()), lbResp, nil)
	if err != nil {
		return diag.FromErr(util.CheckNotFound(d, err, "Error retrieving vnpaycloud_lb_loadbalancer"))
	}

	tflog.Debug(ctx, "Retrieved vnpaycloud_lb_loadbalancer "+d.Id(), map[string]interface{}{"load_balancer": lbResp.LoadBalancer})

	d.Set("name", lbResp.LoadBalancer.Name)
	d.Set("description", lbResp.LoadBalancer.Description)
	d.Set("subnet_id", lbResp.LoadBalancer.VipSubnetID)
	d.Set("flavor", lbResp.LoadBalancer.Flavor)
	d.Set("floating_ip_id", lbResp.LoadBalancer.FloatingIPID)
	d.Set("vip_address", lbResp.LoadBalancer.VipAddress)
	d.Set("vip_subnet_id", lbResp.LoadBalancer.VipSubnetID)
	d.Set("status", lbResp.LoadBalancer.Status)
	d.Set("created_at", lbResp.LoadBalancer.CreatedAt)

	return nil
}

func resourceLoadBalancerUpdate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	cfg := meta.(*config.Config)

	if d.HasChanges("name", "description") {
		waitBefore := &retry.StateChangeConf{
			Pending:    []string{"initiating", "creating", "pending_create", "pending_update"},
			Target:     []string{"active", "created"},
			Refresh:    loadBalancerStateRefreshFunc(ctx, cfg.Client, cfg.ProjectID, d.Id()),
			Timeout:    d.Timeout(schema.TimeoutUpdate),
			Delay:      3 * time.Second,
			MinTimeout: 3 * time.Second,
		}
		if _, err := waitBefore.WaitForStateContext(ctx); err != nil {
			return diag.Errorf("Error waiting for vnpaycloud_lb_loadbalancer %s to become ready before update: %s", d.Id(), err)
		}

		updateOpts := dto.UpdateLoadBalancerRequest{
			Name:        d.Get("name").(string),
			Description: d.Get("description").(string),
		}

		tflog.Debug(ctx, "vnpaycloud_lb_loadbalancer update options", map[string]interface{}{"update_opts": updateOpts})

		_, err := cfg.Client.Put(ctx, client.ApiPath.LoadBalancerWithID(cfg.ProjectID, d.Id()), updateOpts, nil, nil)
		if err != nil {
			return diag.Errorf("Error updating vnpaycloud_lb_loadbalancer %s: %s", d.Id(), err)
		}

		stateConf := &retry.StateChangeConf{
			Pending:    []string{"pending_update"},
			Target:     []string{"active", "created"},
			Refresh:    loadBalancerStateRefreshFunc(ctx, cfg.Client, cfg.ProjectID, d.Id()),
			Timeout:    d.Timeout(schema.TimeoutUpdate),
			Delay:      5 * time.Second,
			MinTimeout: 3 * time.Second,
		}

		_, err = stateConf.WaitForStateContext(ctx)
		if err != nil {
			return diag.Errorf("Error waiting for vnpaycloud_lb_loadbalancer %s to converge after update: %s", d.Id(), err)
		}
	}

	if d.HasChange("flavor") {
		waitBefore := &retry.StateChangeConf{
			Pending:    []string{"initiating", "creating", "pending_create", "pending_update"},
			Target:     []string{"active", "created"},
			Refresh:    loadBalancerStateRefreshFunc(ctx, cfg.Client, cfg.ProjectID, d.Id()),
			Timeout:    d.Timeout(schema.TimeoutUpdate),
			Delay:      3 * time.Second,
			MinTimeout: 3 * time.Second,
		}
		if _, err := waitBefore.WaitForStateContext(ctx); err != nil {
			return diag.Errorf("Error waiting for vnpaycloud_lb_loadbalancer %s to become ready before flavor change: %s", d.Id(), err)
		}

		changeOpts := dto.ChangeFlavorLoadBalancerRequest{
			Flavor: d.Get("flavor").(string),
		}

		tflog.Debug(ctx, "vnpaycloud_lb_loadbalancer change flavor options", map[string]interface{}{"change_opts": changeOpts})

		_, err := cfg.Client.Post(ctx, client.ApiPath.LoadBalancerChangeFlavor(cfg.ProjectID, d.Id()), changeOpts, nil, nil)
		if err != nil {
			return diag.Errorf("Error changing flavor of vnpaycloud_lb_loadbalancer %s: %s", d.Id(), err)
		}

		stateConf := &retry.StateChangeConf{
			Pending:    []string{"pending_update", "creating"},
			Target:     []string{"active", "created"},
			Refresh:    loadBalancerStateRefreshFunc(ctx, cfg.Client, cfg.ProjectID, d.Id()),
			Timeout:    d.Timeout(schema.TimeoutUpdate),
			Delay:      5 * time.Second,
			MinTimeout: 3 * time.Second,
		}

		if _, err := stateConf.WaitForStateContext(ctx); err != nil {
			return diag.Errorf("Error waiting for vnpaycloud_lb_loadbalancer %s to converge after flavor change: %s", d.Id(), err)
		}

		wantFlavor := d.Get("flavor").(string)
		flavorErr := retry.RetryContext(ctx, d.Timeout(schema.TimeoutUpdate), func() *retry.RetryError {
			lbResp := &dto.LoadBalancerResponse{}
			if _, e := cfg.Client.Get(ctx, client.ApiPath.LoadBalancerWithID(cfg.ProjectID, d.Id()), lbResp, nil); e != nil {
				return retry.NonRetryableError(e)
			}
			if !strings.EqualFold(lbResp.LoadBalancer.Flavor, wantFlavor) {
				return retry.RetryableError(fmt.Errorf("flavor not yet propagated in read model: have %q, want %q", lbResp.LoadBalancer.Flavor, wantFlavor))
			}
			return nil
		})
		if flavorErr != nil {
			return diag.Errorf("Error waiting for vnpaycloud_lb_loadbalancer %s flavor to propagate after change: %s", d.Id(), flavorErr)
		}
	}

	return resourceLoadBalancerRead(ctx, d, meta)
}

func resourceLoadBalancerDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	cfg := meta.(*config.Config)

	deleteErr := retry.RetryContext(ctx, d.Timeout(schema.TimeoutDelete), func() *retry.RetryError {
		_, err := cfg.Client.Delete(ctx, client.ApiPath.LoadBalancerWithID(cfg.ProjectID, d.Id()), nil)
		if err != nil && strings.Contains(err.Error(), "not active") {
			return retry.RetryableError(err)
		}
		if err != nil {
			return retry.NonRetryableError(err)
		}
		return nil
	})
	if deleteErr != nil {
		return diag.FromErr(util.CheckDeleted(d, deleteErr, "Error deleting vnpaycloud_lb_loadbalancer"))
	}

	stateConf := &retry.StateChangeConf{
		Pending:    []string{"deleting", "active", "created", "pending_delete", "unknown"},
		Target:     []string{"deleted"},
		Refresh:    loadBalancerStateRefreshFunc(ctx, cfg.Client, cfg.ProjectID, d.Id()),
		Timeout:    d.Timeout(schema.TimeoutDelete),
		Delay:      5 * time.Second,
		MinTimeout: 3 * time.Second,
	}

	_, err := stateConf.WaitForStateContext(ctx)
	if err != nil {
		return diag.Errorf("Error waiting for vnpaycloud_lb_loadbalancer %s to delete: %s", d.Id(), err)
	}

	return nil
}
