package loadbalancer

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

func ResourceLoadBalancer() *schema.Resource {
	return &schema.Resource{
		CreateContext: resourceLoadBalancerCreate,
		ReadContext:   resourceLoadBalancerRead,
		UpdateContext: resourceLoadBalancerUpdate,
		DeleteContext: resourceLoadBalancerDelete,
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
			"flavor": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},
			"floating_ip_id": {
				Type:     schema.TypeString,
				Optional: true,
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
			"listener_ids": {
				Type:     schema.TypeList,
				Computed: true,
				Elem:     &schema.Schema{Type: schema.TypeString},
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
	d.Set("vip_address", lbResp.LoadBalancer.VipAddress)
	d.Set("vip_subnet_id", lbResp.LoadBalancer.VipSubnetID)
	d.Set("status", lbResp.LoadBalancer.Status)
	d.Set("listener_ids", lbResp.LoadBalancer.ListenerIDs)
	d.Set("created_at", lbResp.LoadBalancer.CreatedAt)

	return nil
}

func resourceLoadBalancerUpdate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	cfg := meta.(*config.Config)

	if d.HasChanges("name", "description") {
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
			return diag.Errorf("Error waiting for vnpaycloud_lb_loadbalancer %s to become ready: %s", d.Id(), err)
		}
	}

	return resourceLoadBalancerRead(ctx, d, meta)
}

func resourceLoadBalancerDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	cfg := meta.(*config.Config)

	if _, err := cfg.Client.Delete(ctx, client.ApiPath.LoadBalancerWithID(cfg.ProjectID, d.Id()), nil); err != nil {
		return diag.FromErr(util.CheckDeleted(d, err, "Error deleting vnpaycloud_lb_loadbalancer"))
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
