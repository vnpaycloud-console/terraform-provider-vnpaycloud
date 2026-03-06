---
page_title: "vnpaycloud_volume Data Source - VNPayCloud"
subcategory: "Storage"
description: |-
  Get information about a volume in VNPayCloud.
---

# vnpaycloud_volume (Data Source)

Use this data source to get information about an existing block storage volume, including its size, type, encryption status, and current attachment.

## Example Usage

```hcl
data "vnpaycloud_volume" "example" {
  name = "my-data-volume"
}

output "volume_size" {
  value = data.vnpaycloud_volume.example.size
}
```

```hcl
data "vnpaycloud_volume" "by_id" {
  id = "vol-mno55667"
}
```

## Schema

### Optional (filter)

- `id` (String) The ID of the volume.
- `name` (String) The name of the volume.

### Read-Only

- `description` (String) A human-readable description of the volume.
- `size` (Number) The size of the volume in gigabytes (GB).
- `volume_type` (String) The type of the volume (e.g., `SSD`, `HDD`, `NVMe`).
- `zone` (String) The availability zone where the volume resides.
- `status` (String) The current status of the volume (e.g., `available`, `in-use`, `error`, `creating`, `deleting`).
- `iops` (Number) The provisioned IOPS for the volume, if applicable.
- `is_encrypted` (Boolean) Whether the volume is encrypted at rest.
- `is_multiattach` (Boolean) Whether the volume supports attachment to multiple instances simultaneously.
- `is_bootable` (Boolean) Whether the volume can be used as a boot volume.
- `attached_server_id` (String) The ID of the server instance the volume is currently attached to, if any.
- `attached_server_name` (String) The name of the server instance the volume is currently attached to, if any.
- `created_at` (String) The timestamp when the volume was created, in ISO 8601 format.
