---
page_title: "vnpaycloud_backup_kubernetes_restore_points Data Source - VNPayCloud"
subcategory: "Backup"
description: |-
  Lists Kubernetes cluster restore points.
---

# vnpaycloud_backup_kubernetes_restore_points (Data Source)

Lists the restore points produced by a Kubernetes cluster backup, so you can feed one into a [`vnpaycloud_backup_kubernetes_restore`](../resources/backup_kubernetes_restore.md).

## Example Usage

```hcl
data "vnpaycloud_backup_kubernetes_restore_points" "points" {
  cluster_backup_id = vnpaycloud_backup_kubernetes.prod.id
}

output "latest_restore_point" {
  value = data.vnpaycloud_backup_kubernetes_restore_points.points.restore_points[0].id
}
```

## Schema

### Optional

- `cluster_backup_id` (String) — Filter by the cluster backup the restore points belong to.
- `cluster_id` (String) — Filter by protected cluster.
- `pvc_backup_policy_id` (String) — Filter by backup policy.
- `name` (String) — Filter by restore point name.

### Read-Only

- `restore_points` (List) — Matching restore points, newest first. Each element has:
  - `id` (String) — Restore point ID.
  - `name` (String) — Restore point name.
  - `cluster_id` (String) — Protected cluster ID.
  - `cluster_name` (String) — Protected cluster name.
  - `cluster_backup_id` (String) — Owning cluster backup ID.
  - `pvc_backup_policy_id` (String) — Backup policy ID.
  - `zone_id` (String) — Zone.
  - `backup_vault_ids` (List of String) — Backup vaults holding the data.
  - `velero_backup_name` (String) — Underlying backup name.
  - `backup_point` (String) — When the backup was taken (RFC3339).
  - `kubernetes_version` (String) — Kubernetes version captured.
  - `status` (String) — Restore point status.
