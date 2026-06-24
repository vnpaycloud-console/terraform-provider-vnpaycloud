package l7policy

import (
	"context"
	"terraform-provider-vnpaycloud/vnpaycloud/config"
	"terraform-provider-vnpaycloud/vnpaycloud/dto"
	"terraform-provider-vnpaycloud/vnpaycloud/helper/client"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

func DataSourceL7Policy() *schema.Resource {
	return &schema.Resource{
		ReadContext: dataSourceL7PolicyRead,
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
			"listener_id": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"action": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"position": {
				Type:     schema.TypeInt,
				Computed: true,
			},
			"redirect_pool_id": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"redirect_url": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"status": {
				Type:     schema.TypeString,
				Computed: true,
			},
		},
	}
}

func dataSourceL7PolicyRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	cfg := meta.(*config.Config)
	id := d.Get("id").(string)

	resp := &dto.L7PolicyResponse{}
	_, err := cfg.Client.Get(ctx, client.ApiPath.L7PolicyWithID(cfg.ProjectID, id), resp, nil)
	if err != nil {
		return diag.Errorf("Error retrieving vnpaycloud_lb_l7policy %s: %s", id, err)
	}

	d.SetId(resp.L7Policy.ID)
	d.Set("name", resp.L7Policy.Name)
	d.Set("description", resp.L7Policy.Description)
	d.Set("listener_id", resp.L7Policy.ListenerID)
	d.Set("action", resp.L7Policy.Action)
	d.Set("position", resp.L7Policy.Position)
	d.Set("redirect_pool_id", resp.L7Policy.RedirectPoolID)
	d.Set("redirect_url", resp.L7Policy.RedirectURL)
	d.Set("status", resp.L7Policy.Status)

	return nil
}

func DataSourceL7Policies() *schema.Resource {
	return &schema.Resource{
		ReadContext: dataSourceL7PoliciesRead,
		Schema: map[string]*schema.Schema{
			"l7policies": {
				Type:     schema.TypeList,
				Computed: true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"id":               {Type: schema.TypeString, Computed: true},
						"name":             {Type: schema.TypeString, Computed: true},
						"description":      {Type: schema.TypeString, Computed: true},
						"listener_id":      {Type: schema.TypeString, Computed: true},
						"action":           {Type: schema.TypeString, Computed: true},
						"position":         {Type: schema.TypeInt, Computed: true},
						"redirect_pool_id": {Type: schema.TypeString, Computed: true},
						"redirect_url":     {Type: schema.TypeString, Computed: true},
						"status":           {Type: schema.TypeString, Computed: true},
					},
				},
			},
		},
	}
}

func dataSourceL7PoliciesRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	cfg := meta.(*config.Config)

	resp := &dto.ListL7PoliciesResponse{}
	_, err := cfg.Client.Get(ctx, client.ApiPath.L7Policies(cfg.ProjectID), resp, nil)
	if err != nil {
		return diag.Errorf("Error listing vnpaycloud_lb_l7policies: %s", err)
	}

	var policies []map[string]interface{}
	for _, p := range resp.L7Policies {
		policies = append(policies, map[string]interface{}{
			"id":               p.ID,
			"name":             p.Name,
			"description":      p.Description,
			"listener_id":      p.ListenerID,
			"action":           p.Action,
			"position":         p.Position,
			"redirect_pool_id": p.RedirectPoolID,
			"redirect_url":     p.RedirectURL,
			"status":           p.Status,
		})
	}

	d.SetId("l7policies")
	d.Set("l7policies", policies)

	return nil
}
