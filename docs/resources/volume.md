---
page_title: "vnpaycloud_volume Resource - VNPayCloud"
subcategory: "Storage"
description: |-
  Manages a block storage volume within VNPayCloud.
---

# vnpaycloud_volume (Resource)

Manages a block storage volume within VNPayCloud. Volumes provide persistent block storage that can be attached to compute instances. Volumes can be created from scratch or restored from a snapshot.

~> **Note:** The `size` attribute can only be increased (grown). Shrinking a volume is not supported and will result in an error.

~> **Expand the filesystem inside the guest after resizing.** Growing `size` only enlarges the underlying block device — it does **not** automatically extend the partition or filesystem, so the extra space is not usable until you resize them on the VM. For an attached volume, log in to the instance and grow the partition and filesystem to match, for example:

```shell
# Example: grow partition 2 on /dev/sda, then resize its ext filesystem
growpart /dev/sda 2
resize2fs /dev/sda2
```

Run `df -h` first to see where the volume is mounted and confirm your disk and partition names — they may differ. Adjust the partition number and use the matching filesystem tool for your layout (e.g. `xfs_growfs <mountpoint>` for XFS).

## Example Usage

### Creating a standard volume

```hcl
resource "vnpaycloud_volume" "data" {
  name        = "app-data-volume"
  size        = 100
  volume_type = "c1-standard"
  description = "Persistent data volume for the application"
}
```

### Creating an encrypted multi-attach volume from a snapshot

```hcl
resource "vnpaycloud_volume" "shared" {
  name        = "shared-data-volume"
  size        = 200
  volume_type = "c1-standard"
  description = "Shared encrypted volume"
  encrypt     = true
  multiattach = true
  snapshot_id = "snap-abc12345"
}
```

## Schema

### Required

- `name` (String) The name of the volume.
- `size` (Number) The size of the volume in gigabytes. Minimum `10`. Can only be increased after creation.
- `volume_type` (String) The type of the volume (e.g., `c1-standard`). Use the `vnpaycloud_volume_types` data source to list available values. Changing this updates the volume type in place (data is preserved) as long as the new type shares the same features as the current one. Switching between types with different features — encrypted vs. unencrypted, or multi-attach vs. single-attach — is rejected by the backend; create a new volume instead. Changing the volume type requires the volume to be attached to a server; changing it on a detached volume is rejected.

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

When importing a volume that was originally created from a snapshot, omit
`snapshot_id` from the imported configuration unless you intend Terraform to
replace the volume. `snapshot_id` is a create-only input and is not returned by
the read API.
