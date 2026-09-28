---
page_title: "vnpaycloud_maas_role_bindings Data Source - VNPayCloud"
subcategory: "Monitoring as a Service"
description: |-
  Lists role bindings on the organization's monitoring workspace.
---

# vnpaycloud_maas_role_bindings (Data Source)

Lists who has access to the organization's monitoring workspace.

## Example Usage

```hcl
data "vnpaycloud_maas_role_bindings" "editors" {
  role = "editor"
}

output "editor_user_ids" {
  value = [for b in data.vnpaycloud_maas_role_bindings.editors.role_bindings : b.portal_user_id]
}
```

## Schema

### Optional

- `portal_user_id` (String) — Return only bindings for this user.
- `role` (String) — Return only bindings with this role: `admin`, `editor` or `viewer` (validated at plan time).

### Read-Only

- `role_bindings` (List) — Matching bindings, each with:
  - `id` (String)
  - `portal_user_id` (String)
  - `role` (String)
  - `status` (String)
  - `created_at` (String)
