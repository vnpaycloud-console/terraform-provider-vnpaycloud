---
page_title: "vnpaycloud_volumes Data Source - VNPayCloud"
subcategory: "Storage"
description: |-
  List all volumes in VNPayCloud.
---

# vnpaycloud_volumes (Data Source)

Use this data source to list all volumes in the current project.

## Example Usage

```hcl
data "vnpaycloud_volumes" "all" {}

output "all_volume_names" {
  value = data.vnpaycloud_volumes.all.volumes[*].name
}

output "available_volume_ids" {
  value = [
    for v in data.vnpaycloud_volumes.all.volumes :
    v.id if v.status == "available"
  ]
}

output "total_storage_gb" {
  value = sum(data.vnpaycloud_volumes.all.volumes[*].size)
}
```

## Schema

### Read-Only

- `volumes` (List of Object) List of volumes. Each element contains:
  - `id` (String) The unique identifier of the volume.
  - `name` (String) The name of the volume.
  - `description` (String) A human-readable description of the volume.
  - `size` (Number) The size of the volume in GB.
  - `volume_type` (String) The type of the volume (e.g., `SSD`, `HDD`).
  - `zone` (String) The availability zone where the volume resides.
  - `status` (String) The current status of the volume (e.g., `available`, `in-use`, `error`).
  - `iops` (Number) The IOPS provisioned for this volume, if applicable.
  - `is_encrypted` (Boolean) Whether the volume is encrypted.
  - `is_multiattach` (Boolean) Whether the volume supports multi-attach.
  - `is_bootable` (Boolean) Whether the volume is bootable.
  - `attached_server_id` (String) The ID of the server this volume is attached to, if any.
  - `attached_server_name` (String) The name of the server this volume is attached to, if any.
  - `created_at` (String) The timestamp when the volume was created, in ISO 8601 format.
