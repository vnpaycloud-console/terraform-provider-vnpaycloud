---
page_title: "vnpaycloud_backup_server_restore_points Data Source - VNPayCloud"
subcategory: "Backup"
description: |-
  Lists backup restore points, optionally filtered by backup server or name.
---

# vnpaycloud_backup_server_restore_points (Data Source)

Lists the restore points produced by server backups. Use it to look up a restore point ID to boot a new instance from with the `restore_point_id` argument of [`vnpaycloud_instance`](../resources/instance.md).

## Example Usage

```hcl
data "vnpaycloud_backup_server_restore_points" "web" {
  backup_server_id = vnpaycloud_backup_server.web.id
}

# create a new VM from the most recent restore point
resource "vnpaycloud_instance" "restored" {
  name                  = "restored-web"
  restore_point_id      = data.vnpaycloud_backup_server_restore_points.web.restore_points[0].id
  flavor                = "a-pro-small.2x2"
  network_interface_ids = [vnpaycloud_network_interface.nic.id]
}
```

## Schema

### Optional

- `backup_server_id` (String) — Filter to restore points of this backup server.
- `name` (String) — Filter by restore point name.

### Read-Only

- `restore_points` (List of Object) — Matching restore points, newest first. Each element contains:
  - `id` (String) — Restore point ID. Use as `restore_point_id` on `vnpaycloud_instance`.
  - `name` (String) — Restore point name.
  - `backup_server_id` (String) — The backup server this restore point belongs to.
  - `backup_policy_id` (String) — The backup policy that produced it.
  - `type` (String) — `vm` or `volume`.
  - `purpose` (String) — `on_demand`.
  - `zone_id` (String) — Zone the restore point lives in.
  - `is_visible` (Boolean) — Whether the restore point is visible.
  - `create_by_backup_now` (Boolean) — Whether it was produced by an on-demand "Backup Now".
  - `backup_point` (String) — The point in time captured (RFC3339).
  - `status` (String) — Lifecycle status.
  - `created_at` (String) — Creation timestamp (RFC3339).
