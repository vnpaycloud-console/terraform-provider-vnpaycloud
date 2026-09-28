---
page_title: "vnpaycloud_maas_role_binding Resource - VNPayCloud"
subcategory: "Monitoring as a Service"
description: |-
  Grants a portal user a role on the organization's monitoring workspace.
---

# vnpaycloud_maas_role_binding (Resource)

Grants a portal user a role on the organization's monitoring workspace, where dashboards and
explored data live. One binding per user.

## Example Usage

```hcl
data "vnpaycloud_users" "all" {}

locals {
  analyst_id = one([
    for u in data.vnpaycloud_users.all.users : u.id if u.email == "analyst@example.com"
  ])
}

resource "vnpaycloud_maas_role_binding" "analyst" {
  portal_user_id = local.analyst_id
  role           = "viewer"
}
```

## Schema

### Required

- `portal_user_id` (String, Forces new resource) — Portal user receiving the role. Use the `vnpaycloud_users` data source to look one up by email.
- `role` (String) — One of `editor`, `viewer` (validated at plan time). Can be updated in place. The organization owner already holds full access and is not managed through this resource.

### Read-Only

- `id` (String) — Binding ID.
- `status` (String) — Lifecycle status (`active`, `creating`, `deleting`, `deleted`, `error`).
- `created_at` (String) — Creation timestamp (RFC3339).

## Import

```shell
terraform import vnpaycloud_maas_role_binding.analyst <binding-id>
```
