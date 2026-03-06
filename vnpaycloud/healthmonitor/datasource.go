package healthmonitor

import (
	"context"
	"terraform-provider-vnpaycloud/vnpaycloud/config"
	"terraform-provider-vnpaycloud/vnpaycloud/dto"
	"terraform-provider-vnpaycloud/vnpaycloud/helper/client"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

func DataSourceHealthMonitor() *schema.Resource {
	return &schema.Resource{
		ReadContext: dataSourceHealthMonitorRead,
		Schema: map[string]*schema.Schema{
			"id": {
				Type:     schema.TypeString,
				Required: true,
			},
			"pool_id": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"type": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"delay": {
				Type:     schema.TypeInt,
				Computed: true,
			},
			"timeout": {
				Type:     schema.TypeInt,
				Computed: true,
			},
			"max_retries": {
				Type:     schema.TypeInt,
				Computed: true,
			},
			"http_method": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"url_path": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"expected_codes": {
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

func dataSourceHealthMonitorRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	cfg := meta.(*config.Config)

	id := d.Get("id").(string)

	resp := &dto.HealthMonitorResponse{}
	_, err := cfg.Client.Get(ctx, client.ApiPath.HealthMonitorWithID(cfg.ProjectID, id), resp, nil)
	if err != nil {
		return diag.Errorf("Error retrieving vnpaycloud_lb_health_monitor %s: %s", id, err)
	}

	d.SetId(resp.HealthMonitor.ID)
	d.Set("pool_id", resp.HealthMonitor.PoolID)
	d.Set("type", resp.HealthMonitor.Type)
	d.Set("delay", resp.HealthMonitor.Delay)
	d.Set("timeout", resp.HealthMonitor.Timeout)
	d.Set("max_retries", resp.HealthMonitor.MaxRetries)
	d.Set("http_method", resp.HealthMonitor.HTTPMethod)
	d.Set("url_path", resp.HealthMonitor.URLPath)
	d.Set("expected_codes", resp.HealthMonitor.ExpectedCodes)
	d.Set("status", resp.HealthMonitor.Status)

	return nil
}
