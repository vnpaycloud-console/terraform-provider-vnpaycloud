---
page_title: "vnpaycloud_kubernetes_worker_groups Data Source - VNPayCloud"
subcategory: "Kubernetes"
description: |-
  List all worker groups in a Kubernetes cluster in VNPayCloud.
---

# vnpaycloud_kubernetes_worker_groups (Data Source)

Use this data source to list all worker groups belonging to a specific Kubernetes cluster. The `cluster_id` must refer to an existing, accessible cluster.

## Example Usage

```hcl
data "vnpaycloud_kubernetes_clusters" "all" {}

data "vnpaycloud_kubernetes_worker_groups" "main_cluster" {
  cluster_id = data.vnpaycloud_kubernetes_clusters.all.clusters[0].id
}

output "all_worker_group_names" {
  value = data.vnpaycloud_kubernetes_worker_groups.main_cluster.worker_groups[*].name
}

output "autoscaling_groups" {
  value = [
    for wg in data.vnpaycloud_kubernetes_worker_groups.main_cluster.worker_groups :
    {
      name        = wg.name
      min_workers = wg.min_workers
      max_workers = wg.max_workers
    }
    if wg.auto_scaling == true
  ]
}

output "total_worker_count" {
  value = sum(data.vnpaycloud_kubernetes_worker_groups.main_cluster.worker_groups[*].num_workers)
}
```

## Schema

### Required (filter)

- `cluster_id` (String) The ID of the Kubernetes cluster to list worker groups for.

### Read-Only

- `worker_groups` (List of Object) List of worker groups. Each element contains:
  - `id` (String) The unique identifier of the worker group.
  - `name` (String) The name of the worker group.
  - `flavor` (String) The flavor (instance type) used for nodes in this worker group.
  - `num_workers` (Number) The current number of worker nodes in this group.
  - `auto_scaling` (Boolean) Whether auto-scaling is enabled for this worker group.
  - `min_workers` (Number) The minimum number of workers when auto-scaling is enabled.
  - `max_workers` (Number) The maximum number of workers when auto-scaling is enabled.
  - `volume_type` (String) The volume type for worker node root disks.
  - `volume_size` (Number) The root disk size in gigabytes for worker nodes.
  - `ssh_key_id` (String) The ID of the SSH key pair injected into the worker nodes.
  - `labels` (Map of String) Kubernetes node labels applied to all nodes in this worker group.
  - `auto_healing` (Boolean) Whether machine health checking (auto-healing) is enabled for this worker group.
  - `status` (String) The current status of the worker group, lowercase (e.g., `active`, `creating`, `error`).
  - `created_at` (String) The timestamp when the worker group was created, in ISO 8601 format.
