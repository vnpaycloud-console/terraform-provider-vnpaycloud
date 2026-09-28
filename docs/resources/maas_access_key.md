---
page_title: "vnpaycloud_maas_access_key Resource - VNPayCloud"
subcategory: "Monitoring as a Service"
description: |-
  Manages a VNPayCloud monitoring access key — the credential that ships logs and metrics in and queries them back out.
---

# vnpaycloud_maas_access_key (Resource)

Manages a monitoring access key. A key carries a separate permission for logs and for metrics, may
attach a pipeline per signal, and may narrow what it can read down to a set of label selectors.

The key's credentials — `username`, `password` and `api_key` — are returned once, when the key is
created, and are stored in Terraform state. Treat the state as a secret.

## Example Usage

```hcl
resource "vnpaycloud_maas_access_key" "shipper" {
  name              = "prod-shipper"
  description       = "Agent credentials for the production cluster"
  log_permission    = "write"
  metric_permission = "write"

  log_pipeline_id    = vnpaycloud_maas_pipeline.app_logs.id
  metric_pipeline_id = vnpaycloud_maas_pipeline.app_metrics.id
}

output "shipper_endpoints" {
  value = vnpaycloud_maas_access_key.shipper.endpoints
}
```

### Read-only key scoped to one product

Label selectors restrict what the key can read. They require a read permission on the matching
signal, and at least one of them must be positive (`eq` or `re`).

```hcl
resource "vnpaycloud_maas_access_key" "payments_reader" {
  name              = "payments-reader"
  log_permission    = "read"
  metric_permission = "read"

  log_label_matcher {
    name  = "vnpaycloud_product"
    op    = "eq"
    value = "payments"
  }

  metric_label_matcher {
    name  = "job"
    op    = "re"
    value = "payments-.+"
  }
}
```

### Disabling a key

```hcl
resource "vnpaycloud_maas_access_key" "shipper" {
  name           = "prod-shipper"
  log_permission = "write"
  status         = "inactive"
}
```

## Schema

### Required

- `name` (String) — Key name, unique within the organization. Letters, digits, spaces, hyphens, underscores and dots only. 3–255 characters (validated at plan time). Can be updated in place.

### Optional

- `description` (String) — Free-form description. Same character set as `name`, up to 255 characters (validated at plan time). Can be updated in place.
- `log_permission` (String) — What the key may do with logs. One of `none`, `read`, `write`, `read_write` (validated at plan time). Defaults to `none`.
- `metric_permission` (String) — What the key may do with metrics. Same values as `log_permission`. Defaults to `none`.
- `log_pipeline_id` (String) — Pipeline applied to logs sent with this key. Must be a pipeline whose `type` is `log`.
- `metric_pipeline_id` (String) — Pipeline applied to metrics sent with this key. Must be a pipeline whose `type` is `metric`.
- `log_label_matcher` (Block Set, max 16) — Restricts which logs the key can read. Requires `log_permission` to include read.
- `metric_label_matcher` (Block Set, max 16) — Restricts which metrics the key can read. Requires `metric_permission` to include read.
- `status` (String) — `active` or `inactive` (validated at plan time). Defaults to `active`. An inactive key is rejected at authentication.

#### `log_label_matcher` / `metric_label_matcher`

- `name` (String, Required) — Label name. Must match `^[a-zA-Z_][a-zA-Z0-9_]*$` (validated at plan time). `__name__` is not accepted.
- `op` (String, Required) — Comparison: `eq`, `neq`, `re` or `nre` (validated at plan time). `re` and `nre` are anchored to the whole value.
- `value` (String, Required) — Value or regular expression, 1–2048 characters.

A set of selectors must contain at least one `eq` or `re`, must not repeat a label name, and a `re`
must not match the empty string. Selectors left unset mean the key can read everything the
organization holds.

### Read-Only

- `id` (String) — Key ID.
- `api_key` (String, Sensitive) — Token for `Authorization: Bearer`, `X-Api-Key` or `X-Auth-Token`.
- `username` (String, Sensitive) — Basic-auth user. Returned only when the key is created.
- `password` (String, Sensitive) — Basic-auth password. Returned only when the key is created.
- `endpoints` (List) — Ingest endpoints for this key:
  - `log_otlp_push` (String)
  - `metric_otlp_push` (String)
  - `log_loki_push` (String)
  - `metric_prometheus_push` (String)
- `created_at` (String) — Creation timestamp (RFC3339).

## Import

Importing brings back everything except the credentials, which the service only discloses when the
key is created. `username` and `password` stay empty on an imported key, and `api_key` comes back
empty once the key has been updated at least once.

```shell
terraform import vnpaycloud_maas_access_key.shipper <access-key-id>
```
