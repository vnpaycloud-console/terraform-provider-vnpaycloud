package listener

import (
	"context"
	"terraform-provider-vnpaycloud/vnpaycloud/config"
	"terraform-provider-vnpaycloud/vnpaycloud/dto"
	"terraform-provider-vnpaycloud/vnpaycloud/helper/client"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

func DataSourceListener() *schema.Resource {
	return &schema.Resource{
		ReadContext: dataSourceListenerRead,
		Schema: map[string]*schema.Schema{
			"id": {
				Type:     schema.TypeString,
				Required: true,
			},
			"name": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"load_balancer_id": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"protocol": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"protocol_port": {
				Type:     schema.TypeInt,
				Computed: true,
			},
			"default_pool_id": {
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

func dataSourceListenerRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	cfg := meta.(*config.Config)

	id := d.Get("id").(string)

	resp := &dto.ListenerResponse{}
	_, err := cfg.Client.Get(ctx, client.ApiPath.ListenerWithID(cfg.ProjectID, id), resp, nil)
	if err != nil {
		return diag.Errorf("Error retrieving vnpaycloud_lb_listener %s: %s", id, err)
	}

	d.SetId(resp.Listener.ID)
	d.Set("name", resp.Listener.Name)
	d.Set("load_balancer_id", resp.Listener.LoadBalancerID)
	d.Set("protocol", resp.Listener.Protocol)
	d.Set("protocol_port", resp.Listener.ProtocolPort)
	d.Set("default_pool_id", resp.Listener.DefaultPoolID)
	d.Set("status", resp.Listener.Status)
	d.Set("created_at", resp.Listener.CreatedAt)

	return nil
}
