---
page_title: "vnpaycloud_database_postgres_instance Resource - VNPayCloud"
subcategory: "Database"
description: |-
  Manages a VNPayCloud DBaaS PostgreSQL instance (standalone or cluster).
---

# vnpaycloud_database_postgres_instance (Resource)

Manages a VNPayCloud DBaaS PostgreSQL instance. Provisioning is asynchronous — the provider waits until the instance becomes `active` (and fails fast if it reports `error`).

~> **Day-2 operations map to discrete actions, not a single update.** Changing `replica` scales, `flavor_database_id` changes flavor, `volume_size` expands the disk, and `enable_tls`/`enable_read_only_endpoint` toggle those features — each is applied independently and waits for the instance to return to `active`.

## Example Usage

```hcl
data "vnpaycloud_database_flavors" "all" {}

resource "vnpaycloud_database_postgres_instance" "main" {
  name               = "my-pg"
  flavor_database_id = data.vnpaycloud_database_flavors.all.flavors[0].id
  version            = "15.13"
  volume_type        = "c1-standard"
  volume_size        = 20
  mode               = "standalone"
  replica            = 1

  enable_tls = true
  tls_mode   = "require"

  is_auto_expand_volume = true
  usage_threshold       = 80
  scale_percent         = 20
}
```

## Schema

### Required

- `name` (String, Forces new resource) — Max 20 characters.
- `flavor_database_id` (String) — Compute flavor; updating issues a change-flavor.
- `version` (String, Forces new resource) — one of `15.13`, `16.9`, `17.2`, `17.4`, `17.5`. See the `vnpaycloud_database_postgres_versions` data source.
- `volume_type` (String, Forces new resource) — e.g. `c1-standard` (see the `vnpaycloud_volume_types` data source).
- `volume_size` (Number) — GiB, 10–2000; updating issues an expand-volume (grow only).
- `mode` (String, Forces new resource) — `standalone` or `cluster`. A `standalone` instance must have `replica = 1`; a `cluster` must have `replica >= 2`.
- `replica` (Number) — Node count. `standalone` is fixed at `1`. For `cluster`, updating issues a scale — the new value must be `>= 2` and differ from the current count. Scaling is only supported in `cluster` mode.

### Optional

- `description` (String, Forces new resource)
- `purpose` (String, Forces new resource)
- `enable_tls` (Boolean) — Enable TLS. `certificate_id` / `tls_mode` may only be set when `true`.
- `certificate_id` (String)
- `tls_mode` (String) — `require` or `verify-ca`.
- `is_auto_expand_volume` (Boolean) — Enable disk auto-expand. Defaults to `false`.
- `usage_threshold` (Number) — Disk-usage % that triggers auto-expand. Only applies when `is_auto_expand_volume = true`; the API reports `0` while auto-expand is disabled, so leave it unset (or expect it to read back as `0`) unless the feature is on.
- `scale_percent` (Number) — % to grow by on auto-expand. Only applies when `is_auto_expand_volume = true` (see `usage_threshold`).
- `enable_read_only_endpoint` (Boolean) — Expose a read-only endpoint backed by the standby. Only supported when `mode = cluster`. Defaults to `false`. When enabled, `standby_ip` / `standby_port` are populated.

### Read-Only

- `id` (String)
- `primary_ip` / `primary_port` — Primary (read-write) endpoint.
- `standby_ip` / `standby_port` — Standby (read-only) endpoint. Only populated when `enable_read_only_endpoint = true` (cluster mode).
- `status` (String) — `active`, `creating`, `error`, `deleting`, `deleted`.
- `created_at` (String)

## Import

```shell
terraform import vnpaycloud_database_postgres_instance.main <instance-id>
```
