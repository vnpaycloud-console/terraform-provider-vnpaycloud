---
page_title: "vnpaycloud_snapshot Data Source - VNPayCloud"
subcategory: "Storage"
description: |-
  Get information about a volume snapshot in VNPayCloud.
---

# vnpaycloud_snapshot (Data Source)

Use this data source to get information about an existing volume snapshot, including the source volume and its current status. Snapshots can be used to restore volumes or create new volumes from a known state.

## Example Usage

```hcl
data "vnpaycloud_snapshot" "example" {
  name = "my-volume-snapshot"
}

output "snapshot_volume_id" {
  value = data.vnpaycloud_snapshot.example.volume_id
}
```

```hcl
data "vnpaycloud_snapshot" "by_id" {
  id = "snap-stu99001"
}
```

## Schema

### Optional (filter)

- `id` (String) The ID of the snapshot.
- `name` (String) The name of the snapshot.

### Read-Only

- `description` (String) A human-readable description of the snapshot.
- `volume_id` (String) The ID of the source volume from which this snapshot was taken.
- `size` (Number) The size of the snapshot in gigabytes (GB), which matches the source volume size.
- `status` (String) The current status of the snapshot (e.g., `available`, `creating`, `deleting`, `error`).
- `created_at` (String) The timestamp when the snapshot was created, in ISO 8601 format.
