---
page_title: "vnpaycloud_maas_pipeline_executor Data Source - VNPayCloud"
subcategory: "Monitoring as a Service"
description: |-
  Looks up a single transformation step inside a VNPayCloud monitoring pipeline by ID.
---

# vnpaycloud_maas_pipeline_executor (Data Source)

Looks up one transformation step by ID.

## Example Usage

```hcl
data "vnpaycloud_maas_pipeline_executor" "step" {
  id = "e3d1c0f2-..."
}
```

## Schema

### Required

- `id` (String) — Executor ID.

### Read-Only

- `pipeline_id` (String) — Pipeline the step belongs to.
- `order` (Number) — Position in the pipeline.
- `type` (String) — Signal the step handles: `log` or `metric`.
- `executor_type` (String) — What the step does: `static_labels`, `rename_fields` or `rename_labels`.
- `static_labels` (Map of String) — Labels added to every stream.
- `rename_fields` (Map of String) — Old field name to new field name.
- `rename_labels` (Map of String) — Old label name to new label name.
- `description` (String) — Free-form description.
- `status` (String) — Lifecycle status.
- `created_at` (String) — Creation timestamp (RFC3339).
