---
page_title: "vnpaycloud_kubernetes_roles Data Source - VNPayCloud"
subcategory: "Kubernetes"
description: |-
  List the roles available in a Kubernetes cluster.
---

# vnpaycloud_kubernetes_roles (Data Source)

Use this data source to list the cluster roles available in a Kubernetes cluster. Each cluster is seeded with the default cluster roles `cluster-admin`, `admin`, `edit`, and `view`. This is useful for discovering the role names to use with the `vnpaycloud_kubernetes_rbac` resource.

This data source returns cluster roles only. Namespace-scoped roles created for `role_binding` RBAC entries are not returned here.

## Example Usage

```hcl
data "vnpaycloud_kubernetes_roles" "example" {
  cluster_id = vnpaycloud_kubernetes_cluster.example.id
}

output "role_names" {
  value = [for r in data.vnpaycloud_kubernetes_roles.example.roles : r.name]
}
```

## Schema

### Required

- `cluster_id` (String) The ID of the cluster to list roles for.

### Read-Only

- `roles` (List of Object) The cluster roles available in the cluster. Each element has:
  - `id` (String) The ID of the role.
  - `name` (String) The role name, e.g. `cluster-admin`, `admin`, `edit`, `view`.
  - `namespace` (String) Empty for cluster roles.
  - `is_cluster_role` (Boolean) Always `true` for this data source.
  - `status` (String) The current status of the role, lowercase.
  - `created_at` (String) The creation timestamp, in ISO 8601 format.
