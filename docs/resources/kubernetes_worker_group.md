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
resource "vnpaycloud_kubernetes_worker_group" "extra" {
  cluster_id  = vnpaycloud_kubernetes_cluster.main.id
  name        = "extra-workers"
  flavor      = "a-pro-small.2x2"
  num_workers = 2
  volume_type = "c1-standard"
  volume_size = 50
  ssh_key_id  = vnpaycloud_keypair.k8s.id

  labels = {
    "workload-type" = "batch"
    "team"          = "platform"
  }
}
```

### Auto-scaling worker group

```hcl
resource "vnpaycloud_kubernetes_worker_group" "auto_scale" {
  cluster_id   = vnpaycloud_kubernetes_cluster.main.id
  name         = "general-workers"
  flavor       = "a-pro-small.2x2"
  num_workers  = 3
  auto_scaling = true
  min_workers  = 2
  max_workers  = 10
  volume_type  = "c1-standard"
  volume_size  = 50
}
```

## Schema

### Required

- `cluster_id` (String, ForceNew) The ID of the Kubernetes cluster this worker group belongs to. Changing this creates a new worker group.
- `name` (String, ForceNew) The name of the worker group. Must be unique within the cluster. Changing this creates a new worker group.
- `flavor` (String, ForceNew) The flavor (instance type) for nodes in this worker group. List available flavors with the `vnpaycloud_flavors` data source. Changing this creates a new worker group.
- `num_workers` (Number) The desired number of worker nodes in the group. Must be at least `1`. This attribute can be updated in-place to manually scale the group. When `auto_scaling` is enabled, this value is still sent as the desired/current node count together with the autoscaling policy.

### Optional

- `auto_scaling` (Boolean) Whether to enable automatic scaling for this worker group. When enabled, the cluster autoscaler will adjust `num_workers` between `min_workers` and `max_workers` based on pending workloads. Defaults to `false`. Can be updated in-place.
- `min_workers` (Number) The minimum number of worker nodes when auto-scaling is enabled. Required when `auto_scaling` is `true`; must be greater than `0` and less than `max_workers`. Can be updated in-place.
- `max_workers` (Number) The maximum number of worker nodes when auto-scaling is enabled. Required when `auto_scaling` is `true`; must be greater than `min_workers`. Can be updated in-place.
- `volume_type` (String, ForceNew) The volume type for worker node root disks, e.g. `c1-standard`. List available volume types with the `vnpaycloud_volume_types` data source. Changing this creates a new worker group.
- `volume_size` (Number, ForceNew) The root disk size in gigabytes for worker nodes. Must be between `50` and `200` (validated at plan time). Changing this creates a new worker group.
- `ssh_key_id` (String, ForceNew) The ID of the SSH key pair to inject into the worker nodes for direct SSH access. Changing this creates a new worker group.
- `labels` (Map of String) A map of Kubernetes node labels to apply to all nodes in this worker group. Useful for node selectors and affinity rules. Can be updated in-place. Label keys and values must be 63 characters or less.
- `auto_healing` (Boolean) Whether machine health checking (auto-healing) is enabled for this worker group. When enabled, unhealthy nodes are automatically replaced. Can be updated in-place. If omitted, the default applies (read back into state).

### Read-Only

- `id` (String) The ID of the worker group.
- `status` (String) The current status of the worker group, lowercase (e.g., `active`, `creating`, `error`).
- `created_at` (String) The creation timestamp of the worker group in ISO 8601 format.

## Timeouts

- `create` - (Default `30 minutes`) Used for creating the worker group.
- `update` - (Default `30 minutes`) Used for scaling or updating the worker group.
- `delete` - (Default `15 minutes`) Used for deleting the worker group.

## Import

A worker group is scoped to its cluster, so it is imported using `<cluster_id>/<worker_group_id>`:

```shell
terraform import vnpaycloud_kubernetes_worker_group.example <cluster_id>/<worker-group-id>
```

~> **Note:** Import requires the `<cluster_id>/<worker_group_id>` form because the worker group is scoped to its cluster. All attributes (including `flavor`, `volume_type`, `volume_size`, `ssh_key_id`, and `labels`) are read back on import.

### Managing the cluster's default worker group

The default worker group created with a cluster (via the `default_worker_*` attributes on `vnpaycloud_kubernetes_cluster`) is create-only through the cluster resource — those attributes are all ForceNew. To perform day-2 actions on it (resize, auto-scaling, labels, auto-healing) without recreating the cluster, import it into a `vnpaycloud_kubernetes_worker_group` resource and manage it there. Declare a matching resource block, then import it using the cluster ID and the default worker group's ID:

```shell
terraform import vnpaycloud_kubernetes_worker_group.default <cluster_id>/<default_worker_group_id>
```

Set the resource's `name`, `flavor`, `volume_type`, `volume_size`, and `ssh_key_id` to the default group's current values so the post-import plan is empty (these are ForceNew — a mismatch would replace the group).

~> **Auto-scaling:** Always set `min_workers` and `max_workers` when `auto_scaling = true` so the worker group has an explicit scaling range. Invalid ranges are rejected by the API.
