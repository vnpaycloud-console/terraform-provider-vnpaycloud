package maasaccesskey

import (
	"terraform-provider-vnpaycloud/vnpaycloud/dto"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

const (
	statusActive   = "active"
	statusInactive = "inactive"
)

var permissions = []string{"none", "read", "write", "read_write"}

var matcherOps = []string{"eq", "neq", "re", "nre"}

func labelMatcherResource() *schema.Resource {
	return &schema.Resource{
		Schema: map[string]*schema.Schema{
			"name":  {Type: schema.TypeString, Required: true},
			"op":    {Type: schema.TypeString, Required: true},
			"value": {Type: schema.TypeString, Required: true},
		},
	}
}

func expandLabelMatchers(raw interface{}) []dto.MaasLabelMatcher {
	set, ok := raw.(*schema.Set)
	if !ok || set.Len() == 0 {
		return nil
	}
	out := make([]dto.MaasLabelMatcher, 0, set.Len())
	for _, item := range set.List() {
		m := item.(map[string]interface{})
		out = append(out, dto.MaasLabelMatcher{
			Name:  m["name"].(string),
			Op:    m["op"].(string),
			Value: m["value"].(string),
		})
	}
	return out
}

func flattenLabelMatchers(ms []dto.MaasLabelMatcher) []map[string]interface{} {
	out := make([]map[string]interface{}, 0, len(ms))
	for _, m := range ms {
		out = append(out, map[string]interface{}{
			"name":  m.Name,
			"op":    m.Op,
			"value": m.Value,
		})
	}
	return out
}

func flattenEndpoints(e *dto.MaasEndpoints) []map[string]interface{} {
	if e == nil {
		return nil
	}
	return []map[string]interface{}{{
		"log_otlp_push":          e.LogOtlpPush,
		"metric_otlp_push":       e.MetricOtlpPush,
		"log_loki_push":          e.LogLokiPush,
		"metric_prometheus_push": e.MetricPrometheusPush,
	}}
}

func endpointsSchema() *schema.Schema {
	return &schema.Schema{
		Type:     schema.TypeList,
		Computed: true,
		Elem: &schema.Resource{
			Schema: map[string]*schema.Schema{
				"log_otlp_push":          {Type: schema.TypeString, Computed: true},
				"metric_otlp_push":       {Type: schema.TypeString, Computed: true},
				"log_loki_push":          {Type: schema.TypeString, Computed: true},
				"metric_prometheus_push": {Type: schema.TypeString, Computed: true},
			},
		},
	}
}

// setAccessKeyData writes everything the read path returns. The credentials are deliberately left
// alone when the response omits them: username and password are handed out only in the create
// response, and api_key stops being returned once the key has been updated.
func setAccessKeyData(d *schema.ResourceData, k *dto.MaasAccessKey) {
	d.Set("name", k.Name)
	d.Set("description", k.Description)
	d.Set("log_permission", k.LogPermission)
	d.Set("metric_permission", k.MetricPermission)
	d.Set("log_pipeline_id", k.LogPipelineID)
	d.Set("metric_pipeline_id", k.MetricPipelineID)
	d.Set("log_label_matcher", flattenLabelMatchers(k.LogLabelMatchers))
	d.Set("metric_label_matcher", flattenLabelMatchers(k.MetricLabelMatchers))
	d.Set("endpoints", flattenEndpoints(k.Endpoints))
	if k.APIKey != "" {
		d.Set("api_key", k.APIKey)
	}
	if k.Status != "" {
		d.Set("status", k.Status)
	}
	d.Set("created_at", k.CreatedAt)
}

func accessKeyComputedSchema() map[string]*schema.Schema {
	return map[string]*schema.Schema{
		"id":                 {Type: schema.TypeString, Computed: true},
		"name":               {Type: schema.TypeString, Computed: true},
		"description":        {Type: schema.TypeString, Computed: true},
		"log_permission":     {Type: schema.TypeString, Computed: true},
		"metric_permission":  {Type: schema.TypeString, Computed: true},
		"log_pipeline_id":    {Type: schema.TypeString, Computed: true},
		"metric_pipeline_id": {Type: schema.TypeString, Computed: true},
		"api_key":            {Type: schema.TypeString, Computed: true, Sensitive: true},
		"status":             {Type: schema.TypeString, Computed: true},
		"created_at":         {Type: schema.TypeString, Computed: true},
		"log_label_matcher": {
			Type:     schema.TypeList,
			Computed: true,
			Elem: &schema.Resource{
				Schema: map[string]*schema.Schema{
					"name":  {Type: schema.TypeString, Computed: true},
					"op":    {Type: schema.TypeString, Computed: true},
					"value": {Type: schema.TypeString, Computed: true},
				},
			},
		},
		"metric_label_matcher": {
			Type:     schema.TypeList,
			Computed: true,
			Elem: &schema.Resource{
				Schema: map[string]*schema.Schema{
					"name":  {Type: schema.TypeString, Computed: true},
					"op":    {Type: schema.TypeString, Computed: true},
					"value": {Type: schema.TypeString, Computed: true},
				},
			},
		},
		"endpoints": endpointsSchema(),
	}
}
