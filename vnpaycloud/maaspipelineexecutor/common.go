package maaspipelineexecutor

import (
	"fmt"
	"regexp"
	"terraform-provider-vnpaycloud/vnpaycloud/dto"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

const (
	executorTypeStaticLabels = "static_labels"
	executorTypeRenameFields = "rename_fields"
	executorTypeRenameLabels = "rename_labels"
)

// Label and field names are Prometheus/Loki identifiers.
var identifierRegex = regexp.MustCompile(`^[a-zA-Z_][a-zA-Z0-9_]*$`)

func expandStringMap(raw interface{}) map[string]string {
	if raw == nil {
		return nil
	}
	in, ok := raw.(map[string]interface{})
	if !ok || len(in) == 0 {
		return nil
	}
	out := make(map[string]string, len(in))
	for k, v := range in {
		out[k] = v.(string)
	}
	return out
}

func flattenStringMap(in map[string]string) map[string]interface{} {
	out := make(map[string]interface{}, len(in))
	for k, v := range in {
		out[k] = v
	}
	return out
}

// validateMap checks label/field names, and values too when they are identifiers rather than free text.
func validateMap(field string, m map[string]string, valuesAreIdentifiers bool) error {
	for k, v := range m {
		if !identifierRegex.MatchString(k) {
			return fmt.Errorf("%s key %q must match %s", field, k, identifierRegex.String())
		}
		if v == "" {
			return fmt.Errorf("%s value for key %q must not be empty", field, k)
		}
		if valuesAreIdentifiers && !identifierRegex.MatchString(v) {
			return fmt.Errorf("%s value %q must match %s", field, v, identifierRegex.String())
		}
	}
	return nil
}

// validateExecutorShape enforces that exactly the map matching executor_type is set. The pipeline
// telemetry type decides which executor types are legal, and that is only known server-side.
func validateExecutorShape(executorType string, staticLabels, renameFields, renameLabels map[string]string) error {
	// Each executor type is named after the one map it reads.
	type executorMap struct {
		name                 string
		values               map[string]string
		valuesAreIdentifiers bool
	}

	maps := []executorMap{
		{executorTypeStaticLabels, staticLabels, false},
		{executorTypeRenameFields, renameFields, true},
		{executorTypeRenameLabels, renameLabels, true},
	}

	var expected *executorMap
	for i := range maps {
		if maps[i].name == executorType {
			expected = &maps[i]
			break
		}
	}
	if expected == nil {
		return fmt.Errorf("executor_type must be one of: %s, %s, %s",
			executorTypeStaticLabels, executorTypeRenameFields, executorTypeRenameLabels)
	}

	for _, m := range maps {
		if m.name == executorType || len(m.values) == 0 {
			continue
		}
		if len(expected.values) == 0 {
			return fmt.Errorf("%s must not be set when executor_type is %q; put the mapping in %s instead",
				m.name, executorType, executorType)
		}
		return fmt.Errorf("%s must not be set when executor_type is %q", m.name, executorType)
	}

	if len(expected.values) == 0 {
		return fmt.Errorf("%s must not be empty when executor_type is %q", expected.name, executorType)
	}

	return validateMap(expected.name, expected.values, expected.valuesAreIdentifiers)
}

func setExecutorData(d *schema.ResourceData, e *dto.MaasPipelineExecutor) {
	d.Set("pipeline_id", e.PipelineID)
	d.Set("order", e.Order)
	d.Set("description", e.Description)
	d.Set("type", e.Type)
	d.Set("executor_type", e.ExecutorType)
	d.Set("static_labels", flattenStringMap(e.StaticLabels))
	d.Set("rename_fields", flattenStringMap(e.RenameFields))
	d.Set("rename_labels", flattenStringMap(e.RenameLabels))
	d.Set("status", e.Status)
	d.Set("created_at", e.CreatedAt)
}

func executorComputedSchema() map[string]*schema.Schema {
	return map[string]*schema.Schema{
		"id":            {Type: schema.TypeString, Computed: true},
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
	}
}
