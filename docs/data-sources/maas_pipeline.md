---
page_title: "vnpaycloud_maas_pipeline Data Source - VNPayCloud"
subcategory: "Monitoring as a Service"
description: |-
  Looks up a single VNPayCloud monitoring pipeline by ID or name.
---

# vnpaycloud_maas_pipeline (Data Source)

Looks up one monitoring pipeline. Give `id`, `name`, or both — with both, the name is checked
against the fetched pipeline and a mismatch is an error. Looking up by name fails if more than one
pipeline carries it, which happens when a log and a metric pipeline share a name — use `id` there.

## Example Usage

```hcl
data "vnpaycloud_maas_pipeline" "app_logs" {
  name = "app-logs"
}

resource "vnpaycloud_maas_pipeline_executor" "tag_env" {
  pipeline_id   = data.vnpaycloud_maas_pipeline.app_logs.id
  order         = 1
  executor_type = "static_labels"

  static_labels = {
    environment = "production"
  }
}
```

## Schema

### Optional

At least one of `id` or `name` is required.

- `id` (String) — Pipeline ID.
- `name` (String) — Pipeline name.

### Read-Only

- `type` (String) — Signal the pipeline handles: `log` or `metric`.
- `description` (String) — Free-form description.
- `status` (String) — Lifecycle status.
- `created_at` (String) — Creation timestamp (RFC3339).
