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
			"name": {
				Type:     schema.TypeString,
				Computed: true,
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
			"max_retries_down": {
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
	d.Set("name", resp.HealthMonitor.Name)
	d.Set("pool_id", resp.HealthMonitor.PoolID)
	d.Set("type", resp.HealthMonitor.Type)
	d.Set("delay", resp.HealthMonitor.Delay)
	d.Set("timeout", resp.HealthMonitor.Timeout)
	d.Set("max_retries", resp.HealthMonitor.MaxRetries)
	d.Set("max_retries_down", resp.HealthMonitor.MaxRetriesDown)
	d.Set("http_method", resp.HealthMonitor.HTTPMethod)
	d.Set("url_path", resp.HealthMonitor.URLPath)
	d.Set("expected_codes", resp.HealthMonitor.ExpectedCodes)
	d.Set("status", resp.HealthMonitor.Status)

	return nil
}

func DataSourceHealthMonitors() *schema.Resource {
	return &schema.Resource{
		ReadContext: dataSourceHealthMonitorsRead,
		Schema: map[string]*schema.Schema{
			"health_monitors": {
				Type:     schema.TypeList,
				Computed: true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"id":               {Type: schema.TypeString, Computed: true},
						"name":             {Type: schema.TypeString, Computed: true},
						"pool_id":          {Type: schema.TypeString, Computed: true},
						"type":             {Type: schema.TypeString, Computed: true},
						"delay":            {Type: schema.TypeInt, Computed: true},
						"timeout":          {Type: schema.TypeInt, Computed: true},
						"max_retries":      {Type: schema.TypeInt, Computed: true},
						"max_retries_down": {Type: schema.TypeInt, Computed: true},
						"http_method":      {Type: schema.TypeString, Computed: true},
						"url_path":         {Type: schema.TypeString, Computed: true},
						"expected_codes":   {Type: schema.TypeString, Computed: true},
						"status":           {Type: schema.TypeString, Computed: true},
					},
				},
			},
		},
	}
}

func dataSourceHealthMonitorsRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	cfg := meta.(*config.Config)

	resp := &dto.ListHealthMonitorsResponse{}
	_, err := cfg.Client.Get(ctx, client.ApiPath.HealthMonitors(cfg.ProjectID), resp, nil)
	if err != nil {
		return diag.Errorf("Error listing vnpaycloud_lb_health_monitors: %s", err)
	}

	var monitors []map[string]interface{}
	for _, m := range resp.HealthMonitors {
		monitors = append(monitors, map[string]interface{}{
			"id":               m.ID,
			"name":             m.Name,
			"pool_id":          m.PoolID,
			"type":             m.Type,
			"delay":            m.Delay,
			"timeout":          m.Timeout,
			"max_retries":      m.MaxRetries,
			"max_retries_down": m.MaxRetriesDown,
			"http_method":      m.HTTPMethod,
			"url_path":         m.URLPath,
			"expected_codes":   m.ExpectedCodes,
			"status":           m.Status,
		})
	}

	d.SetId("health_monitors")
	d.Set("health_monitors", monitors)

	return nil
}
