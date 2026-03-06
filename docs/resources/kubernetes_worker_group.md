---
page_title: "vnpaycloud_kubernetes_worker_group Resource - VNPayCloud"
subcategory: "Kubernetes"
description: |-
  Manages a worker node group for a Kubernetes cluster within VNPayCloud.
---

# vnpaycloud_kubernetes_worker_group (Resource)

Manages an additional worker node group within a VNPayCloud managed Kubernetes cluster. Worker groups allow you to run workloads on nodes with different flavors, configurations, or labels. They support manual scaling as well as auto-scaling policies.

## Example Usage

### Fixed-size worker group

```hcl
resource "vnpaycloud_kubernetes_worker_group" "gpu_workers" {
  cluster_id   = vnpaycloud_kubernetes_cluster.main.id
  name         = "gpu-workers"
  flavor       = "gpu.4c16r.v100"
  num_workers  = 2
  volume_type  = "SSD"
  volume_size  = 100
  ssh_key_id   = "keypair-abc12345"

  labels = {
    "workload-type" = "gpu"
    "team"          = "ml-platform"
  }
}
```

### Auto-scaling worker group

```hcl
resource "vnpaycloud_kubernetes_worker_group" "auto_scale" {
  cluster_id   = vnpaycloud_kubernetes_cluster.main.id
  name         = "general-workers"
  flavor       = "s.4c8r"
  num_workers  = 3
  auto_scaling = true
  min_workers  = 2
  max_workers  = 10
  volume_type  = "SSD"
  volume_size  = 50
}
```

## Schema

### Required

- `cluster_id` (String, ForceNew) The ID of the Kubernetes cluster this worker group belongs to. Changing this creates a new worker group.
- `name` (String, ForceNew) The name of the worker group. Must be unique within the cluster. Changing this creates a new worker group.
- `flavor` (String, ForceNew) The flavor (instance type) for nodes in this worker group. Changing this creates a new worker group.
- `num_workers` (Number) The desired number of worker nodes in the group. Must be at least `1`. This attribute can be updated in-place to manually scale the group.

### Optional

- `auto_scaling` (Boolean) Whether to enable automatic scaling for this worker group. When enabled, the cluster autoscaler will adjust `num_workers` between `min_workers` and `max_workers` based on pending workloads. Defaults to `false`. Can be updated in-place.
- `min_workers` (Number) The minimum number of worker nodes when auto-scaling is enabled. Required when `auto_scaling` is `true`. Can be updated in-place.
- `max_workers` (Number) The maximum number of worker nodes when auto-scaling is enabled. Required when `auto_scaling` is `true`. Can be updated in-place.
- `volume_type` (String, ForceNew) The volume type for worker node root disks (e.g., `SSD`, `HDD`). Changing this creates a new worker group.
- `volume_size` (Number, ForceNew) The root disk size in gigabytes for worker nodes. Changing this creates a new worker group.
- `ssh_key_id` (String, ForceNew) The ID of the SSH key pair to inject into the worker nodes for direct SSH access. Changing this creates a new worker group.
- `labels` (Map of String, ForceNew) A map of Kubernetes node labels to apply to all nodes in this worker group. Useful for node selectors and affinity rules. Changing this creates a new worker group.

### Read-Only

- `id` (String) The ID of the worker group.
- `status` (String) The current status of the worker group (e.g., `ACTIVE`, `SCALING`, `ERROR`).
- `created_at` (String) The creation timestamp of the worker group in ISO 8601 format.

## Timeouts

- `create` - (Default `30 minutes`) Used for creating the worker group.
- `update` - (Default `30 minutes`) Used for scaling or updating the worker group.
- `delete` - (Default `15 minutes`) Used for deleting the worker group.

## Import

Kubernetes worker groups can be imported using the `id`:

```shell
terraform import vnpaycloud_kubernetes_worker_group.example <worker-group-id>
```
