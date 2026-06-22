---
page_title: "vnpaycloud_database_redis_instance Resource - VNPayCloud"
subcategory: "Database"
description: |-
  Manages a VNPayCloud DBaaS Redis instance (standalone).
---

# vnpaycloud_database_redis_instance (Resource)

Manages a VNPayCloud DBaaS Redis instance (standalone topology). Provisioning is asynchronous — the provider waits until the instance becomes `active` (and fails fast if it reports `error`). For a highly-available topology use `vnpaycloud_database_redis_sentinel_instance`.

~> **Day-2 operations map to discrete actions.** `flavor_database_id` changes flavor, `volume_size` expands the disk, and `enable_tls` toggles TLS — each applied independently. `replica` is fixed at create time (standalone Redis is not scaled in place).

## Example Usage

```hcl
data "vnpaycloud_database_flavors" "all" {}

resource "vnpaycloud_database_redis_instance" "cache" {
  name               = "my-redis"
  flavor_database_id = data.vnpaycloud_database_flavors.all.flavors[0].id
  version            = "7.4.1"
  volume_type        = "c1-standard"
  volume_size        = 10
  replica            = 1
}
```

## Schema

### Required

- `name` (String, Forces new resource) — Max 20 characters.
- `flavor_database_id` (String) — Compute flavor; updating issues a change-flavor.
- `version` (String, Forces new resource) — one of `6.2.16`, `7.2.6`, `7.4.1`, `valkey-7.2.9`, `valkey-8.0.3`, `valkey-8.1.1`. See the `vnpaycloud_database_redis_versions` data source.
- `volume_type` (String, Forces new resource)
- `volume_size` (Number) — GiB, 10–2000; updating issues an expand-volume (grow only).
- `replica` (Number, Forces new resource) — Standalone Redis only supports `1`.

### Optional

- `description` (String, Forces new resource)
- `purpose` (String, Forces new resource)
- `enable_tls` (Boolean) — `certificate_id` may only be set when `true`.
- `certificate_id` (String)
- `is_auto_expand_volume` (Boolean) — Enable disk auto-expand. Defaults to `false`.
- `usage_threshold` (Number) — Disk-usage % that triggers auto-expand. Only applies when `is_auto_expand_volume = true`; reads back as `0` while auto-expand is disabled.
- `scale_percent` (Number) — % to grow by on auto-expand. Only applies when `is_auto_expand_volume = true`.

### Read-Only

- `id` (String)
- `primary_ip` / `primary_port`
- `status` (String) — `active`, `creating`, `error`, `deleting`, `deleted`.
- `created_at` (String)

## Import

```shell
terraform import vnpaycloud_database_redis_instance.cache <instance-id>
```
