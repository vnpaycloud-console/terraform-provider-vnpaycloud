---
page_title: "vnpaycloud_maas_pipeline_executor Resource - VNPayCloud"
subcategory: "Monitoring as a Service"
description: |-
  Manages a single transformation step inside a VNPayCloud monitoring pipeline.
---

# vnpaycloud_maas_pipeline_executor (Resource)

Manages one transformation step inside a `vnpaycloud_maas_pipeline`. Steps run in ascending `order`,
and each step either adds fixed labels or renames incoming fields.

Which steps are allowed depends on the pipeline's signal:

| Pipeline `type` | Allowed `executor_type` | Map to set    |
|-----------------|-------------------------|---------------|
| `log`           | `static_labels`         | `static_labels` |
| `log`           | `rename_fields`         | `rename_fields` |
| `metric`        | `static_labels`         | `static_labels` |
| `metric`        | `rename_labels`         | `rename_labels` |

## Example Usage

```hcl
resource "vnpaycloud_maas_pipeline" "app_logs" {
  name = "app-logs"
  type = "log"
}

resource "vnpaycloud_maas_pipeline_executor" "tag_env" {
  pipeline_id   = vnpaycloud_maas_pipeline.app_logs.id
  order         = 1
  executor_type = "static_labels"
  description   = "Stamp the environment on every stream"

  static_labels = {
    environment = "production"
    team        = "payments"
  }
}

resource "vnpaycloud_maas_pipeline_executor" "rename_fields" {
  pipeline_id   = vnpaycloud_maas_pipeline.app_logs.id
  order         = 2
  executor_type = "rename_fields"

  rename_fields = {
    lvl = "level"
    msg = "message"
  }
}
```

### Metric pipeline

```hcl
resource "vnpaycloud_maas_pipeline" "app_metrics" {
  name = "app-metrics"
  type = "metric"
}

resource "vnpaycloud_maas_pipeline_executor" "rename_labels" {
  pipeline_id   = vnpaycloud_maas_pipeline.app_metrics.id
  order         = 1
  executor_type = "rename_labels"

  rename_labels = {
    svc = "service"
  }
}
```

## Schema

### Required

- `pipeline_id` (String, Forces new resource) — ID of the pipeline this step belongs to.
- `order` (Number) — Position in the pipeline, 1–100 (validated at plan time). Unique within the pipeline. Can be updated in place.
- `executor_type` (String) — What the step does. One of `static_labels`, `rename_fields`, `rename_labels` (validated at plan time). Can be updated in place.

### Optional

Exactly one map must be set, and it must be the one matching `executor_type` — this is checked at plan time.

- `static_labels` (Map of String) — Labels added to every stream. Keys must match `^[a-zA-Z_][a-zA-Z0-9_]*$`; values are free text up to 2048 characters. Required when `executor_type` is `static_labels`.
- `rename_fields` (Map of String) — Old field name to new field name, for `log` pipelines. Both sides must match `^[a-zA-Z_][a-zA-Z0-9_]*$`. Required when `executor_type` is `rename_fields`.
- `rename_labels` (Map of String) — Old label name to new label name, for `metric` pipelines. Both sides must match `^[a-zA-Z_][a-zA-Z0-9_]*$`. Required when `executor_type` is `rename_labels`.
- `description` (String) — Free-form description. Letters, digits, spaces, hyphens, underscores and dots only, up to 255 characters (validated at plan time). Can be updated in place.

### Read-Only

- `id` (String) — Executor ID.
- `type` (String) — Signal the step handles, inherited from the pipeline: `log` or `metric`.
- `status` (String) — Lifecycle status (`active`, `creating`, `deleting`, `deleted`, `error`).
- `created_at` (String) — Creation timestamp (RFC3339).

## Import

```shell
terraform import vnpaycloud_maas_pipeline_executor.tag_env <executor-id>
```
