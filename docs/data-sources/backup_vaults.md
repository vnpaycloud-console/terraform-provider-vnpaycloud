---
page_title: "vnpaycloud_backup_vaults Data Source - VNPayCloud"
subcategory: "Backup"
description: |-
  Lists VNPayCloud backup vaults, optionally filtered by name, purpose or type.
---

# vnpaycloud_backup_vaults (Data Source)

Lists backup vaults in the project's zone. All filters are optional; with none set, every vault in the zone is returned.

## Example Usage

```hcl
data "vnpaycloud_backup_vaults" "all" {}

data "vnpaycloud_backup_vaults" "ceph_server_vaults" {
  purpose = "backup_server"
  type    = "ceph"
}

output "vault_names" {
  value = [for v in data.vnpaycloud_backup_vaults.all.backup_vaults : v.name]
}
```

## Schema

### Optional

- `name` (String) — Filter by vault name.
- `purpose` (String) — Filter by purpose. One of `backup_server`, `workload_cluster` (validated at plan time).
- `type` (String) — Filter by storage backend. One of `ceph`, `s3` (validated at plan time).

### Read-Only

- `backup_vaults` (List of Object) — Matching vaults. Each element contains:
  - `id` (String) — Vault ID.
  - `name` (String) — Vault name.
  - `purpose` (String) — What the vault stores.
  - `type` (String) — Storage backend.
  - `description` (String) — Vault description.
  - `object_lock` (Boolean) — Whether recovery points are locked against deletion.
  - `lock_time_unit` (String) — Retention unit for object lock.
  - `lock_time_number` (Number) — Retention length in `lock_time_unit` units.
  - `zone_id` (String) — Zone the vault lives in.
  - `storage_location_id` (String) — Storage location backing the vault.
  - `disk_used` (Number) — Storage currently used, in GB.
  - `quota` (Number) — Storage quota, in GB.
  - `committed_quota` (Number) — Storage capacity allotted to the vault, in GB. `-1` means unlimited.
  - `status` (String) — Lifecycle status.
  - `created_at` (String) — Creation timestamp (RFC3339).
