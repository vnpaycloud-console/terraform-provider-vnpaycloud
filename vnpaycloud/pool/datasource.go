package pool

import (
	"context"
	"terraform-provider-vnpaycloud/vnpaycloud/config"
	"terraform-provider-vnpaycloud/vnpaycloud/dto"
	"terraform-provider-vnpaycloud/vnpaycloud/helper/client"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

func DataSourcePool() *schema.Resource {
	return &schema.Resource{
		ReadContext: dataSourcePoolRead,
		Schema: map[string]*schema.Schema{
			"id": {
				Type:     schema.TypeString,
				Required: true,
			},
			"name": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"description": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"load_balancer_id": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"listener_id": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"lb_algorithm": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"protocol": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"session_persistence": {
				Type:     schema.TypeList,
				Computed: true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"type": {
							Type:     schema.TypeString,
							Computed: true,
						},
						"cookie_name": {
							Type:     schema.TypeString,
							Computed: true,
						},
					},
				},
			},
			"tls_enabled": {
				Type:     schema.TypeBool,
				Computed: true,
			},
			"member": {
				Type:     schema.TypeList,
				Computed: true,
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
							Computed: true,
						},
						"protocol_port": {
							Type:     schema.TypeInt,
							Computed: true,
						},
						"weight": {
							Type:     schema.TypeInt,
							Computed: true,
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

func dataSourcePoolRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	cfg := meta.(*config.Config)

	id := d.Get("id").(string)

	resp := &dto.PoolResponse{}
	_, err := cfg.Client.Get(ctx, client.ApiPath.PoolWithID(cfg.ProjectID, id), resp, nil)
	if err != nil {
		return diag.Errorf("Error retrieving vnpaycloud_lb_pool %s: %s", id, err)
	}

	d.SetId(resp.Pool.ID)
	d.Set("name", resp.Pool.Name)
	d.Set("description", resp.Pool.Description)
	d.Set("load_balancer_id", resp.Pool.LoadBalancerID)
	d.Set("listener_id", resp.Pool.ListenerID)
	d.Set("lb_algorithm", resp.Pool.LBAlgorithm)
	d.Set("protocol", resp.Pool.Protocol)
	d.Set("session_persistence", flattenSessionPersistence(resp.Pool.SessionPersistence))
	d.Set("tls_enabled", resp.Pool.TlsEnabled)
	d.Set("member", flattenPoolMembers(resp.Pool.Members))
	d.Set("status", resp.Pool.Status)
	d.Set("created_at", resp.Pool.CreatedAt)

	return nil
}

func DataSourcePools() *schema.Resource {
	memberElem := &schema.Resource{
		Schema: map[string]*schema.Schema{
			"id":            {Type: schema.TypeString, Computed: true},
			"name":          {Type: schema.TypeString, Computed: true},
			"address":       {Type: schema.TypeString, Computed: true},
			"protocol_port": {Type: schema.TypeInt, Computed: true},
			"weight":        {Type: schema.TypeInt, Computed: true},
			"status":        {Type: schema.TypeString, Computed: true},
		},
	}
	spElem := &schema.Resource{
		Schema: map[string]*schema.Schema{
			"type":        {Type: schema.TypeString, Computed: true},
			"cookie_name": {Type: schema.TypeString, Computed: true},
		},
	}
	return &schema.Resource{
		ReadContext: dataSourcePoolsRead,
		Schema: map[string]*schema.Schema{
			"pools": {
				Type:     schema.TypeList,
				Computed: true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"id":                  {Type: schema.TypeString, Computed: true},
						"name":                {Type: schema.TypeString, Computed: true},
						"description":         {Type: schema.TypeString, Computed: true},
						"load_balancer_id":    {Type: schema.TypeString, Computed: true},
						"listener_id":         {Type: schema.TypeString, Computed: true},
						"lb_algorithm":        {Type: schema.TypeString, Computed: true},
						"protocol":            {Type: schema.TypeString, Computed: true},
						"session_persistence": {Type: schema.TypeList, Computed: true, Elem: spElem},
						"tls_enabled":         {Type: schema.TypeBool, Computed: true},
						"member":              {Type: schema.TypeList, Computed: true, Elem: memberElem},
						"status":              {Type: schema.TypeString, Computed: true},
						"created_at":          {Type: schema.TypeString, Computed: true},
					},
				},
			},
		},
	}
}

func dataSourcePoolsRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	cfg := meta.(*config.Config)

	resp := &dto.ListPoolsResponse{}
	_, err := cfg.Client.Get(ctx, client.ApiPath.Pools(cfg.ProjectID), resp, nil)
	if err != nil {
		return diag.Errorf("Error listing vnpaycloud_lb_pools: %s", err)
	}

	var pools []map[string]interface{}
	for _, p := range resp.Pools {
		pools = append(pools, map[string]interface{}{
			"id":                  p.ID,
			"name":                p.Name,
			"description":         p.Description,
			"load_balancer_id":    p.LoadBalancerID,
			"listener_id":         p.ListenerID,
			"lb_algorithm":        p.LBAlgorithm,
			"protocol":            p.Protocol,
			"session_persistence": flattenSessionPersistence(p.SessionPersistence),
			"tls_enabled":         p.TlsEnabled,
			"member":              flattenPoolMembers(p.Members),
			"status":              p.Status,
			"created_at":          p.CreatedAt,
		})
	}

	d.SetId("pools")
	d.Set("pools", pools)

	return nil
}
