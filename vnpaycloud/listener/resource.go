package listener

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

func ResourceListener() *schema.Resource {
	return &schema.Resource{
		CreateContext: resourceListenerCreate,
		ReadContext:   resourceListenerRead,
		UpdateContext: resourceListenerUpdate,
		DeleteContext: resourceListenerDelete,
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
			"load_balancer_id": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},
			"protocol": {
				Type:         schema.TypeString,
				Required:     true,
				ForceNew:     true,
				ValidateFunc: validation.StringInSlice([]string{"HTTP", "HTTPS", "TCP", "UDP"}, false),
			},
			"protocol_port": {
				Type:         schema.TypeInt,
				Required:     true,
				ForceNew:     true,
				ValidateFunc: validation.IntBetween(1, 65535),
			},
			"default_pool_id": {
				Type:     schema.TypeString,
				Optional: true,
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

func resourceListenerCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	cfg := meta.(*config.Config)

	createOpts := dto.CreateListenerRequest{
		Name:           d.Get("name").(string),
		LoadBalancerID: d.Get("load_balancer_id").(string),
		Protocol:       d.Get("protocol").(string),
		ProtocolPort:   d.Get("protocol_port").(int),
	}

	if v, ok := d.GetOk("default_pool_id"); ok {
		createOpts.DefaultPoolID = v.(string)
	}

	tflog.Debug(ctx, "vnpaycloud_lb_listener create options", map[string]interface{}{"create_opts": createOpts})

	createResp := &dto.ListenerResponse{}
	_, err := cfg.Client.Post(ctx, client.ApiPath.Listeners(cfg.ProjectID), createOpts, createResp, nil)
	if err != nil {
		return diag.Errorf("Error creating vnpaycloud_lb_listener: %s", err)
	}

	d.SetId(createResp.Listener.ID)

	stateConf := &retry.StateChangeConf{
		Pending:    []string{"initiating", "creating", "pending_create"},
		Target:     []string{"active", "created"},
		Refresh:    listenerStateRefreshFunc(ctx, cfg.Client, cfg.ProjectID, createResp.Listener.ID),
		Timeout:    d.Timeout(schema.TimeoutCreate),
		Delay:      5 * time.Second,
		MinTimeout: 3 * time.Second,
	}

	_, err = stateConf.WaitForStateContext(ctx)
	if err != nil {
		return diag.Errorf("Error waiting for vnpaycloud_lb_listener %s to become ready: %s", createResp.Listener.ID, err)
	}

	return resourceListenerRead(ctx, d, meta)
}

func resourceListenerRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	cfg := meta.(*config.Config)

	resp := &dto.ListenerResponse{}
	_, err := cfg.Client.Get(ctx, client.ApiPath.ListenerWithID(cfg.ProjectID, d.Id()), resp, nil)
	if err != nil {
		return diag.FromErr(util.CheckNotFound(d, err, "Error retrieving vnpaycloud_lb_listener"))
	}

	tflog.Debug(ctx, "Retrieved vnpaycloud_lb_listener "+d.Id(), map[string]interface{}{"listener": resp.Listener})

	d.Set("name", resp.Listener.Name)
	d.Set("load_balancer_id", resp.Listener.LoadBalancerID)
	d.Set("protocol", resp.Listener.Protocol)
	d.Set("protocol_port", resp.Listener.ProtocolPort)
	d.Set("default_pool_id", resp.Listener.DefaultPoolID)
	d.Set("status", resp.Listener.Status)
	d.Set("created_at", resp.Listener.CreatedAt)

	return nil
}

func resourceListenerUpdate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	cfg := meta.(*config.Config)

	if d.HasChanges("name", "default_pool_id") {
		updateOpts := dto.UpdateListenerRequest{
			Name:          d.Get("name").(string),
			DefaultPoolID: d.Get("default_pool_id").(string),
		}

		tflog.Debug(ctx, "vnpaycloud_lb_listener update options", map[string]interface{}{"update_opts": updateOpts})

		_, err := cfg.Client.Put(ctx, client.ApiPath.ListenerWithID(cfg.ProjectID, d.Id()), updateOpts, nil, nil)
		if err != nil {
			return diag.Errorf("Error updating vnpaycloud_lb_listener %s: %s", d.Id(), err)
		}
	}

	return resourceListenerRead(ctx, d, meta)
}

func resourceListenerDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	cfg := meta.(*config.Config)

	if _, err := cfg.Client.Delete(ctx, client.ApiPath.ListenerWithID(cfg.ProjectID, d.Id()), nil); err != nil {
		return diag.FromErr(util.CheckDeleted(d, err, "Error deleting vnpaycloud_lb_listener"))
	}

	stateConf := &retry.StateChangeConf{
		Pending:    []string{"deleting", "active", "created", "pending_delete"},
		Target:     []string{"deleted"},
		Refresh:    listenerStateRefreshFunc(ctx, cfg.Client, cfg.ProjectID, d.Id()),
		Timeout:    d.Timeout(schema.TimeoutDelete),
		Delay:      5 * time.Second,
		MinTimeout: 3 * time.Second,
	}

	_, err := stateConf.WaitForStateContext(ctx)
	if err != nil {
		return diag.Errorf("Error waiting for vnpaycloud_lb_listener %s to delete: %s", d.Id(), err)
	}

	return nil
}
