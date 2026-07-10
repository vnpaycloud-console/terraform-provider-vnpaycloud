---
page_title: "vnpaycloud_kubernetes_cluster Data Source - VNPayCloud"
subcategory: "Kubernetes"
description: |-
  Get information about a Kubernetes cluster in VNPayCloud.
---

# vnpaycloud_kubernetes_cluster (Data Source)

Use this data source to get information about an existing managed Kubernetes cluster, including its API endpoint, network configuration, and current status. This is useful for configuring other resources that depend on the cluster.

## Example Usage

```hcl
data "vnpaycloud_kubernetes_cluster" "example" {
  name = "my-k8s-cluster"
}

output "cluster_api_endpoint" {
  value = data.vnpaycloud_kubernetes_cluster.example.api_endpoint
}

output "cluster_version" {
  value = data.vnpaycloud_kubernetes_cluster.example.k8s_version
}
```

```hcl
data "vnpaycloud_kubernetes_cluster" "by_id" {
  id = "k8s-qrs56789"
}
```

## Schema

### Optional (filter)

- `id` (String) The ID of the Kubernetes cluster.
- `name` (String) The name of the Kubernetes cluster.

~> **Deterministic lookup:** Provide exactly one of `id` or `name`. Prefer `id` because cluster names are not guaranteed to be globally unique. If `name` matches multiple clusters, the lookup fails.

### Read-Only

- `k8s_version` (String) The Kubernetes version running on the cluster (e.g., `1.30.14`).
- `purpose` (String) The intended purpose or environment of the cluster (e.g., `development`, `staging`, `production`).
- `private_gw_id` (String) The ID of the private gateway associated with the cluster, if any (present only for private clusters).
- `subnet_id` (String) The ID of the subnet in which the cluster's control plane and nodes are deployed.
- `cni_plugin` (String) The Container Network Interface (CNI) plugin used by the cluster. One of `calico` or `cilium`.
- `pod_cidr` (String) The CIDR block used for pod IP addresses within the cluster (e.g., `192.168.0.0/16`).
- `service_cidr` (String) The CIDR block used for Kubernetes service IP addresses (e.g., `10.96.0.0/12`).
- `cluster_size` (String) The control plane size, such as `small`, `medium`, `large`, or `extra_large`.
- `zone` (String) The availability zone where the cluster is deployed.
- `api_endpoint` (String) The IP/URL of the Kubernetes API server. Empty for private clusters.
- `private_ip` (String) The private IP address of the cluster's API server, accessible within the VPC.
- `status` (String) The current status of the cluster, lowercase (e.g., `active`, `creating`, `deleting`, `error`, `failed`).
- `created_at` (String) The timestamp when the cluster was created, in ISO 8601 format.
