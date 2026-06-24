package pool

import (
	"context"
	"fmt"
	"strings"
	"terraform-provider-vnpaycloud/vnpaycloud/config"
	"terraform-provider-vnpaycloud/vnpaycloud/dto"
	"terraform-provider-vnpaycloud/vnpaycloud/helper/client"
	"terraform-provider-vnpaycloud/vnpaycloud/helper/hashcode"
	"terraform-provider-vnpaycloud/vnpaycloud/util"
	"time"

	"github.com/hashicorp/terraform-plugin-log/tflog"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/retry"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

func ResourcePool() *schema.Resource {
	return &schema.Resource{
		CreateContext: resourcePoolCreate,
		ReadContext:   resourcePoolRead,
		UpdateContext: resourcePoolUpdate,
		DeleteContext: resourcePoolDelete,
		Importer: &schema.ResourceImporter{
			StateContext: func(ctx context.Context, d *schema.ResourceData, meta interface{}) ([]*schema.ResourceData, error) {
				cfg := meta.(*config.Config)
				resp := &dto.PoolResponse{}
				if _, err := cfg.Client.Get(ctx, client.ApiPath.PoolWithID(cfg.ProjectID, d.Id()), resp, nil); err != nil {
					return nil, fmt.Errorf("vnpaycloud_lb_pool %q not found: %w", d.Id(), err)
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
			},
			"description": {
				Type:     schema.TypeString,
				Optional: true,
			},
			"load_balancer_id": {
				Type:        schema.TypeString,
				Required:    true,
				ForceNew:    true,
				Description: "ID of the parent load balancer. Pools belong to a load balancer (1 LB → many pools); attaching to a listener is a separate, optional step via `listener_id` below.",
			},
			"listener_id": {
				Type:     schema.TypeString,
				Optional: true,
				Computed: true,
				ForceNew: true,
				Description: "Optional — set only when you want this pool to become the listener's `default_pool_id` at create time. The listener must currently have no default pool (a listener accepts at most one default; swap an existing default via `vnpaycloud_lb_listener.default_pool_id` instead). " +
					"`Computed`: if the listener attaches this pool as its default out-of-band (e.g. you set `vnpaycloud_lb_listener.default_pool_id`), the backend writes this back-pointer and Terraform keeps it in state instead of forcing a spurious recreate.",
			},
			"lb_algorithm": {
				Type:     schema.TypeString,
				Required: true,
			},
			"protocol": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},
			"session_persistence": {
				Type:     schema.TypeList,
				Optional: true,
				MaxItems: 1,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"type": {
							Type:     schema.TypeString,
							Required: true,
						},
						"cookie_name": {
							Type:     schema.TypeString,
							Optional: true,
						},
					},
				},
			},
			"tls_enabled": {
				Type:     schema.TypeBool,
				Optional: true,
				Computed: true,
			},
			"member": {
				Type:     schema.TypeSet,
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
							Type:     schema.TypeInt,
							Required: true,
						},
						"weight": {
							Type:     schema.TypeInt,
							Optional: true,
							Default:  1,
						},
						"status": {
							Type:     schema.TypeString,
							Computed: true,
						},
					},
				},
				Set: func(v interface{}) int {
					m := v.(map[string]interface{})
					return hashcode.String(fmt.Sprintf("%s:%d:%d", m["address"].(string), m["protocol_port"].(int), m["weight"].(int)))
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

func expandSessionPersistence(raw interface{}) *dto.SessionPersistence {
	list, ok := raw.([]interface{})
	if !ok || len(list) == 0 {
		return nil
	}
	m := list[0].(map[string]interface{})
	return &dto.SessionPersistence{
		Type:       m["type"].(string),
		CookieName: m["cookie_name"].(string),
	}
}

func flattenSessionPersistence(sp *dto.SessionPersistence) []map[string]interface{} {
	if sp == nil || sp.Type == "" {
		return nil
	}
	return []map[string]interface{}{
		{
			"type":        sp.Type,
			"cookie_name": sp.CookieName,
		},
	}
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

	listenerID := d.Get("listener_id").(string)

	createOpts := dto.CreatePoolRequest{
		Name:               d.Get("name").(string),
		Description:        d.Get("description").(string),
		LoadBalancerID:     d.Get("load_balancer_id").(string),
		ListenerID:         listenerID,
		LBAlgorithm:        d.Get("lb_algorithm").(string),
		Protocol:           d.Get("protocol").(string),
		TlsEnabled:         d.Get("tls_enabled").(bool),
		SessionPersistence: expandSessionPersistence(d.Get("session_persistence")),
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

	if listenerID != "" {
		listenerResp := &dto.ListenerResponse{}
		if _, err := cfg.Client.Get(ctx, client.ApiPath.ListenerWithID(cfg.ProjectID, listenerID), listenerResp, nil); err != nil {
			return diag.Errorf("Error fetching listener %s for default_pool_id sync: %s", listenerID, err)
		}
		listenerUpdate := dto.UpdateListenerRequest{
			Name:                   listenerResp.Listener.Name,
			Description:            listenerResp.Listener.Description,
			DefaultPoolID:          d.Id(),
			InsertHeaders:          listenerResp.Listener.InsertHeaders,
			AllowedCidrs:           listenerResp.Listener.AllowedCidrs,
			ConnectionLimit:        listenerResp.Listener.ConnectionLimit,
			TimeoutClientData:      listenerResp.Listener.TimeoutClientData,
			TimeoutMemberConnect:   listenerResp.Listener.TimeoutMemberConnect,
			TimeoutMemberData:      listenerResp.Listener.TimeoutMemberData,
			CertificateID:          listenerResp.Listener.CertificateID,
			CertificateAuthorityID: listenerResp.Listener.CertificateAuthorityID,
			SniCertificateIDs:      listenerResp.Listener.SniCertificateIDs,
		}
		err := util.RetryLBPendingPut(ctx, d.Timeout(schema.TimeoutCreate), func() error {
			_, putErr := cfg.Client.Put(ctx, client.ApiPath.ListenerWithID(cfg.ProjectID, listenerID), listenerUpdate, nil, nil)
			return putErr
		})
		if err != nil {
			return diag.Errorf("Error setting default_pool_id=%s on listener %s: %s", d.Id(), listenerID, err)
		}
	}

	if v, ok := d.GetOk("member"); ok {
		updateOpts := dto.UpdatePoolRequest{
			Name:               d.Get("name").(string),
			Description:        d.Get("description").(string),
			LBAlgorithm:        d.Get("lb_algorithm").(string),
			TlsEnabled:         d.Get("tls_enabled").(bool),
			SessionPersistence: expandSessionPersistence(d.Get("session_persistence")),
			Members:            expandPoolMembers(v.(*schema.Set).List()),
		}

		tflog.Debug(ctx, "vnpaycloud_lb_pool adding members via update", map[string]interface{}{"update_opts": updateOpts})

		err := util.RetryLBPendingPut(ctx, d.Timeout(schema.TimeoutCreate), func() error {
			_, putErr := cfg.Client.Put(ctx, client.ApiPath.PoolWithID(cfg.ProjectID, d.Id()), updateOpts, nil, nil)
			return putErr
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
	d.Set("description", resp.Pool.Description)
	d.Set("load_balancer_id", resp.Pool.LoadBalancerID)
	if resp.Pool.ListenerID != "" {
		d.Set("listener_id", resp.Pool.ListenerID)
	}
	d.Set("lb_algorithm", resp.Pool.LBAlgorithm)
	d.Set("protocol", resp.Pool.Protocol)
	d.Set("session_persistence", flattenSessionPersistence(resp.Pool.SessionPersistence))
	d.Set("tls_enabled", resp.Pool.TlsEnabled)
	d.Set("member", flattenPoolMembers(resp.Pool.Members))
	d.Set("status", resp.Pool.Status)
	d.Set("created_at", resp.Pool.CreatedAt)

	return nil
}

func resourcePoolUpdate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	cfg := meta.(*config.Config)

	if d.HasChanges("name", "description", "lb_algorithm", "session_persistence", "tls_enabled", "member") {
		waitBefore := &retry.StateChangeConf{
			Pending:    []string{"initiating", "creating", "pending_create", "pending_update"},
			Target:     []string{"active", "created"},
			Refresh:    poolStateRefreshFunc(ctx, cfg.Client, cfg.ProjectID, d.Id()),
			Timeout:    d.Timeout(schema.TimeoutUpdate),
			Delay:      3 * time.Second,
			MinTimeout: 3 * time.Second,
		}
		if _, err := waitBefore.WaitForStateContext(ctx); err != nil {
			return diag.Errorf("Error waiting for vnpaycloud_lb_pool %s to become ready before update: %s", d.Id(), err)
		}

		updateOpts := dto.UpdatePoolRequest{
			Name:               d.Get("name").(string),
			Description:        d.Get("description").(string),
			LBAlgorithm:        d.Get("lb_algorithm").(string),
			TlsEnabled:         d.Get("tls_enabled").(bool),
			SessionPersistence: expandSessionPersistence(d.Get("session_persistence")),
		}

		if v, ok := d.GetOk("member"); ok {
			updateOpts.Members = expandPoolMembers(v.(*schema.Set).List())
		}

		tflog.Debug(ctx, "vnpaycloud_lb_pool update options", map[string]interface{}{"update_opts": updateOpts})

		err := util.RetryLBPendingPut(ctx, d.Timeout(schema.TimeoutUpdate), func() error {
			_, putErr := cfg.Client.Put(ctx, client.ApiPath.PoolWithID(cfg.ProjectID, d.Id()), updateOpts, nil, nil)
			return putErr
		})
		if err != nil {
			return diag.Errorf("Error updating vnpaycloud_lb_pool %s: %s", d.Id(), err)
		}

		waitAfter := &retry.StateChangeConf{
			Pending:    []string{"pending_update", "creating"},
			Target:     []string{"active", "created"},
			Refresh:    poolStateRefreshFunc(ctx, cfg.Client, cfg.ProjectID, d.Id()),
			Timeout:    d.Timeout(schema.TimeoutUpdate),
			Delay:      3 * time.Second,
			MinTimeout: 3 * time.Second,
		}
		if _, err := waitAfter.WaitForStateContext(ctx); err != nil {
			return diag.Errorf("Error waiting for vnpaycloud_lb_pool %s to converge after update: %s", d.Id(), err)
		}
	}

	return resourcePoolRead(ctx, d, meta)
}

func resourcePoolDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	cfg := meta.(*config.Config)

	deleteErr := retry.RetryContext(ctx, d.Timeout(schema.TimeoutDelete), func() *retry.RetryError {
		_, err := cfg.Client.Delete(ctx, client.ApiPath.PoolWithID(cfg.ProjectID, d.Id()), nil)
		if err != nil && strings.Contains(err.Error(), "not active") {
			return retry.RetryableError(err)
		}
		if err != nil {
			return retry.NonRetryableError(err)
		}
		return nil
	})
	if deleteErr != nil {
		return diag.FromErr(util.CheckDeleted(d, deleteErr, "Error deleting vnpaycloud_lb_pool"))
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
