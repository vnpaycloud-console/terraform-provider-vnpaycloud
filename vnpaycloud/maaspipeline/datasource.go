package maaspipeline

import (
	"context"
	"net/url"
	"terraform-provider-vnpaycloud/vnpaycloud/config"
	"terraform-provider-vnpaycloud/vnpaycloud/dto"
	"terraform-provider-vnpaycloud/vnpaycloud/helper/client"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
)

func DataSourceMaasPipeline() *schema.Resource {
	return &schema.Resource{
		ReadContext: dataSourceMaasPipelineRead,
		Schema: map[string]*schema.Schema{
			"id": {
				Type:         schema.TypeString,
				Optional:     true,
				Computed:     true,
				AtLeastOneOf: []string{"id", "name"},
			},
			"name": {
				Type:         schema.TypeString,
				Optional:     true,
				Computed:     true,
				AtLeastOneOf: []string{"id", "name"},
			},
			"type":        {Type: schema.TypeString, Computed: true},
			"description": {Type: schema.TypeString, Computed: true},
			"status":      {Type: schema.TypeString, Computed: true},
			"created_at":  {Type: schema.TypeString, Computed: true},
		},
	}
}

func dataSourceMaasPipelineRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	cfg := meta.(*config.Config)

	if id, ok := d.GetOk("id"); ok && id.(string) != "" {
		resp := &dto.MaasPipelineResponse{}
		if _, err := cfg.Client.Get(ctx, client.ApiPath.MaasPipelineWithID(cfg.ProjectID, id.(string)), resp, nil); err != nil {
			return diag.Errorf("Error fetching vnpaycloud_maas_pipeline %s: %s", id, err)
		}
		if v, ok := d.GetOk("name"); ok && resp.Pipeline.Name != v.(string) {
			return diag.Errorf("vnpaycloud_maas_pipeline %q does not match name %q", id.(string), v.(string))
		}
		setPipelineData(d, &resp.Pipeline)
		return nil
	}

	name := d.Get("name").(string)
	listResp := &dto.ListMaasPipelinesResponse{}
	path := client.ApiPath.MaasPipelines(cfg.ProjectID) + "?name=" + url.QueryEscape(name)
	if _, err := cfg.Client.Get(ctx, path, listResp, nil); err != nil {
		return diag.Errorf("Error listing vnpaycloud_maas_pipeline: %s", err)
	}

	var matched []dto.MaasPipeline
	for _, p := range listResp.Pipelines {
		if p.Name == name {
			matched = append(matched, p)
		}
	}

	if len(matched) == 0 {
		return diag.Errorf("No vnpaycloud_maas_pipeline found with name %q", name)
	}
	if len(matched) > 1 {
		return diag.Errorf("Found %d vnpaycloud_maas_pipeline matching name %q; use id to select one", len(matched), name)
	}

	setPipelineData(d, &matched[0])
	return nil
}

func setPipelineData(d *schema.ResourceData, p *dto.MaasPipeline) {
	d.SetId(p.ID)
	d.Set("name", p.Name)
	d.Set("type", p.Type)
	d.Set("description", p.Description)
	d.Set("status", p.Status)
	d.Set("created_at", p.CreatedAt)
}

func DataSourceMaasPipelines() *schema.Resource {
	return &schema.Resource{
		ReadContext: dataSourceMaasPipelinesRead,
		Schema: map[string]*schema.Schema{
			"name": {
				Type:     schema.TypeString,
				Optional: true,
			},
			"type": {
				Type:         schema.TypeString,
				Optional:     true,
				ValidateFunc: validation.StringInSlice([]string{"log", "metric"}, false),
			},
			"pipelines": {
				Type:     schema.TypeList,
				Computed: true,
				Elem: &schema.Resource{
					Schema: pipelineComputedSchema(),
				},
			},
		},
	}
}

func dataSourceMaasPipelinesRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	cfg := meta.(*config.Config)

	path := client.ApiPath.MaasPipelines(cfg.ProjectID)
	q := url.Values{}
	if v, ok := d.GetOk("name"); ok {
		q.Set("name", v.(string))
	}
	if v, ok := d.GetOk("type"); ok {
		q.Set("type", v.(string))
	}
	if len(q) > 0 {
		path += "?" + q.Encode()
	}

	listResp := &dto.ListMaasPipelinesResponse{}
	if _, err := cfg.Client.Get(ctx, path, listResp, nil); err != nil {
		return diag.Errorf("Unable to query vnpaycloud_maas_pipelines: %s", err)
	}

	pipelines := make([]map[string]interface{}, 0, len(listResp.Pipelines))
	for _, p := range listResp.Pipelines {
		pipelines = append(pipelines, map[string]interface{}{
			"id":          p.ID,
			"name":        p.Name,
			"type":        p.Type,
			"description": p.Description,
			"status":      p.Status,
			"created_at":  p.CreatedAt,
		})
	}

	d.SetId(cfg.ProjectID)
	d.Set("pipelines", pipelines)

	return nil
}
