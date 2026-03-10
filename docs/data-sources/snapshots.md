---
page_title: "vnpaycloud_snapshots Data Source - VNPayCloud"
subcategory: "Storage"
description: |-
  List all volume snapshots in VNPayCloud.
---

# vnpaycloud_snapshots (Data Source)

Use this data source to list all volume snapshots in the current project.

## Example Usage

```hcl
data "vnpaycloud_snapshots" "all" {}

output "all_snapshot_names" {
  value = data.vnpaycloud_snapshots.all.snapshots[*].name
}

output "available_snapshot_ids" {
  value = [
    for snap in data.vnpaycloud_snapshots.all.snapshots :
    snap.id if snap.status == "available"
  ]
}

output "snapshots_by_volume" {
  value = {
    for snap in data.vnpaycloud_snapshots.all.snapshots :
    snap.name => snap.volume_id
  }
}
```

## Schema

### Read-Only

- `snapshots` (List of Object) List of snapshots. Each element contains:
  - `id` (String) The unique identifier of the snapshot.
  - `name` (String) The name of the snapshot.
  - `description` (String) A human-readable description of the snapshot.
  - `volume_id` (String) The ID of the source volume from which this snapshot was created.
  - `size` (Number) The size of the snapshot in GB.
  - `status` (String) The current status of the snapshot (e.g., `available`, `creating`, `error`).
  - `created_at` (String) The timestamp when the snapshot was created, in ISO 8601 format.
