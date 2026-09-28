---
page_title: "vnpaycloud_backup_kubernetes Data Source - VNPayCloud"
subcategory: "Backup"
description: |-
  Reads a VNPayCloud Kubernetes cluster backup by ID or name.
---

# vnpaycloud_backup_kubernetes (Data Source)

Reads a single Kubernetes cluster backup by its ID or its name.

To list every cluster backup in the project, use [`vnpaycloud_backup_kubernetes_list`](backup_kubernetes_list.md).

## Example Usage

```hcl
data "vnpaycloud_backup_kubernetes" "prod" {
  id = "<cluster-backup-id>"
}

data "vnpaycloud_backup_kubernetes" "by_name" {
  name = "prod-cluster"
}

output "backup_status" {
  value = data.vnpaycloud_backup_kubernetes.prod.status
}
```

## Schema

### Optional

- `id` (String) — Cluster backup ID. At least one of `id` or `name` must be set.
- `name` (String) — Backup name. Must match exactly one cluster backup. At least one of `id` or `name` must be set.

### Read-Only

- `description` (String) — Backup description.
- `protected_cluster_id` (String) — ID of the protected workload cluster.
- `pvc_backup_policy_id` (String) — ID of the Kubernetes backup policy in effect.
- `cluster_backup_type` (String) — Backup type (`on_demand`).
- `zone_id` (String) — Zone the backup lives in.
- `project_id` (String) — Project the backup belongs to.
- `customer_username` (String) — Owner of the backup.
- `status` (String) — Lifecycle status.
- `created_at` (String) — Creation timestamp (RFC3339).
