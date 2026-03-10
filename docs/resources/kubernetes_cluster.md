---
page_title: "vnpaycloud_kubernetes_cluster Resource - VNPayCloud"
subcategory: "Kubernetes"
description: |-
  Manages a managed Kubernetes cluster within VNPayCloud.
---

# vnpaycloud_kubernetes_cluster (Resource)

Manages a managed Kubernetes cluster within VNPayCloud. The cluster includes a control plane managed by VNPayCloud and a default worker node group. Additional worker groups can be added using the `vnpaycloud_kubernetes_worker_group` resource.

~> **Note:** This resource does not support in-place updates. All attributes are ForceNew — any change will destroy the existing cluster and create a new one. Worker node group configuration (count, scaling) is managed separately via `vnpaycloud_kubernetes_worker_group`.

## Example Usage

### Basic cluster with default workers

```hcl
resource "vnpaycloud_keypair" "k8s" {
  name = "k8s-node-key"
}

resource "vnpaycloud_kubernetes_cluster" "main" {
  name                      = "production-cluster"
  subnet_id                 = "subnet-abc12345"
  default_worker_flavor     = "s.4c8r"
  k8s_version               = "v1.29"
  cluster_size              = "medium"
  default_worker_name       = "default-workers"
  default_worker_count      = 3
  default_worker_volume_type = "SSD"
  default_worker_volume_size = 50
  default_worker_ssh_key_id = vnpaycloud_keypair.k8s.id
  cni_plugin                = "cilium"
}

output "kubeconfig" {
  value     = vnpaycloud_kubernetes_cluster.main.kubeconfig
  sensitive = true
}
```

### Cluster with custom CNI and network settings

```hcl
resource "vnpaycloud_kubernetes_cluster" "custom_net" {
  name                  = "custom-network-cluster"
  subnet_id             = "subnet-xyz98765"
  default_worker_flavor = "s.8c16r"
  purpose               = "production"
  cni_plugin            = "calico"
  pod_cidr              = "10.244.0.0/16"
  service_cidr          = "10.96.0.0/12"
  default_worker_count  = 2
}
```

## Schema

### Required

- `name` (String, ForceNew) The name of the Kubernetes cluster. Changing this creates a new cluster.
- `subnet_id` (String, ForceNew) The ID of the subnet where the cluster nodes will be deployed. Changing this creates a new cluster.
- `default_worker_flavor` (String, ForceNew) The flavor (instance type) for the default worker node group. Changing this creates a new cluster.

### Optional

- `k8s_version` (String, ForceNew, Computed) The Kubernetes version to deploy (e.g., `v1.29`, `v1.28`). If not specified, the latest stable version is used. Changing this creates a new cluster.
- `purpose` (String, ForceNew) The intended purpose of the cluster, used for grouping or labeling (e.g., `production`, `staging`). Changing this creates a new cluster.
- `private_gw_id` (String, ForceNew) The ID of the private gateway to use for the cluster's outbound traffic. Changing this creates a new cluster.
- `cni_plugin` (String, ForceNew, Computed) The Container Network Interface plugin to use for pod networking. Valid values are `calico` and `cilium`. If not specified, defaults to the platform default. Changing this creates a new cluster.
- `pod_cidr` (String, ForceNew, Computed) The CIDR block for pod IP addresses. If not specified, a default is assigned. Changing this creates a new cluster.
- `service_cidr` (String, ForceNew, Computed) The CIDR block for Kubernetes service IP addresses. If not specified, a default is assigned. Changing this creates a new cluster.
- `cluster_size` (String, ForceNew, Computed) The control plane size. Valid values are `small`, `medium`, `large`, `extra_large`. If not specified, a default is assigned based on worker count. Changing this creates a new cluster.
- `default_worker_name` (String, ForceNew) The name for the default worker node group. Changing this creates a new cluster.
- `default_worker_count` (Number, ForceNew) The initial number of worker nodes in the default group. Defaults to `1`. Changing this creates a new cluster.
- `default_worker_volume_type` (String, ForceNew) The volume type for default worker node root disks (e.g., `SSD`, `HDD`). Changing this creates a new cluster.
- `default_worker_volume_size` (Number, ForceNew) The root disk size in gigabytes for default worker nodes. Changing this creates a new cluster.
- `default_worker_ssh_key_id` (String, ForceNew) The ID of the SSH key pair to inject into the default worker nodes. Changing this creates a new cluster.

### Read-Only

- `id` (String) The ID of the Kubernetes cluster.
- `zone` (String) The availability zone where the cluster control plane is deployed.
- `api_endpoint` (String) The HTTPS URL of the Kubernetes API server endpoint.
- `private_ip` (String) The private IP address of the Kubernetes API server.
- `status` (String) The current status of the cluster (e.g., `ACTIVE`, `CREATING`, `ERROR`).
- `created_at` (String) The creation timestamp of the cluster in ISO 8601 format.
- `kubeconfig` (String, Sensitive) The kubeconfig file content for authenticating with the cluster using `kubectl`.

## Timeouts

- `create` - (Default `30 minutes`) Used for creating the Kubernetes cluster.
- `delete` - (Default `15 minutes`) Used for deleting the Kubernetes cluster.

## Import

Kubernetes clusters can be imported using the `id`:

```shell
terraform import vnpaycloud_kubernetes_cluster.example <cluster-id>
```
