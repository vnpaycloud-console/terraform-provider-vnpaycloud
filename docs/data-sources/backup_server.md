---
page_title: "vnpaycloud_backup_server Data Source - VNPayCloud"
subcategory: "Backup"
description: |-
  Looks up a single VNPayCloud server backup by ID or name.
---

# vnpaycloud_backup_server (Data Source)

Looks up a single server backup. Specify at least one of `id` or `name`. When both are set, the record is fetched by `id` and the lookup fails if its name does not match.

## Example Usage

```hcl
data "vnpaycloud_backup_server" "by_name" {
  name = "web-01-backup"
}

data "vnpaycloud_backup_server" "by_id" {
  id = "bs-0a1b2c3d"
}
```

## Schema

### Optional

- `id` (String) — Server backup ID. At least one of `id` or `name` must be set.
- `name` (String) — Server backup name. Must match exactly one record. At least one of `id` or `name` must be set.

### Read-Only

- `server_id` (String) — Protected server ID.
- `volume_ids` (List of String) — Protected volume IDs.
- `backup_policy_id` (String) — Policy governing the schedule and retention.
- `purpose` (String) — Backup purpose. Always `on_demand`.
- `zone_id` (String) — Zone the backup lives in.
- `is_compliant` (String) — Whether the backup is meeting its policy.
- `compliance_msg` (String) — Detail behind `is_compliant`.
- `status` (String) — Lifecycle status.
- `created_at` (String) — Creation timestamp (RFC3339).
