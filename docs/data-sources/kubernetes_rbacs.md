---
page_title: "vnpaycloud_kubernetes_rbacs Data Source - VNPayCloud"
subcategory: "Kubernetes"
description: |-
  List the RBAC role bindings of a Kubernetes cluster.
---

# vnpaycloud_kubernetes_rbacs (Data Source)

Use this data source to list the RBAC role bindings of a Kubernetes cluster — the users bound to roles in the cluster.

## Example Usage

```hcl
data "vnpaycloud_kubernetes_rbacs" "example" {
  cluster_id = vnpaycloud_kubernetes_cluster.example.id
}

output "bound_users" {
  value = [for b in data.vnpaycloud_kubernetes_rbacs.example.rbacs : b.email]
}
```

## Schema

### Required

- `cluster_id` (String) The ID of the cluster to list role bindings for.

### Read-Only

- `rbacs` (List of Object) The role bindings of the cluster. Each element has:
  - `id` (String) The ID of the role binding.
  - `user_id` (String) The bound user id.
  - `email` (String) The email of the bound user.
  - `role` (String) The role name, e.g. `cluster-admin`.
  - `kubernetes_role_id` (String) The ID of the bound role.
  - `binding_type` (String) The binding type: `cluster_role_binding` or `role_binding`.
  - `is_cluster_role_binding` (Boolean) Whether the binding is a cluster role binding.
  - `namespace` (String) The namespace of the binding (empty for cluster role bindings).
  - `name` (String) The display name of the role binding.
  - `status` (String) The current status, lowercase.
  - `created_at` (String) The creation timestamp, in ISO 8601 format.
