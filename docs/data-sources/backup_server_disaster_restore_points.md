---
page_title: "vnpaycloud_backup_server_disaster_restore_points Data Source - VNPayCloud"
subcategory: "Backup"
description: |-
  Lists disaster restore points, optionally filtered by backup server or name.
---

# vnpaycloud_backup_server_disaster_restore_points (Data Source)

Lists the **disaster** restore points available to the project. They are produced by the platform's disaster protection, not by Terraform — unlike the on-demand points listed by [`vnpaycloud_backup_server_restore_points`](backup_server_restore_points.md), which come from a [`vnpaycloud_backup_server_restore_point`](../resources/backup_server_restore_point.md) run. A project with no disaster protection in effect returns an empty list.

Each entry carries the server it was taken from, so a point can be picked by server name. Feed the `id` to the `restore_point_id` argument of [`vnpaycloud_instance`](../resources/instance.md) to boot a new VM from it.

## Example Usage

```hcl
data "vnpaycloud_backup_server_disaster_restore_points" "web" {
  backup_server_id = vnpaycloud_backup_server.web.id
}

# boot a new VM from the most recent disaster restore point
resource "vnpaycloud_instance" "recovered" {
  name                  = "recovered-web"
  restore_point_id      = data.vnpaycloud_backup_server_disaster_restore_points.web.restore_points[0].id
  flavor                = "a-pro-small.2x2"
  network_interface_ids = [vnpaycloud_network_interface.nic.id]
}
```

## Schema

### Optional

- `backup_server_id` (String) — Filter to restore points of this backup server.
- `name` (String) — Filter by restore point name.

### Read-Only

- `restore_points` (List of Object) — Matching disaster restore points, newest first. Each element contains:
  - `id` (String) — Restore point ID. Use as `restore_point_id` on `vnpaycloud_instance`.
  - `name` (String) — Restore point name.
  - `backup_server_id` (String) — The backup server this restore point belongs to.
  - `backup_policy_id` (String) — The disaster policy that produced it.
  - `type` (String) — `vm` or `volume`.
  - `purpose` (String) — Always `disaster`.
  - `zone_id` (String) — Zone the restore point lives in.
  - `backup_point` (String) — The point in time captured (RFC3339).
  - `status` (String) — Lifecycle status.
  - `created_at` (String) — Creation timestamp (RFC3339).
  - `server_id` (String) — ID of the server the restore point was taken from.
  - `server_name` (String) — Name of that server.
  - `server_zone_id` (String) — Zone of that server.
