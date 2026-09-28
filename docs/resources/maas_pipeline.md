---
page_title: "vnpaycloud_maas_pipeline Resource - VNPayCloud"
subcategory: "Monitoring as a Service"
description: |-
  Manages a VNPayCloud monitoring pipeline — an ordered set of transformations applied to logs or metrics as they are ingested.
---

# vnpaycloud_maas_pipeline (Resource)

Manages a monitoring pipeline. A pipeline groups the transformations that are applied to telemetry
on ingestion; the transformations themselves are `vnpaycloud_maas_pipeline_executor` resources that
point at this pipeline. A pipeline handles one signal: either logs or metrics, fixed at creation.

An access key attaches a pipeline per signal, so the same pipeline can be shared by several keys.

## Example Usage

```hcl
resource "vnpaycloud_maas_pipeline" "app_logs" {
  name        = "app-logs"
  type        = "log"
  description = "Tag and normalise application logs"
}

resource "vnpaycloud_maas_pipeline" "app_metrics" {
  name        = "app-metrics"
  type        = "metric"
  description = "Tag application metrics"
}
```

## Schema

### Required

- `name` (String) — Pipeline name, unique among pipelines of the same `type`. A log and a metric pipeline may share a name. Letters, digits, spaces, hyphens, underscores and dots only. 3–255 characters (validated at plan time). Can be updated in place.
- `type` (String, Forces new resource) — Signal the pipeline handles. One of `log`, `metric` (validated at plan time).

### Optional

- `description` (String) — Free-form description. Same character set as `name`, up to 255 characters (validated at plan time). Can be updated in place.

### Read-Only

- `id` (String) — Pipeline ID.
- `status` (String) — Lifecycle status (`active`, `creating`, `deleting`, `deleted`, `error`).
- `created_at` (String) — Creation timestamp (RFC3339).

## Deletion

A pipeline that an access key still references cannot be deleted; detach it from the key first.
Deleting a pipeline also deletes every executor attached to it, so those resources disappear from
the backend along with it — keep them in the same configuration so Terraform destroys them together.

## Import

```shell
terraform import vnpaycloud_maas_pipeline.app_logs <pipeline-id>
```
