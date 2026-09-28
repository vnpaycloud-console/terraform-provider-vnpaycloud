---
page_title: "vnpaycloud_maas_access_keys Data Source - VNPayCloud"
subcategory: "Monitoring as a Service"
description: |-
  Lists VNPayCloud monitoring access keys.
---

# vnpaycloud_maas_access_keys (Data Source)

Lists the organization's monitoring access keys. Credentials are not exposed except for `api_key`.

## Example Usage

```hcl
data "vnpaycloud_maas_access_keys" "all" {}

output "inactive_keys" {
  value = [
    for k in data.vnpaycloud_maas_access_keys.all.access_keys :
    k.name if k.status == "inactive"
  ]
}
```

## Schema

### Optional

- `name` (String) — Filter by key name.

### Read-Only

- `access_keys` (List) — Matching keys, each with:
  - `id` (String)
  - `name` (String)
  - `description` (String)
  - `log_permission` (String)
  - `metric_permission` (String)
  - `log_pipeline_id` (String)
  - `metric_pipeline_id` (String)
  - `log_label_matcher` (List) — each with `name`, `op` and `value`
  - `metric_label_matcher` (List) — each with `name`, `op` and `value`
  - `api_key` (String, Sensitive)
  - `endpoints` (List) — with `log_otlp_push`, `metric_otlp_push`, `log_loki_push` and `metric_prometheus_push`
  - `status` (String)
  - `created_at` (String)
