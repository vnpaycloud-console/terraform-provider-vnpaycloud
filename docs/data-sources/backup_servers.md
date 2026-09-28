---
page_title: "vnpaycloud_backup_servers Data Source - VNPayCloud"
subcategory: "Backup"
description: |-
  Lists VNPayCloud server backups, optionally filtered by name, protected server or purpose.
---

# vnpaycloud_backup_servers (Data Source)

Lists server backups in the project's zone. All filters are optional; with none set, every backup in the zone is returned.

## Example Usage

```hcl
data "vnpaycloud_backup_servers" "all" {}

data "vnpaycloud_backup_servers" "for_web" {
  server_ids = [vnpaycloud_instance.web.id]
}

output "backup_names" {
  value = [for b in data.vnpaycloud_backup_servers.all.backup_servers : b.name]
}
```

## Schema

### Optional

- `name` (String) — Filter by backup name.
- `server_ids` (List of String) — Filter to backups protecting these servers.
- `purpose` (String) — Filter by backup purpose. Only `on_demand` is supported (validated at plan time). Defaults to `on_demand`.

### Read-Only

- `backup_servers` (List of Object) — Matching backups. Each element contains:
  - `id` (String) — Server backup ID.
  - `name` (String) — Server backup name.
  - `server_id` (String) — Protected server ID.
  - `volume_ids` (List of String) — Protected volume IDs.
  - `backup_policy_id` (String) — Policy governing the schedule and retention.
  - `purpose` (String) — Backup purpose.
  - `zone_id` (String) — Zone the backup lives in.
  - `is_compliant` (String) — Whether the backup is meeting its policy.
  - `compliance_msg` (String) — Detail behind `is_compliant`.
  - `status` (String) — Lifecycle status.
  - `created_at` (String) — Creation timestamp (RFC3339).
