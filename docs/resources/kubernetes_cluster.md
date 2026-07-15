---
page_title: "vnpaycloud_kubernetes_cluster Resource - VNPayCloud"
subcategory: "Kubernetes"
description: |-
  Manages a managed Kubernetes cluster within VNPayCloud.
---

# vnpaycloud_kubernetes_cluster (Resource)

Manages a managed Kubernetes cluster within VNPayCloud. The cluster includes a control plane managed by VNPayCloud and a default worker node group. Additional worker groups can be added using the `vnpaycloud_kubernetes_worker_group` resource.

~> **Note:** Only `k8s_version` (control-plane upgrade) and `cluster_size` (deployment-size change) can be updated in place. All other attributes are ForceNew — changing them destroys and recreates the cluster. Worker node group configuration (count, scaling) is managed separately via `vnpaycloud_kubernetes_worker_group`. The `default_worker_*` attributes here are create-only; to change the default worker group after creation, import it into a `vnpaycloud_kubernetes_worker_group` resource (see that resource's *Managing the cluster's default worker group*).

## Cluster types: public vs private

Every cluster is one of two types, chosen at creation via the `private_gw_id` attribute:

| | **Public cluster** | **Private cluster** |
|---|---|---|
| How to select | `private_gw_id` **omitted** (default) | `private_gw_id` set to a `vnpaycloud_private_gateway` |
| API server exposure | Public endpoint | Reachable only within the VPC (private) |
| `api_endpoint` | Public IP/URL of the API server | **Empty** |
| `private_ip` | In-VPC API address | In-VPC API address |
| kubeconfig access | `kubeconfig` attribute, or `vnpaycloud_kubernetes_kubeconfig` data source | `vnpaycloud_kubernetes_kubeconfig` data source with `is_private_access = true` |

