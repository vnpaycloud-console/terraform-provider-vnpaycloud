---
page_title: "vnpaycloud_database_redis_sentinel_instance Resource - VNPayCloud"
subcategory: "Database"
description: |-
  Manages a VNPayCloud DBaaS Redis Sentinel instance (HA).
---

# vnpaycloud_database_redis_sentinel_instance (Resource)

Manages a VNPayCloud DBaaS Redis Sentinel instance — a highly-available Redis topology with a separate data tier and sentinel tier, each with its own replica count, flavor and volume. Provisioning is asynchronous; the provider waits until the instance is `active` (and fails fast on `error`).

~> **Check zone support first.** Before creating a database instance, call `GET /v2/iac/projects/{project_id}/database/pools?zone_id=<zone>` to confirm that the target zone has a supported database pool (an empty result means the zone is not yet supported).

~> **Two tiers.** Data-tier fields (`replica`, `flavor_database_id`, `volume_size`) and sentinel-tier fields (`sentinel_replica`, `sentinel_flavor_database_id`, `sentinel_volume_size`) scale/change independently via their own actions.

~> **Name length.** The combined length of `name` and `sentinel_name` must not exceed 20 characters.

## Example Usage

```hcl
data "vnpaycloud_database_flavors" "all" {}

resource "vnpaycloud_database_redis_sentinel_instance" "ha" {
  name               = "my-redis-ha"
  flavor_database_id = data.vnpaycloud_database_flavors.all.flavors[0].id
  version            = "7.4.1"
  volume_type        = "c1-standard"
  volume_size        = 10
  replica            = 2

  sentinel_name               = "my-sentinel"
  sentinel_replica            = 3
  sentinel_flavor_database_id = data.vnpaycloud_database_flavors.all.flavors[0].id
  sentinel_volume_size        = 5

  enable_read_only_endpoint = true
}
```

## Schema

### Required

- `name` (String, Forces new resource) — Combined with `sentinel_name` must be ≤ 20 characters, and must differ from `sentinel_name`.
- `flavor_database_id` (String) — Data-tier flavor; updating issues a change-flavor.
- `version` (String, Forces new resource) — one of `6.2.16`, `7.2.6`, `7.4.1`, `valkey-7.2.9`, `valkey-8.0.3`, `valkey-8.1.1`. See the `vnpaycloud_database_redis_versions` data source.
- `volume_type` (String, Forces new resource)
- `volume_size` (Number) — Data-tier disk (GiB), 10–2000; updating expands.
- `replica` (Number) — Data-tier node count (≥2); updating scales (new value must be ≥2 and differ from current).
- `sentinel_name` (String, Forces new resource) — Combined with `name` must be ≤ 20 characters, and must differ from `name`.
- `sentinel_replica` (Number) — Sentinel node count (≥3); updating issues a sentinel-scale (new value must be ≥3 and differ from current).
- `sentinel_flavor_database_id` (String) — Sentinel flavor; updating issues a sentinel-change-flavor.
- `sentinel_volume_size` (Number, Forces new resource) — Sentinel disk (GiB), 10–2000.

### Optional

- `description` (String, Forces new resource)
- `purpose` (String, Forces new resource)
- `enable_tls` (Boolean) — `certificate_id` may only be set when `true`.
- `certificate_id` (String)
- `is_auto_expand_volume` (Boolean) — Enable disk auto-expand. Defaults to `false`.
- `usage_threshold` (Number) — Only applies when `is_auto_expand_volume = true`; reads back as `0` while disabled.
- `scale_percent` (Number) — Only applies when `is_auto_expand_volume = true`.
- `enable_read_only_endpoint` (Boolean) — Expose a read-only endpoint backed by the standby. Defaults to `false`. When enabled, `standby_ip` / `standby_port` are populated.

### Read-Only

- `id` (String)
- `primary_ip` / `primary_port`
- `standby_ip` / `standby_port` — Only populated when `enable_read_only_endpoint = true`.
- `status` (String) — `active`, `creating`, `error`, `deleting`, `deleted`.
- `created_at` (String)

## Import

```shell
terraform import vnpaycloud_database_redis_sentinel_instance.ha <instance-id>
```
