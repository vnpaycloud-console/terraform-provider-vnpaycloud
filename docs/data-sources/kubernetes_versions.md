---
page_title: "vnpaycloud_kubernetes_versions Data Source - VNPayCloud"
subcategory: "Kubernetes"
description: |-
  List the Kubernetes versions available for new clusters in VNPayCloud.
---

# vnpaycloud_kubernetes_versions (Data Source)

Use this data source to list the Kubernetes versions currently available (enabled) for creating a `vnpaycloud_kubernetes_cluster`. This avoids hardcoding a version string that may later be disabled: reference the default, or pick from the list.

## Example Usage

```hcl
data "vnpaycloud_kubernetes_versions" "available" {}

# Use the default version for a new cluster.
resource "vnpaycloud_kubernetes_cluster" "main" {
  name        = "production-cluster"
  k8s_version = data.vnpaycloud_kubernetes_versions.available.default_version
  # ... other required attributes ...
}

output "all_versions" {
  value = [for v in data.vnpaycloud_kubernetes_versions.available.versions : v.version]
}
```

## Schema

### Read-Only

- `versions` (List of Object) The Kubernetes versions available for new clusters. Each element has:
  - `version` (String) The version string, e.g. `1.30.14` (no `v` prefix).
  - `is_default` (Boolean) Whether this is the default version assigned when `k8s_version` is omitted on a cluster.
- `default_version` (String) The version marked as default (empty if none is flagged), for convenient reference.
