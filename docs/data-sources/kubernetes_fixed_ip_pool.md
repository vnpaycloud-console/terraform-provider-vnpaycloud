---
page_title: "vnpaycloud_kubernetes_fixed_ip_pool Data Source - VNPayCloud"
subcategory: "Kubernetes"
description: |-
  Read the fixed IP pool of a Kubernetes control plane or worker group.
---

# vnpaycloud_kubernetes_fixed_ip_pool (Data Source)

Reads the fixed IP pool of a cluster that was created with fixed IPs: which slot holds which IP, whether the slot is claimed, and by which machine. Without `worker_group_id` the control-plane pool is returned; with it, the pool of that worker group.

The lookup fails for clusters or worker groups that do not use fixed IPs.

## Example Usage

```hcl
data "vnpaycloud_kubernetes_fixed_ip_pool" "control_plane" {
  cluster_id = vnpaycloud_kubernetes_cluster.main.id
}

data "vnpaycloud_kubernetes_fixed_ip_pool" "workers" {
  cluster_id      = vnpaycloud_kubernetes_cluster.main.id
  worker_group_id = vnpaycloud_kubernetes_worker_group.extra.id
}

output "free_worker_ips" {
  value = [for s in data.vnpaycloud_kubernetes_fixed_ip_pool.workers.slots : s.ip_address if !s.claimed]
}
```

## Schema

### Required

- `cluster_id` (String) The ID of the Kubernetes cluster.

### Optional

- `worker_group_id` (String) The ID of a worker group in the cluster. When omitted the control-plane pool is read.

### Read-Only

- `id` (String) `<cluster_id>` or `<cluster_id>/<worker_group_id>`.
- `name` (String) The name of the pool on the management cluster.
- `slots` (List of Object) One entry per slot, in pool order:
  - `name` (String) The slot name.
  - `ip_address` (String) The IP reserved for the slot.
  - `claimed` (Boolean) Whether a machine currently holds the slot.
  - `machine_name` (String) The machine holding the slot, empty when free.
