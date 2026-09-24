---
page_title: "vnpaycloud_backup_kubernetes_restore_point Resource - VNPayCloud"
subcategory: "Backup"
description: |-
  Manages the lifecycle of an existing Kubernetes restore point.
---

# vnpaycloud_backup_kubernetes_restore_point (Resource)

Manages an existing Kubernetes restore point produced by a [`vnpaycloud_backup_kubernetes`](backup_kubernetes.md) backup. Restore points are created automatically by the backup schedule — this resource does **not** create them. Instead, it adopts an existing restore point (looked up by [`vnpaycloud_backup_kubernetes_restore_points`](../data-sources/backup_kubernetes_restore_points.md)) so that destroying the resource **deletes that restore point** from the backup vault, mirroring the portal's *Restore Point → Actions → Delete*.

Applying the resource verifies the restore point exists and reads its attributes; `terraform destroy` deletes it. A restore point that is currently being used by a restore cannot be deleted (the delete is rejected until the restore finishes).

## Example Usage

```hcl
data "vnpaycloud_backup_kubernetes_restore_points" "points" {
  cluster_backup_id = vnpaycloud_backup_kubernetes.prod.id
}

resource "vnpaycloud_backup_kubernetes_restore_point" "newest" {
  restore_point_id = data.vnpaycloud_backup_kubernetes_restore_points.points.restore_points[0].id
}
```

## Schema

### Required

- `restore_point_id` (String, Forces new resource) — ID of an existing restore point to manage. 3–100 characters (validated at plan time).

### Read-Only

- `id` (String) — Same as `restore_point_id`.
- `name` (String) — Restore point name.
- `cluster_id` (String) — ID of the cluster the restore point was taken from.
- `cluster_name` (String) — Name of that cluster.
- `cluster_backup_id` (String) — ID of the cluster backup that produced the restore point.
- `pvc_backup_policy_id` (String) — ID of the backup policy in effect.
- `project_id` (String) — Project the restore point belongs to.
- `zone_id` (String) — Zone (data center) of the restore point.
- `backup_vault_ids` (List of String) — Backup vaults holding the restore point data.
- `velero_backup_name` (String) — Underlying Velero backup name.
- `backup_point` (String) — Timestamp of the backup, in ISO 8601 format.
- `kubernetes_version` (String) — Kubernetes version captured in the restore point.
- `status` (String) — Restore point status, lowercase (e.g. `active`, `deleting`).

## Timeouts

- `delete` — default 15 minutes.
## Import

Restore points can be imported using the restore point `id`:

```shell
terraform import vnpaycloud_backup_kubernetes_restore_point.example <restore-point-id>
```