The type is **fixed at creation**: `private_gw_id` is ForceNew, so switching a public cluster to private (or back) recreates the cluster. Both types share the same [prerequisites](#prerequisites) below.

## Prerequisites

A cluster can only be created on a subnet that is prepared for Kubernetes. Before creating a cluster, ensure:

- The subnet has `used_by_k8s = true`.
- The VPC has **SNAT enabled** so nodes can reach the internet. This requires an internet gateway, a default route (`0.0.0.0/0`) to it, and a `vnpaycloud_subnet_snat` on the subnet (see the example below).
- `pod_cidr` and `service_cidr` do **not** overlap each other, the subnet CIDR, any other subnet in the project, or the pod CIDR of another live cluster in the same zone.
- `default_worker_volume_size` is between **50** and **200** (GB).
- `default_worker_name`, `default_worker_volume_type`, and `default_worker_ssh_key_id` are set.

## Example Usage

### Public cluster

`private_gw_id` is omitted, so the cluster's API server is exposed on a public endpoint.

```hcl
resource "vnpaycloud_vpc" "vpc" {
  name = "k8s-vpc"
  cidr = "10.10.0.0/16"
}

resource "vnpaycloud_subnet" "subnet" {
  name        = "k8s-subnet"
  vpc_id      = vnpaycloud_vpc.vpc.id
  cidr        = "10.10.1.0/24"
  used_by_k8s = true
}

# ---- SNAT chain (required so cluster nodes can reach the internet) ----
resource "vnpaycloud_internet_gateway" "igw" {
  name       = "k8s-igw"
  vpc_id     = vnpaycloud_vpc.vpc.id
  depends_on = [vnpaycloud_subnet.subnet]
}

resource "vnpaycloud_route_table" "internet" {
  vpc_id      = vnpaycloud_vpc.vpc.id
  dest_cidr   = "0.0.0.0/0"
  target_id   = vnpaycloud_internet_gateway.igw.id
  target_type = "internet_gateway"
}

resource "vnpaycloud_floating_ip" "snat_ip" {
  vpc_id     = vnpaycloud_vpc.vpc.id
  depends_on = [vnpaycloud_internet_gateway.igw]
}

resource "vnpaycloud_subnet_snat" "snat" {
  subnet_id      = vnpaycloud_subnet.subnet.id
  floating_ip_id = vnpaycloud_floating_ip.snat_ip.id
  depends_on     = [vnpaycloud_route_table.internet]
}

resource "vnpaycloud_keypair" "k8s" {
  name = "k8s-node-key"
}

resource "vnpaycloud_kubernetes_cluster" "main" {
  name                       = "production-cluster"
  k8s_version                = "1.30.14"
  subnet_id                  = vnpaycloud_subnet.subnet.id
  default_worker_flavor      = "a-pro-small.2x2"
  cluster_size               = "small"
  cni_plugin                 = "cilium"
  pod_cidr                   = "10.244.0.0/16"
  service_cidr               = "10.96.0.0/12"
  default_worker_name        = "default-workers"
  default_worker_count       = 1
  default_worker_volume_type = "c1-standard"
  default_worker_volume_size = 50
  default_worker_ssh_key_id  = vnpaycloud_keypair.k8s.id

  depends_on = [vnpaycloud_subnet_snat.snat]
}

output "kubeconfig" {
  value     = vnpaycloud_kubernetes_cluster.main.kubeconfig
  sensitive = true
}
```

### Private cluster

Set `private_gw_id` to a `vnpaycloud_private_gateway` to make the cluster **private**: the Kubernetes API is reachable only within the VPC (through the private gateway) and `api_endpoint` will be empty. Retrieve access via the kubeconfig data source with `is_private_access = true`. The SNAT/network prerequisites above still apply.

```hcl
resource "vnpaycloud_private_gateway" "pgw" {
  name = "k8s-private-gw"
}

resource "vnpaycloud_kubernetes_cluster" "private" {
  name                       = "private-cluster"
  private_gw_id              = vnpaycloud_private_gateway.pgw.id
  subnet_id                  = vnpaycloud_subnet.subnet.id
  default_worker_flavor      = "a-pro-small.2x2"
  cni_plugin                 = "cilium"
  pod_cidr                   = "10.245.0.0/16"
  service_cidr               = "10.96.0.0/12"
  default_worker_name        = "default-workers"
  default_worker_volume_type = "c1-standard"
  default_worker_volume_size = 50
  default_worker_ssh_key_id  = vnpaycloud_keypair.k8s.id

  depends_on = [vnpaycloud_subnet_snat.snat]
}
```

## Schema

### Required

- `name` (String, ForceNew) The name of the Kubernetes cluster. Must be unique within the project. Changing this creates a new cluster.
- `subnet_id` (String, ForceNew) The ID of the subnet where the cluster nodes will be deployed. The subnet must have `used_by_k8s = true`. Changing this creates a new cluster.
- `default_worker_flavor` (String, ForceNew) The flavor (instance type) for the default worker node group, e.g. `a-pro-small.2x2`. List available flavors with the `vnpaycloud_flavors` data source. Changing this creates a new cluster.
- `default_worker_name` (String, ForceNew) The name for the default worker node group. Changing this creates a new cluster.
- `default_worker_volume_type` (String, ForceNew) The volume type for default worker node root disks, e.g. `c1-standard`. List available volume types with the `vnpaycloud_volume_types` data source. Changing this creates a new cluster.
- `default_worker_ssh_key_id` (String, ForceNew) The ID of the SSH key pair to inject into the default worker nodes. Changing this creates a new cluster.

### Optional

- `k8s_version` (String, Computed) The Kubernetes version to deploy, e.g. `1.30.14` (no `v` prefix). Must be an enabled version whose node image is available in the target zone — list the available versions (and the default) with the `vnpaycloud_kubernetes_versions` data source. If omitted, the default version is used — set this explicitly for deterministic creates. **Updatable in place:** changing this to a newer, supported version triggers an in-place control-plane upgrade (no cluster recreation). The target must be on the current version's allowed upgrade path — typically the next supported minor version — so upgrade one step at a time; downgrades and version skips are rejected. Note that the set of versions offered for new clusters (the `vnpaycloud_kubernetes_versions` data source) is not necessarily the same as the valid upgrade target from a given version. Upgrading the control plane does not upgrade existing worker groups.
- `purpose` (String, ForceNew) A free-form label describing the intended purpose of the cluster (e.g. `production`, `staging`). Changing this creates a new cluster.
- `private_gw_id` (String, ForceNew) The ID of a `vnpaycloud_private_gateway`. When set, the cluster is **private** — its API server is reachable only within the VPC (through the private gateway) and `api_endpoint` is empty. Changing this creates a new cluster.
- `cni_plugin` (String, ForceNew, Computed) The Container Network Interface plugin. Valid values are `calico` and `cilium`. Note that `calico` may not be available for every Kubernetes version; `cilium` is recommended. Changing this creates a new cluster.
- `pod_cidr` (String, ForceNew, Computed) The CIDR block for pod IP addresses. A malformed value is rejected at plan time; overlap checks (against `service_cidr`, the subnet, or another live cluster's pod CIDR in the same zone) are enforced by the backend. Changing this creates a new cluster.
- `service_cidr` (String, ForceNew, Computed) The canonical CIDR block for Kubernetes service IP addresses, for example `10.96.0.0/12`. A malformed value is rejected at plan time; the non-overlap-with-`pod_cidr` rule is enforced by the backend. Changing this creates a new cluster.
- `cluster_size` (String, Computed) The control plane size (deployment size). Valid values are `small`, `medium`, `large`, `extra_large` — each raises the supported worker/pod limits. If not specified, a default is assigned based on worker count. **Updatable in place:** changing this resizes the control plane (both up and down) without recreating the cluster.
- `default_worker_count` (Number, ForceNew) The initial number of worker nodes in the default group. Must be at least `1`. Defaults to `1`. Changing this creates a new cluster.
- `default_worker_volume_size` (Number, ForceNew) The root disk size in gigabytes for default worker nodes. Must be between `50` and `200` (validated at plan time). Changing this creates a new cluster.

~> **Name length:** The cluster name plus the default worker group name have a combined length limit. Keep both short and use lowercase letters, numbers, and hyphens.

### Read-Only

- `id` (String) The ID of the Kubernetes cluster.
- `zone` (String) The availability zone where the cluster control plane is deployed.
- `api_endpoint` (String) The IP/URL of the Kubernetes API server endpoint. Empty for private clusters (see `private_gw_id`).
- `private_ip` (String) The private IP address of the Kubernetes API server.
- `status` (String) The current status of the cluster, lowercase (e.g. `active`, `creating`, `error`, `failed`).
- `created_at` (String) The creation timestamp of the cluster in ISO 8601 format.
- `kubeconfig` (String, Sensitive) The kubeconfig file content for authenticating with the cluster using `kubectl`.

## Timeouts

- `create` - (Default `30 minutes`) Used for creating the Kubernetes cluster. Cluster provisioning can take a while; if creates time out, raise this with a `timeouts { create = "90m" }` block.
- `update` - (Default `30 minutes`) Used for in-place version upgrades and deployment-size changes. A control-plane upgrade or resize can take 15–25 minutes; if you change both `k8s_version` and `cluster_size` in one apply (applied sequentially) or hit slower zones, raise this with a `timeouts { update = "60m" }` block.
- `delete` - (Default `15 minutes`) Used for deleting the Kubernetes cluster.

## Import

Kubernetes clusters can be imported using the `id`:

```shell
terraform import vnpaycloud_kubernetes_cluster.example <cluster-id>
```

~> **Note:** `k8s_version`, `cluster_size`, and the network settings are read back from the API on import. The `default_worker_*` attributes are not read back, so set them in configuration to match (they are ignored on import to avoid spurious diffs).
