---
page_title: "vnpaycloud_kubernetes_clusters Data Source - VNPayCloud"
subcategory: "Kubernetes"
description: |-
  List all Kubernetes clusters in VNPayCloud.
---

# vnpaycloud_kubernetes_clusters (Data Source)

Use this data source to list all Kubernetes clusters in the current project.

## Example Usage

```hcl
data "vnpaycloud_kubernetes_clusters" "all" {}

output "all_cluster_names" {
  value = data.vnpaycloud_kubernetes_clusters.all.clusters[*].name
}

output "active_cluster_endpoints" {
  value = {
    for cluster in data.vnpaycloud_kubernetes_clusters.all.clusters :
    cluster.name => cluster.api_endpoint if cluster.status == "ACTIVE"
  }
}

output "cluster_versions" {
  value = {
    for cluster in data.vnpaycloud_kubernetes_clusters.all.clusters :
    cluster.name => cluster.k8s_version
  }
}
```

## Schema

### Read-Only

- `clusters` (List of Object) List of Kubernetes clusters. Each element contains:
  - `id` (String) The unique identifier of the Kubernetes cluster.
  - `name` (String) The name of the Kubernetes cluster.
  - `k8s_version` (String) The Kubernetes version running on this cluster (e.g., `1.28.5`).
  - `subnet_id` (String) The ID of the subnet where the cluster control plane is deployed.
  - `cluster_size` (Number) The total number of worker nodes in the cluster.
  - `zone` (String) The availability zone where the cluster is deployed.
  - `api_endpoint` (String) The HTTPS endpoint of the Kubernetes API server.
  - `status` (String) The current status of the cluster (e.g., `ACTIVE`, `PROVISIONING`, `ERROR`).
  - `created_at` (String) The timestamp when the cluster was created, in ISO 8601 format.
