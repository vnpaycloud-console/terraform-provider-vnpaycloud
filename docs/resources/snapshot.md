---
page_title: "vnpaycloud_snapshot Resource - VNPayCloud"
subcategory: "Storage"
description: |-
  Manages a volume snapshot within VNPayCloud.
---

# vnpaycloud_snapshot (Resource)

Manages a point-in-time snapshot of a block storage volume within VNPayCloud. Snapshots can be used to back up volume data and to create new volumes with pre-populated content.

~> **Note:** All attributes are ForceNew. Any change to the snapshot configuration will destroy the existing snapshot and create a new one.

## Example Usage

```hcl
resource "vnpaycloud_volume" "data" {
  name        = "app-data-volume"
  size        = 100
  volume_type = "SSD"
}

resource "vnpaycloud_snapshot" "daily_backup" {
  name        = "app-data-snapshot-2024-01-15"
  volume_id   = vnpaycloud_volume.data.id
  description = "Daily backup snapshot of app data volume"
}
```

### Using a snapshot to create a new volume

```hcl
resource "vnpaycloud_volume" "restored" {
  name        = "app-data-restored"
  size        = 100
  volume_type = "SSD"
  snapshot_id = vnpaycloud_snapshot.daily_backup.id
}
```

## Schema

### Required

- `name` (String, ForceNew) The name of the snapshot. Changing this creates a new snapshot.
- `volume_id` (String, ForceNew) The ID of the volume to create a snapshot of. Changing this creates a new snapshot.

### Optional

- `description` (String, ForceNew) A human-readable description of the snapshot. Changing this creates a new snapshot.

### Read-Only

- `id` (String) The ID of the snapshot.
- `size` (Number) The size of the snapshot in gigabytes, inherited from the source volume.
- `status` (String) The current status of the snapshot (e.g., `available`, `creating`, `error`).
- `created_at` (String) The creation timestamp of the snapshot in ISO 8601 format.

## Timeouts

- `create` - (Default `10 minutes`) Used for creating the snapshot.
- `delete` - (Default `10 minutes`) Used for deleting the snapshot.

## Import

Snapshots can be imported using the `id`:

```shell
terraform import vnpaycloud_snapshot.example <snapshot-id>
```
