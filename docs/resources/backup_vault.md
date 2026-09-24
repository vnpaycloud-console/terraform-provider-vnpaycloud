---
page_title: "vnpaycloud_backup_vault Resource - VNPayCloud"
subcategory: "Backup"
description: |-
  Manages a VNPayCloud backup vault — the storage destination that backup policies write recovery points to.
---

# vnpaycloud_backup_vault (Resource)

Manages a backup vault. A vault is the storage destination that backup policies write their recovery points to. Its zone is taken from the provider's `zone_id`.

## Example Usage

```hcl
resource "vnpaycloud_backup_vault" "servers" {
  name        = "prod-server-vault"
  purpose     = "backup_server"
  type        = "s3"
  description = "Recovery points for production servers"
}
```

### Ceph vault

A project holds at most one `ceph` vault, so this normally already exists and is
referenced rather than created. Creating a second one is rejected server-side.

```hcl
resource "vnpaycloud_backup_vault" "first_tier" {
  name    = "prod-first-tier"
  purpose = "backup_server"
  type    = "ceph"
}
```

## Schema

### Required

- `name` (String) — Vault name. Letters, digits, hyphens, underscores and dots only. 3–100 characters (validated at plan time). Can be updated in place.
- `purpose` (String, Forces new resource) — What the vault stores. One of `backup_server`, `workload_cluster` (validated at plan time).
- `type` (String, Forces new resource) — Storage backend. One of `ceph`, `s3` (validated at plan time).

### Optional

- `description` (String) — Free-form description. Can be updated in place.

### Read-Only

- `id` (String) — Vault ID.
- `zone_id` (String) — Zone the vault lives in.
- `storage_location_id` (String) — Storage location backing the vault.
- `disk_used` (Number) — Storage currently used, in GB.
- `quota` (Number) — Storage quota, in GB.
- `committed_quota` (Number) — Storage capacity allotted to the vault, in GB. `-1` means unlimited.
- `object_lock` (Boolean) — Whether recovery points are locked against deletion for a retention window.
- `lock_time_unit` (String) — Retention unit for `object_lock`: `day` or `year`.
- `lock_time_number` (Number) — Retention length in `lock_time_unit` units.
- `status` (String) — Lifecycle status (`active`, `creating`, `error`, `deleting`, `deleted`).
- `created_at` (String) — Creation timestamp (RFC3339).

## Timeouts

- `create` — default 10 minutes.
- `delete` — default 10 minutes.
## Import

```shell
terraform import vnpaycloud_backup_vault.servers <vault-id>
```
