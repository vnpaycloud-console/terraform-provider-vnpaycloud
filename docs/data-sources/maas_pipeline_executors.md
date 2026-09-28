---
page_title: "vnpaycloud_maas_pipeline_executors Data Source - VNPayCloud"
subcategory: "Monitoring as a Service"
description: |-
  Lists the transformation steps of VNPayCloud monitoring pipelines.
---

# vnpaycloud_maas_pipeline_executors (Data Source)

Lists transformation steps, optionally narrowed to one pipeline, one signal or one kind of step.

## Example Usage

```hcl
data "vnpaycloud_maas_pipeline" "app_logs" {
  name = "app-logs"
}

data "vnpaycloud_maas_pipeline_executors" "app_log_steps" {
  pipeline_id = data.vnpaycloud_maas_pipeline.app_logs.id
}

output "step_order" {
  value = [
    for e in data.vnpaycloud_maas_pipeline_executors.app_log_steps.pipeline_executors :
    "${e.order}:${e.executor_type}"
  ]
}
```

## Schema

### Optional

- `pipeline_id` (String) — Return only steps of this pipeline.
- `type` (String) — Return only steps handling this signal: `log` or `metric` (validated at plan time).
- `executor_type` (String) — Return only steps of this kind: `static_labels`, `rename_fields` or `rename_labels` (validated at plan time).

### Read-Only

- `pipeline_executors` (List) — Matching steps, each with:
  - `id` (String)
  - `pipeline_id` (String)
  - `order` (Number)
  - `type` (String)
  - `executor_type` (String)
  - `static_labels` (Map of String)
  - `rename_fields` (Map of String)
  - `rename_labels` (Map of String)
  - `description` (String)
  - `status` (String)
  - `created_at` (String)
