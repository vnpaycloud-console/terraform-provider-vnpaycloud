---
page_title: "vnpaycloud_users Data Source - VNPayCloud"
subcategory: ""
description: |-
  List the users in your organization.
---

# vnpaycloud_users (Data Source)

Use this data source to list the users in your organization. Each user's `id` is the portal user id, which is the value to use as `user_id` in a `vnpaycloud_kubernetes_rbac` resource.

## Example Usage

```hcl
data "vnpaycloud_users" "all" {}

# Look up a specific user by email
data "vnpaycloud_users" "by_email" {
  search = "alice@example.com"
}

output "portal_user_id" {
  value = [for u in data.vnpaycloud_users.by_email.users : u.id if u.email == "alice@example.com"][0]
}
```

## Schema

### Optional

- `search` (String) An optional filter applied to the user name/email.

### Read-Only

- `users` (List of Object) The users in the organization. Each element has:
  - `id` (String) The portal user id (use as `user_id` in a role binding).
  - `email` (String) The user's email.
  - `display_name` (String) The user's display name.
  - `username` (String) The user's username.
