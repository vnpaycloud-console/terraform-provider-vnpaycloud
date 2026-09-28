---
page_title: "vnpaycloud_maas_role_binding Data Source - VNPayCloud"
subcategory: "Monitoring as a Service"
description: |-
  Looks up a single monitoring workspace role binding by ID or portal user.
---

# vnpaycloud_maas_role_binding (Data Source)

Looks up one monitoring workspace role binding, by `id` or by `portal_user_id`.

## Example Usage

```hcl
data "vnpaycloud_users" "all" {}

data "vnpaycloud_maas_role_binding" "analyst" {
  portal_user_id = one([
    for u in data.vnpaycloud_users.all.users : u.id if u.email == "analyst@example.com"
  ])
}
```

## Schema

### Optional

At least one of `id` or `portal_user_id` is required.

- `id` (String) — Binding ID.
- `portal_user_id` (String) — Portal user the binding belongs to.

### Read-Only

- `role` (String) — `admin`, `editor` or `viewer`. `admin` appears for the organization owner.
- `status` (String) — Lifecycle status.
- `created_at` (String) — Creation timestamp (RFC3339).
