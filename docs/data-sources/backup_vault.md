---
page_title: "vnpaycloud_backup_vault Data Source - VNPayCloud"
subcategory: "Backup"
description: |-
  Looks up a single VNPayCloud backup vault by ID or name.
---

# vnpaycloud_backup_vault (Data Source)

Looks up a single backup vault. Specify at least one of `id` or `name`. When both are set, the vault is fetched by `id` and the lookup fails if its name does not match.

## Example Usage

```hcl
data "vnpaycloud_backup_vault" "by_name" {
  name = "prod-server-vault"
}

data "vnpaycloud_backup_vault" "by_id" {
  id = "bv-0a1b2c3d"
}
```

## Schema

### Optional

- `id` (String) — Vault ID. At least one of `id` or `name` must be set.
- `name` (String) — Vault name. Must match exactly one vault. At least one of `id` or `name` must be set.

### Read-Only

- `purpose` (String) — What the vault stores (`backup_server`, `workload_cluster`).
- `type` (String) — Storage backend (`ceph`, `s3`).
- `description` (String) — Vault description.
- `object_lock` (Boolean) — Whether recovery points are locked against deletion.
- `lock_time_unit` (String) — Retention unit for object lock (`day`, `year`).
- `lock_time_number` (Number) — Retention length in `lock_time_unit` units.
- `zone_id` (String) — Zone the vault lives in.
- `storage_location_id` (String) — Storage location backing the vault.
- `disk_used` (Number) — Storage currently used, in GB.
- `quota` (Number) — Storage quota, in GB.
- `committed_quota` (Number) — Storage capacity allotted to the vault, in GB. `-1` means unlimited.
- `status` (String) — Lifecycle status.
- `created_at` (String) — Creation timestamp (RFC3339).
