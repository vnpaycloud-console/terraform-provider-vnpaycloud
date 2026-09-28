---
page_title: "vnpaycloud_maas_pipelines Data Source - VNPayCloud"
subcategory: "Monitoring as a Service"
description: |-
  Lists VNPayCloud monitoring pipelines, optionally filtered by name or signal.
---

# vnpaycloud_maas_pipelines (Data Source)

Lists the organization's monitoring pipelines.

## Example Usage

```hcl
data "vnpaycloud_maas_pipelines" "logs" {
  type = "log"
}

output "log_pipeline_names" {
  value = [for p in data.vnpaycloud_maas_pipelines.logs.pipelines : p.name]
}
```

## Schema

### Optional

- `name` (String) — Filter by pipeline name.
- `type` (String) — Return only pipelines handling this signal: `log` or `metric` (validated at plan time).

### Read-Only

- `pipelines` (List) — Matching pipelines, each with:
  - `id` (String)
  - `name` (String)
  - `type` (String)
  - `description` (String)
  - `status` (String)
  - `created_at` (String)
