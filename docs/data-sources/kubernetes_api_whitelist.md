---
page_title: "vnpaycloud_kubernetes_api_whitelist Data Source - VNPayCloud"
subcategory: "Kubernetes"
description: |-
  Read the API whitelist of a public Kubernetes cluster.
---

# vnpaycloud_kubernetes_api_whitelist (Data Source)

Reads the source-IP whitelist applied to the public Kubernetes API of a cluster, plus whether the cluster supports whitelisting at all. System addresses added by the platform are not included.

## Example Usage

```hcl
data "vnpaycloud_kubernetes_api_whitelist" "main" {
  cluster_id = vnpaycloud_kubernetes_cluster.main.id
}

output "allowed_cidrs" {
  value = data.vnpaycloud_kubernetes_api_whitelist.main.entry[*].cidr
}
```

## Schema

### Required

- `cluster_id` (String) The ID of a public Kubernetes cluster.

### Read-Only

- `entry` (List of Object) Allowed sources, each with `cidr` (String) and `note` (String). Empty means every IP is allowed.
- `endpoint` (String) The public IP of the Kubernetes API.
- `private_access` (Boolean) Whether the cluster is only reachable privately.
- `max_entries` (Number) Maximum number of entries the platform accepts.
- `supported` (Boolean) Whether this cluster can enforce a whitelist.
- `unsupported_reason` (String) Why `supported` is false, empty otherwise.
