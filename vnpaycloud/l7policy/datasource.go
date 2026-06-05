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
