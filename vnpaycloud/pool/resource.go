package pool

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

func ResourcePool() *schema.Resource {
	return &schema.Resource{
		CreateContext: resourcePoolCreate,
		ReadContext:   resourcePoolRead,
		UpdateContext: resourcePoolUpdate,
		DeleteContext: resourcePoolDelete,
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
			"listener_id": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},
			"lb_algorithm": {
				Type:         schema.TypeString,
				Required:     true,
				ValidateFunc: validation.StringInSlice([]string{"ROUND_ROBIN", "LEAST_CONNECTIONS", "SOURCE_IP"}, false),
			},
			"protocol": {
				Type:         schema.TypeString,
				Required:     true,
				ForceNew:     true,
				ValidateFunc: validation.StringInSlice([]string{"HTTP", "HTTPS", "TCP", "UDP", "PROXY"}, false),
			},
			"member": {
				Type:     schema.TypeList,
				Optional: true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"id": {
							Type:     schema.TypeString,
							Computed: true,
						},
						"name": {
							Type:     schema.TypeString,
							Computed: true,
						},
						"address": {
							Type:     schema.TypeString,
							Required: true,
						},
						"protocol_port": {
							Type:         schema.TypeInt,
							Required:     true,
							ValidateFunc: validation.IntBetween(1, 65535),
						},
						"weight": {
							Type:         schema.TypeInt,
							Optional:     true,
							Default:      1,
							ValidateFunc: validation.IntBetween(0, 256),
						},
						"status": {
							Type:     schema.TypeString,
							Computed: true,
						},
					},
				},
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

func expandPoolMembers(raw []interface{}) []dto.PoolMember {
	members := make([]dto.PoolMember, len(raw))
	for i, v := range raw {
		m := v.(map[string]interface{})
		members[i] = dto.PoolMember{
			Address:      m["address"].(string),
			ProtocolPort: m["protocol_port"].(int),
			Weight:       m["weight"].(int),
		}
	}
	return members
}

func flattenPoolMembers(members []dto.PoolMember) []map[string]interface{} {
	result := make([]map[string]interface{}, len(members))
	for i, m := range members {
		result[i] = map[string]interface{}{
			"id":            m.ID,
			"name":          m.Name,
			"address":       m.Address,
			"protocol_port": m.ProtocolPort,
			"weight":        m.Weight,
			"status":        m.Status,
		}
	}
	return result
}

func resourcePoolCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	cfg := meta.(*config.Config)

	// Create pool without members — the API does not support inline members on create.
	createOpts := dto.CreatePoolRequest{
		Name:        d.Get("name").(string),
		ListenerID:  d.Get("listener_id").(string),
		LBAlgorithm: d.Get("lb_algorithm").(string),
		Protocol:    d.Get("protocol").(string),
	}

	tflog.Debug(ctx, "vnpaycloud_lb_pool create options", map[string]interface{}{"create_opts": createOpts})

	createResp := &dto.PoolResponse{}
	_, err := cfg.Client.Post(ctx, client.ApiPath.Pools(cfg.ProjectID), createOpts, createResp, nil)
	if err != nil {
		return diag.Errorf("Error creating vnpaycloud_lb_pool: %s", err)
	}

	d.SetId(createResp.Pool.ID)

	stateConf := &retry.StateChangeConf{
		Pending:    []string{"initiating", "creating", "pending_create"},
		Target:     []string{"active", "created"},
		Refresh:    poolStateRefreshFunc(ctx, cfg.Client, cfg.ProjectID, createResp.Pool.ID),
		Timeout:    d.Timeout(schema.TimeoutCreate),
		Delay:      5 * time.Second,
		MinTimeout: 3 * time.Second,
	}

	_, err = stateConf.WaitForStateContext(ctx)
	if err != nil {
		return diag.Errorf("Error waiting for vnpaycloud_lb_pool %s to become ready: %s", createResp.Pool.ID, err)
	}

	// Add members via PUT update after pool is active.
	// The LB may still be provisioning after pool creation, so retry with backoff.
	if v, ok := d.GetOk("member"); ok {
		updateOpts := dto.UpdatePoolRequest{
			Name:        d.Get("name").(string),
			LBAlgorithm: d.Get("lb_algorithm").(string),
			Members:     expandPoolMembers(v.([]interface{})),
		}

		tflog.Debug(ctx, "vnpaycloud_lb_pool adding members via update", map[string]interface{}{"update_opts": updateOpts})

		err := retry.RetryContext(ctx, 2*time.Minute, func() *retry.RetryError {
			_, putErr := cfg.Client.Put(ctx, client.ApiPath.PoolWithID(cfg.ProjectID, d.Id()), updateOpts, nil, nil)
			if putErr != nil {
				tflog.Warn(ctx, "vnpaycloud_lb_pool member update not ready, retrying", map[string]interface{}{"error": putErr.Error()})
				return retry.RetryableError(putErr)
			}
			return nil
		})
		if err != nil {
			return diag.Errorf("Error adding members to vnpaycloud_lb_pool %s: %s", d.Id(), err)
		}
	}

	return resourcePoolRead(ctx, d, meta)
}

func resourcePoolRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	cfg := meta.(*config.Config)

	resp := &dto.PoolResponse{}
	_, err := cfg.Client.Get(ctx, client.ApiPath.PoolWithID(cfg.ProjectID, d.Id()), resp, nil)
	if err != nil {
		return diag.FromErr(util.CheckNotFound(d, err, "Error retrieving vnpaycloud_lb_pool"))
	}

	tflog.Debug(ctx, "Retrieved vnpaycloud_lb_pool "+d.Id(), map[string]interface{}{"pool": resp.Pool})

	d.Set("name", resp.Pool.Name)
	d.Set("listener_id", resp.Pool.ListenerID)
	d.Set("lb_algorithm", resp.Pool.LBAlgorithm)
	d.Set("protocol", resp.Pool.Protocol)
	d.Set("member", flattenPoolMembers(resp.Pool.Members))
	d.Set("status", resp.Pool.Status)
	d.Set("created_at", resp.Pool.CreatedAt)

	return nil
}

func resourcePoolUpdate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	cfg := meta.(*config.Config)

	if d.HasChanges("name", "lb_algorithm", "member") {
		updateOpts := dto.UpdatePoolRequest{
			Name:        d.Get("name").(string),
			LBAlgorithm: d.Get("lb_algorithm").(string),
		}

		if v, ok := d.GetOk("member"); ok {
			updateOpts.Members = expandPoolMembers(v.([]interface{}))
		}

		tflog.Debug(ctx, "vnpaycloud_lb_pool update options", map[string]interface{}{"update_opts": updateOpts})

		_, err := cfg.Client.Put(ctx, client.ApiPath.PoolWithID(cfg.ProjectID, d.Id()), updateOpts, nil, nil)
		if err != nil {
			return diag.Errorf("Error updating vnpaycloud_lb_pool %s: %s", d.Id(), err)
		}
	}

	return resourcePoolRead(ctx, d, meta)
}

func resourcePoolDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	cfg := meta.(*config.Config)

	if _, err := cfg.Client.Delete(ctx, client.ApiPath.PoolWithID(cfg.ProjectID, d.Id()), nil); err != nil {
		return diag.FromErr(util.CheckDeleted(d, err, "Error deleting vnpaycloud_lb_pool"))
	}

	stateConf := &retry.StateChangeConf{
		Pending:    []string{"deleting", "active", "created", "pending_delete"},
		Target:     []string{"deleted"},
		Refresh:    poolStateRefreshFunc(ctx, cfg.Client, cfg.ProjectID, d.Id()),
		Timeout:    d.Timeout(schema.TimeoutDelete),
		Delay:      5 * time.Second,
		MinTimeout: 3 * time.Second,
	}

	_, err := stateConf.WaitForStateContext(ctx)
	if err != nil {
		return diag.Errorf("Error waiting for vnpaycloud_lb_pool %s to delete: %s", d.Id(), err)
	}

	return nil
}
