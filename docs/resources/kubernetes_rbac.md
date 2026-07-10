---
page_title: "vnpaycloud_kubernetes_rbac Resource - VNPayCloud"
subcategory: "Kubernetes"
description: |-
  Binds a user to a role in a Kubernetes cluster (Kubernetes RBAC).
---

# vnpaycloud_kubernetes_rbac (Resource)

Manages a Kubernetes RBAC role binding: it binds a user to a role in a cluster. Each cluster is seeded with the default cluster roles `cluster-admin`, `admin`, `edit`, and `view`; reference one by name via the `role` attribute. Use the `vnpaycloud_kubernetes_roles` data source to discover the roles available in a cluster.

Two binding types are supported:

- `cluster_role_binding` (default) — grants the role across the whole cluster.
- `role_binding` — grants the role only within a single `namespace`. The `role` value must still be one of the cluster role names; VNPayCloud creates or reuses the namespace-scoped role for the binding.

All attributes are immutable — changing any of them replaces the binding.

## Example Usage

### Cluster-wide binding

```hcl
# Resolve the portal user id from an email via the vnpaycloud_users data source.
data "vnpaycloud_users" "by_email" {
  search = "alice@example.com"
}

resource "vnpaycloud_kubernetes_rbac" "admin" {
  cluster_id   = vnpaycloud_kubernetes_cluster.example.id
  user_id      = [for u in data.vnpaycloud_users.by_email.users : u.id if u.email == "alice@example.com"][0]
  role         = "cluster-admin"
  binding_type = "cluster_role_binding"
}
```

### Namespace-scoped binding

```hcl
resource "vnpaycloud_kubernetes_rbac" "editor" {
  cluster_id   = vnpaycloud_kubernetes_cluster.example.id
  user_id      = [for u in data.vnpaycloud_users.by_email.users : u.id if u.email == "alice@example.com"][0]
  role         = "edit"
  binding_type = "role_binding"
  namespace    = "team-a"
}
```

## Schema

### Required

- `cluster_id` (String, ForceNew) The ID of the cluster the binding applies to.
- `user_id` (String, ForceNew) The portal user id to bind the role to, e.g. `iaas.portal.usr.xxxxxxxx` (not the email). Look it up from an email with the `vnpaycloud_users` data source.
- `role` (String, ForceNew) The role name to bind. Must be one of the cluster roles seeded in the cluster (`cluster-admin`, `admin`, `edit`, `view`); an unknown name is rejected. List the valid names with the `vnpaycloud_kubernetes_roles` data source.

### Optional

- `binding_type` (String, ForceNew) The binding type: `cluster_role_binding` (default) or `role_binding`.
- `namespace` (String, ForceNew) The namespace to scope the binding to. Required when `binding_type` is `role_binding`, and must be omitted for `cluster_role_binding`. The namespace must **already exist** in the cluster; a non-existent namespace is rejected with a clear error.

### Read-Only

- `id` (String) The ID of the role binding.
- `kubernetes_role_id` (String) The ID of the resolved role in the cluster.
- `is_cluster_role_binding` (Boolean) Whether the binding is a cluster role binding.
- `name` (String) The display name of the role binding.
- `email` (String) The email of the bound user.
- `status` (String) The current status of the role binding, lowercase.
- `created_at` (String) The creation timestamp, in ISO 8601 format.

## Import

A role binding is scoped to its cluster, so it is imported using `<cluster_id>/<rbac_id>`:

```shell
terraform import vnpaycloud_kubernetes_rbac.example <cluster_id>/<rbac-id>
```
