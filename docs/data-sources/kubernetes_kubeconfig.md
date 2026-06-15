---
page_title: "vnpaycloud_kubernetes_kubeconfig Data Source - VNPayCloud"
subcategory: "Kubernetes"
description: |-
  Get the admin kubeconfig of a Kubernetes cluster in VNPayCloud.
---

# vnpaycloud_kubernetes_kubeconfig (Data Source)

Use this data source to fetch the admin kubeconfig of an existing managed Kubernetes cluster. The kubeconfig grants administrative access to the cluster's API server and can be written to a file or passed to the `kubernetes`/`helm` providers.

~> **Sensitive data** The kubeconfig contains administrative credentials. Both `kubeconfig_b64` and `content` are marked sensitive; avoid printing them in logs and protect any file you write them to.

## Example Usage

```hcl
data "vnpaycloud_kubernetes_cluster" "example" {
  name = "my-k8s-cluster"
}

data "vnpaycloud_kubernetes_kubeconfig" "example" {
  cluster_id = data.vnpaycloud_kubernetes_cluster.example.id
}

resource "local_sensitive_file" "kubeconfig" {
  content  = data.vnpaycloud_kubernetes_kubeconfig.example.content
  filename = "${path.module}/kubeconfig.yaml"
}
```

## Schema

### Required

- `cluster_id` (String) The ID of the Kubernetes cluster to fetch the kubeconfig for.

### Optional

- `is_private_access` (Boolean) When `true`, return a kubeconfig whose API server points at the cluster's private IP (for access within the VPC). Defaults to `false`.

### Read-Only

- `content` (String, Sensitive) The decoded kubeconfig YAML, ready to write to a file.
- `kubeconfig_b64` (String, Sensitive) The base64-encoded kubeconfig as returned by the API.
