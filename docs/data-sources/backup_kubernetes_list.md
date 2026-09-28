---
page_title: "vnpaycloud_backup_kubernetes_list Data Source - VNPayCloud"
subcategory: "Backup"
description: |-
  Lists cluster backups, optionally filtered by name, protected cluster or backup policy.
---

# vnpaycloud_backup_kubernetes_list (Data Source)

Lists the [`vnpaycloud_backup_kubernetes`](../resources/backup_kubernetes.md) bindings in the project. Use it to find a cluster backup without knowing its ID — for a single one already known by ID, use the [`vnpaycloud_backup_kubernetes`](backup_kubernetes.md) data source instead.

## Example Usage

```hcl
data "vnpaycloud_backup_kubernetes_list" "all" {}

# the backup protecting one specific cluster
data "vnpaycloud_backup_kubernetes_list" "for_cluster" {
  protected_cluster_id = vnpaycloud_kubernetes_cluster.main.id
}

output "cluster_backup_id" {
  value = data.vnpaycloud_backup_kubernetes_list.for_cluster.backup_kubernetes[0].id
}
```

## Schema

### Optional

- `name` (String) — Filter by cluster backup name.
- `protected_cluster_id` (String) — Filter to the backup protecting this cluster.
- `pvc_backup_policy_id` (String) — Filter to backups governed by this policy.

### Read-Only

- `backup_kubernetes` (List of Object) — Matching cluster backups. Each element contains:
  - `id` (String) — Cluster backup ID.
  - `name` (String) — Cluster backup name.
  - `description` (String) — Free-text description.
  - `protected_cluster_id` (String) — The cluster being protected.
  - `pvc_backup_policy_id` (String) — The policy governing schedule and retention.
  - `cluster_backup_type` (String) — `on_demand` or `on_disaster_recover`.
  - `zone_id` (String) — Zone the cluster backup lives in.
  - `project_id` (String) — Project the cluster backup belongs to.
  - `customer_username` (String) — Owner of the backup.
  - `status` (String) — Lifecycle status.
  - `created_at` (String) — Creation timestamp (RFC3339).
