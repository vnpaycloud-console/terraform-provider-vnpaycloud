package maaspipelineexecutor

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

func DataSourceMaasPipelineExecutor() *schema.Resource {
	return &schema.Resource{
		ReadContext: dataSourceMaasPipelineExecutorRead,
		Schema: map[string]*schema.Schema{
			"id": {
				Type:     schema.TypeString,
				Required: true,
			},
			"pipeline_id":   {Type: schema.TypeString, Computed: true},
			"order":         {Type: schema.TypeInt, Computed: true},
			"description":   {Type: schema.TypeString, Computed: true},
			"type":          {Type: schema.TypeString, Computed: true},
			"executor_type": {Type: schema.TypeString, Computed: true},
			"static_labels": {Type: schema.TypeMap, Computed: true, Elem: &schema.Schema{Type: schema.TypeString}},
			"rename_fields": {Type: schema.TypeMap, Computed: true, Elem: &schema.Schema{Type: schema.TypeString}},
			"rename_labels": {Type: schema.TypeMap, Computed: true, Elem: &schema.Schema{Type: schema.TypeString}},
			"status":        {Type: schema.TypeString, Computed: true},
			"created_at":    {Type: schema.TypeString, Computed: true},
		},
	}
}

func dataSourceMaasPipelineExecutorRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	cfg := meta.(*config.Config)

	id := d.Get("id").(string)
	resp := &dto.MaasPipelineExecutorResponse{}
	if _, err := cfg.Client.Get(ctx, client.ApiPath.MaasPipelineExecutorWithID(cfg.ProjectID, id), resp, nil); err != nil {
		return diag.Errorf("Error fetching vnpaycloud_maas_pipeline_executor %s: %s", id, err)
	}

	d.SetId(resp.PipelineExecutor.ID)
	setExecutorData(d, &resp.PipelineExecutor)

	return nil
}

func DataSourceMaasPipelineExecutors() *schema.Resource {
	return &schema.Resource{
		ReadContext: dataSourceMaasPipelineExecutorsRead,
		Schema: map[string]*schema.Schema{
			"pipeline_id": {
				Type:     schema.TypeString,
				Optional: true,
			},
			"type": {
				Type:         schema.TypeString,
				Optional:     true,
				ValidateFunc: validation.StringInSlice([]string{"log", "metric"}, false),
			},
			"executor_type": {
				Type:     schema.TypeString,
				Optional: true,
				ValidateFunc: validation.StringInSlice(
					[]string{executorTypeStaticLabels, executorTypeRenameFields, executorTypeRenameLabels}, false),
			},
			"pipeline_executors": {
				Type:     schema.TypeList,
				Computed: true,
				Elem: &schema.Resource{
					Schema: executorComputedSchema(),
				},
			},
		},
	}
}

func dataSourceMaasPipelineExecutorsRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	cfg := meta.(*config.Config)

	path := client.ApiPath.MaasPipelineExecutors(cfg.ProjectID)
	q := url.Values{}
	if v, ok := d.GetOk("pipeline_id"); ok {
		q.Set("pipelineId", v.(string))
	}
	if v, ok := d.GetOk("type"); ok {
		q.Set("type", v.(string))
	}
	if v, ok := d.GetOk("executor_type"); ok {
		q.Set("executorType", v.(string))
	}
	if len(q) > 0 {
		path += "?" + q.Encode()
	}

	listResp := &dto.ListMaasPipelineExecutorsResponse{}
	if _, err := cfg.Client.Get(ctx, path, listResp, nil); err != nil {
		return diag.Errorf("Unable to query vnpaycloud_maas_pipeline_executors: %s", err)
	}

	executors := make([]map[string]interface{}, 0, len(listResp.PipelineExecutors))
	for _, e := range listResp.PipelineExecutors {
		executors = append(executors, map[string]interface{}{
			"id":            e.ID,
			"pipeline_id":   e.PipelineID,
			"order":         e.Order,
			"description":   e.Description,
			"type":          e.Type,
			"executor_type": e.ExecutorType,
			"static_labels": flattenStringMap(e.StaticLabels),
			"rename_fields": flattenStringMap(e.RenameFields),
			"rename_labels": flattenStringMap(e.RenameLabels),
			"status":        e.Status,
			"created_at":    e.CreatedAt,
		})
	}

	d.SetId(cfg.ProjectID)
	d.Set("pipeline_executors", executors)

	return nil
}
