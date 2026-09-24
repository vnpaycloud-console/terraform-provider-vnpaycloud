---
page_title: "vnpaycloud_backup_kubernetes_restore Resource - VNPayCloud"
subcategory: "Backup"
description: |-
  Restores a Kubernetes workload cluster from a restore point.
---

# vnpaycloud_backup_kubernetes_restore (Resource)

Restores a Kubernetes workload cluster from a restore point produced by a [`vnpaycloud_backup_kubernetes`](backup_kubernetes.md) backup. Pick a restore point (see the [`vnpaycloud_backup_kubernetes_restore_points`](../data-sources/backup_kubernetes_restore_points.md) data source) and a destination cluster; the whole cluster is restored by default.

This resource models a one-shot action: applying it triggers a restore and **waits until the destination cluster's restore finishes** (its `status` reaches `success`); a failed restore fails the apply. Destroying it only removes it from state — it does not undo a completed restore. Every argument forces a new resource.

## Example Usage

```hcl
data "vnpaycloud_backup_kubernetes_restore_points" "points" {
  cluster_backup_id = vnpaycloud_backup_kubernetes.prod.id
}

resource "vnpaycloud_backup_kubernetes_restore" "restore" {
  restore_point_id = data.vnpaycloud_backup_kubernetes_restore_points.points.restore_points[0].id
  dest_cluster_id  = vnpaycloud_kubernetes_cluster.prod.id

  restore_type               = "all_resources"
  transform_lb_to_cluster_ip = true
  keep_node_port_numbers     = false
  restore_persistent_volumes = true
}
```

## Schema

### Required

- `restore_point_id` (String, Forces new resource) — ID of the restore point to restore from. 3–100 characters (validated at plan time).
- `dest_cluster_id` (String, Forces new resource) — ID of the cluster to restore into. 3–100 characters (validated at plan time).

### Optional

- `restore_type` (String, Forces new resource) — What to restore. One of `all_resources` (default), `cluster_scoped_resources`, `specific_namespaces`.
- `transform_lb_to_cluster_ip` (Bool, Forces new resource) — Restore `LoadBalancer` services as `ClusterIP`. Default `false`.
- `keep_node_port_numbers` (Bool, Forces new resource) — Preserve the source cluster's assigned node ports. Default `false`. Cannot be `true` together with `transform_lb_to_cluster_ip`.
- `restore_persistent_volumes` (Bool, Forces new resource) — Restore persistent volumes. Default `true`. Must be `false` when the destination cluster is in a different data center than the restore point.
- `include_cluster_scoped_resource` (Bool, Forces new resource) — Include cluster-scoped resources. Default `true`.
- `included_namespaces` (List of String, Forces new resource) — Restrict the restore to these namespaces.
- `excluded_namespaces` (List of String, Forces new resource) — Skip these namespaces.
- `included_resources` (List of String, Forces new resource) — Restrict the restore to these resource kinds.
- `excluded_resources` (List of String, Forces new resource) — Skip these resource kinds.

### Read-Only

- `id` (String) — Composite of `dest_cluster_id` and `restore_point_id`.
- `status` (String) — Restore status reported by the destination cluster when the restore finished. A restore is a fire-once action, so this records the outcome at apply time and is not refreshed afterwards.

## Timeouts

- `create` — default 15 minutes.
