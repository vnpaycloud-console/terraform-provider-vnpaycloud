---
page_title: "vnpaycloud_backup_server Resource - VNPayCloud"
subcategory: "Backup"
description: |-
  Manages a VNPayCloud server backup — puts a server and its volumes under a backup policy.
---

# vnpaycloud_backup_server (Resource)

Manages a server backup: it puts one server and a selected set of its volumes under a backup policy. The policy decides when backups run and how long recovery points are kept; this resource decides *what* is protected.

Changing which server is protected replaces the resource. The policy and the protected volume set can both be changed in place.

## Example Usage

```hcl
data "vnpaycloud_backup_vault" "servers" {
  name = "prod-server-vault"
}

resource "vnpaycloud_backup_policy_server" "daily" {
  name          = "prod-daily"
  start_hour    = 2
  resource_type = "vm"

  backup_vault_ids = [data.vnpaycloud_backup_vault.servers.id]

  daily {
    retentions = 7
  }
}

resource "vnpaycloud_backup_server" "web" {
  server_id        = vnpaycloud_instance.web.id
  volume_ids       = vnpaycloud_instance.web.volume_ids
  backup_policy_id = vnpaycloud_backup_policy_server.daily.id
}
```

### Protecting several volumes

```hcl
resource "vnpaycloud_backup_server" "db" {
  server_id = vnpaycloud_instance.db.id

  volume_ids = [
    vnpaycloud_volume.db_root.id,
    vnpaycloud_volume.db_data.id,
  ]

  backup_policy_id = vnpaycloud_backup_policy_server.daily.id
}
```

## Schema

### Required

- `server_id` (String, Forces new resource) — ID of the server to protect. 3–100 characters (validated at plan time).
- `volume_ids` (List of String) — Volumes of that server to protect. At least one is required (validated at plan time); IDs must be unique (enforced server-side). Can be updated in place.
- `backup_policy_id` (String) — Policy that governs the schedule and retention. 3–100 characters (validated at plan time). Can be updated in place.

### Read-Only

- `id` (String) — Server backup ID.
- `name` (String) — Backup name. Taken from the protected server; it cannot be set.
- `purpose` (String) — Backup purpose. Always `on_demand`.
- `zone_id` (String) — Zone the backup lives in.
- `is_compliant` (String) — Whether the backup is meeting its policy.
- `compliance_msg` (String) — Detail behind `is_compliant`.
- `status` (String) — Lifecycle status (`active`, `creating`, `error`, `deleting`, `deleted`).
- `created_at` (String) — Creation timestamp (RFC3339).

## Timeouts

- `create` — default 10 minutes.
- `delete` — default 10 minutes.

## Import

```shell
terraform import vnpaycloud_backup_server.web <backup-server-id>
```

## Restoring a backup into a new instance

A backup produces restore points over time. Look one up with the
[`vnpaycloud_backup_server_restore_points`](../data-sources/backup_server_restore_points.md)
data source — they come back newest first — and bring it back as a fresh VM with the
`restore_point_id` source of [`vnpaycloud_instance`](instance.md).

The restore point carries its own disk, so `root_disk_gb` and `root_disk_type` are left out.

```terraform
data "vnpaycloud_backup_server_restore_points" "web" {
  backup_server_id = vnpaycloud_backup_server.web.id
}

resource "vnpaycloud_instance" "restored" {
  name                  = "restored-vm"
  restore_point_id      = data.vnpaycloud_backup_server_restore_points.web.restore_points[0].id
  flavor                = "a-pro-small.2x2"
  network_interface_ids = [vnpaycloud_network_interface.nic.id]
}
```
