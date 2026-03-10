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

### Read-Only

- `k8s_version` (String) The Kubernetes version running on the cluster (e.g., `v1.29.3`).
- `purpose` (String) The intended purpose or environment of the cluster (e.g., `development`, `staging`, `production`).
- `subnet_id` (String) The ID of the subnet in which the cluster's control plane and nodes are deployed.
- `cni_plugin` (String) The Container Network Interface (CNI) plugin used by the cluster (e.g., `calico`, `cilium`, `flannel`).
- `pod_cidr` (String) The CIDR block used for pod IP addresses within the cluster (e.g., `192.168.0.0/16`).
- `service_cidr` (String) The CIDR block used for Kubernetes service IP addresses (e.g., `10.96.0.0/12`).
- `cluster_size` (Number) The total number of worker nodes across all worker groups in the cluster.
- `zone` (String) The availability zone where the cluster is deployed.
- `api_endpoint` (String) The HTTPS endpoint URL for the Kubernetes API server.
- `private_ip` (String) The private IP address of the cluster's API server, accessible within the VPC.
- `status` (String) The current status of the cluster (e.g., `ACTIVE`, `CREATING`, `DELETING`, `ERROR`, `UPGRADING`).
- `created_at` (String) The timestamp when the cluster was created, in ISO 8601 format.
