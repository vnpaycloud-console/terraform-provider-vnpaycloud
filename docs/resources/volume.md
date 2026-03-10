---
page_title: "vnpaycloud_volume Resource - VNPayCloud"
subcategory: "Storage"
description: |-
  Manages a block storage volume within VNPayCloud.
---

# vnpaycloud_volume (Resource)

Manages a block storage volume within VNPayCloud. Volumes provide persistent block storage that can be attached to compute instances. Volumes can be created from scratch or restored from a snapshot.

~> **Note:** The `size` attribute can only be increased (grown). Shrinking a volume is not supported and will result in an error.

## Example Usage

### Creating a standard volume

```hcl
resource "vnpaycloud_volume" "data" {
  name        = "app-data-volume"
  size        = 100
  volume_type = "SSD"
  description = "Persistent data volume for the application"
}
```

### Creating an encrypted multi-attach volume from a snapshot

```hcl
resource "vnpaycloud_volume" "shared" {
  name        = "shared-data-volume"
  size        = 200
  volume_type = "SSD"
  description = "Shared encrypted volume"
  encrypt     = true
  multiattach = true
  snapshot_id = "snap-abc12345"
}
```

## Schema

### Required

- `name` (String) The name of the volume.
- `size` (Number) The size of the volume in gigabytes. Can only be increased after creation.
- `volume_type` (String, ForceNew) The type of the volume (e.g., `SSD`, `HDD`). Changing this creates a new volume.

### Optional

- `description` (String) A human-readable description of the volume.
- `encrypt` (Boolean, ForceNew) Whether to encrypt the volume at rest. Changing this creates a new volume. Defaults to `false`.
- `multiattach` (Boolean, ForceNew) Whether to allow the volume to be attached to multiple instances simultaneously. Changing this creates a new volume. Defaults to `false`.
- `snapshot_id` (String, ForceNew) The ID of a snapshot to create the volume from. Changing this creates a new volume.

### Read-Only

- `id` (String) The ID of the volume.
- `zone` (String) The availability zone where the volume resides.
- `status` (String) The current status of the volume (e.g., `available`, `in-use`, `error`).
- `iops` (Number) The provisioned IOPS for the volume.
- `is_encrypted` (Boolean) Whether the volume is encrypted.
- `is_multiattach` (Boolean) Whether multi-attach is enabled on the volume.
- `is_bootable` (Boolean) Whether the volume can be used as a boot volume.
- `attached_server_id` (String) The ID of the server the volume is currently attached to. Empty if not attached.
- `attached_server_name` (String) The name of the server the volume is currently attached to. Empty if not attached.
- `created_at` (String) The creation timestamp of the volume in ISO 8601 format.

## Timeouts

- `create` - (Default `10 minutes`) Used for creating the volume.
- `update` - (Default `10 minutes`) Used for updating the volume (e.g., resizing or renaming).
- `delete` - (Default `10 minutes`) Used for deleting the volume.

## Import

Volumes can be imported using the `id`:

```shell
terraform import vnpaycloud_volume.example <volume-id>
```
