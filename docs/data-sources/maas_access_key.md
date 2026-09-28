---
page_title: "vnpaycloud_maas_access_key Data Source - VNPayCloud"
subcategory: "Monitoring as a Service"
description: |-
  Looks up a single VNPayCloud monitoring access key by ID or name.
---

# vnpaycloud_maas_access_key (Data Source)

Looks up one monitoring access key. Give `id`, `name`, or both — with both, the name is checked
against the fetched key and a mismatch is an error. Looking up by name fails if more than one key
carries it.

The basic-auth `username` and `password` are not exposed here: the service discloses them only when
the key is created.

## Example Usage

```hcl
data "vnpaycloud_maas_access_key" "shipper" {
  name = "prod-shipper"
}

output "ingest_endpoint" {
  value = data.vnpaycloud_maas_access_key.shipper.endpoints[0].log_otlp_push
}
```

## Schema

### Optional

At least one of `id` or `name` is required.

- `id` (String) — Access key ID.
- `name` (String) — Access key name.

### Read-Only

- `description` (String) — Free-form description.
- `log_permission` (String) — `none`, `read`, `write` or `read_write`.
- `metric_permission` (String) — `none`, `read`, `write` or `read_write`.
- `log_pipeline_id` (String) — Pipeline applied to logs sent with this key.
- `metric_pipeline_id` (String) — Pipeline applied to metrics sent with this key.
- `log_label_matcher` (List) — Read restrictions on logs, each with `name`, `op` and `value`.
- `metric_label_matcher` (List) — Read restrictions on metrics, each with `name`, `op` and `value`.
- `api_key` (String, Sensitive) — Token for `Authorization: Bearer`, `X-Api-Key` or `X-Auth-Token`.
- `endpoints` (List) — Ingest endpoints, with `log_otlp_push`, `metric_otlp_push`, `log_loki_push` and `metric_prometheus_push`.
- `status` (String) — `active` or `inactive`.
- `created_at` (String) — Creation timestamp (RFC3339).
