---
page_title: "vnpaycloud_kubernetes_worker_group Data Source - VNPayCloud"
subcategory: "Kubernetes"
description: |-
  Get information about a Kubernetes worker group in VNPayCloud.
---

# vnpaycloud_kubernetes_worker_group (Data Source)

Use this data source to get information about an existing Kubernetes worker node group, including its scaling configuration, instance flavor, and storage settings.

## Example Usage

```hcl
data "vnpaycloud_kubernetes_worker_group" "example" {
  id         = "wg-tuv67890"
  cluster_id = "k8s-qrs56789"
}

output "worker_group_size" {
  value = data.vnpaycloud_kubernetes_worker_group.example.num_workers
}

output "auto_scaling_enabled" {
  value = data.vnpaycloud_kubernetes_worker_group.example.auto_scaling
}
```

## Schema

### Required (filter)

- `id` (String) The ID of the worker group.
- `cluster_id` (String) The ID of the Kubernetes cluster this worker group belongs to.

### Read-Only

- `name` (String) The name of the worker group.
- `flavor` (String) The compute flavor used for the worker nodes (e.g., `4c-8g`, `8c-16g`).
- `num_workers` (Number) The current number of worker nodes in the group.
- `auto_scaling` (Boolean) Whether horizontal auto-scaling is enabled for this worker group.
- `min_workers` (Number) The minimum number of worker nodes when auto-scaling is enabled.
- `max_workers` (Number) The maximum number of worker nodes when auto-scaling is enabled.
- `volume_type` (String) The storage type for the worker node's data disk (e.g., `SSD`, `HDD`).
- `volume_size` (Number) The size of the worker node's data disk in gigabytes (GB).
- `ssh_key_id` (String) The ID of the SSH key pair used to access the worker nodes.
- `labels` (Map of String) A map of Kubernetes labels applied to all nodes in this worker group.
- `status` (String) The current status of the worker group (e.g., `ACTIVE`, `CREATING`, `SCALING`, `ERROR`).
- `created_at` (String) The timestamp when the worker group was created, in ISO 8601 format.
