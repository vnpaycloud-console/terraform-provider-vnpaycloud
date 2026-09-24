---
page_title: "vnpaycloud_backup_server_restore_point Resource - VNPayCloud"
subcategory: "Backup"
description: |-
  Triggers an on-demand backup ("Backup Now") on a server backup and manages the restore point it produces.
---

# vnpaycloud_backup_server_restore_point (Resource)

Runs an on-demand backup on a [`vnpaycloud_backup_server`](backup_server.md) — the equivalent of "Backup Now" — and manages the restore point it produces. Apply blocks until the restore point is `active`, so its `id` can be used right away to boot a new instance with the `restore_point_id` argument of [`vnpaycloud_instance`](instance.md).

Destroying this resource **deletes the restore point** (the backup). The backend deletes it asynchronously, so destroy waits until the restore point is actually gone rather than returning as soon as the request is accepted.

## Example Usage

```terraform
resource "vnpaycloud_backup_server_restore_point" "now" {
  backup_server_id = vnpaycloud_backup_server.web.id
}

resource "vnpaycloud_instance" "restored" {
  name                  = "restored-web"
  restore_point_id      = vnpaycloud_backup_server_restore_point.now.id
  flavor                = "a-pro-small.2x2"
  network_interface_ids = [vnpaycloud_network_interface.nic.id]
}
```

## Schema

### Required

- `backup_server_id` (String, Forces new resource) — The server backup to run. Changing this creates a new restore point.

### Read-Only

- `id` (String) — Restore point ID. Use as `restore_point_id` on `vnpaycloud_instance`.
- `name` (String) — Restore point name.
- `backup_policy_id` (String) — The backup policy that produced it.
- `type` (String) — `vm` or `volume`.
- `purpose` (String) — `on_demand`.
- `zone_id` (String) — Zone the restore point lives in.
- `is_visible` (Boolean) — Whether the restore point is visible.
- `create_by_backup_now` (Boolean) — Always `true` for this resource.
- `backup_point` (String) — The point in time captured (RFC3339).
- `status` (String) — Lifecycle status.
- `created_at` (String) — Creation timestamp (RFC3339).

## Timeouts

- `create` — default 20 minutes. Time allowed for the backup to reach `active`.
- `delete` — default 60 minutes. Time allowed for the restore point to disappear from the backend.

## Import

```shell
terraform import vnpaycloud_backup_server_restore_point.now <restore-point-id>
```
