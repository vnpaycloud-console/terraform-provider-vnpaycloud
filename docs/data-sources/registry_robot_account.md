---
page_title: "vnpaycloud_registry_robot_account Data Source - VNPayCloud"
subcategory: "Container Registry"
description: |-
  Get information about a container registry robot account in VNPayCloud.
---

# vnpaycloud_registry_robot_account (Data Source)

Use this data source to get information about an existing container registry robot account. Robot accounts are system-level service accounts that can have permissions across multiple registry projects, typically used in CI/CD pipelines.

~> **Note:** The robot account secret (password) is only available at creation time and cannot be retrieved via this data source. If you need to rotate the secret, you must create a new robot account.

## Example Usage

```hcl
data "vnpaycloud_registry_robot_account" "ci_robot" {
  id = "ra-nop45678"
}

output "robot_account_name" {
  value = data.vnpaycloud_registry_robot_account.ci_robot.name
}

output "robot_account_enabled" {
  value = data.vnpaycloud_registry_robot_account.ci_robot.enabled
}
```

## Schema

### Required (filter)

- `id` (String) The ID of the robot account.

### Read-Only

- `name` (String) Friendly robot name as supplied at creation time.
- `username` (String) **Full registry principal** in the form `bot$<YYMMDD>-<random>-<name>`. Pass this to `docker login -u`.
- `description` (String) Free-form description / label.
- `permission` (Block List) Permission set granted across registry projects.
  - `registry_id` (String) Registry project ID.
  - `actions` (List of String) Registry permissions in `resource:action` form (e.g. `repository:push`).
- `expires_in_days` (Number) Validity window in days. `-1` means never expire.
- `expires_at` (String) Expiration timestamp (RFC 3339). Empty when `expires_in_days = -1`.
- `enabled` (Boolean) Whether the account is currently enabled.
- `created_at` (String) Creation timestamp (RFC 3339 nanosecond).

~> The `secret` is **not** available via this data source. The registry only returns it at creation time. To rotate, delete and recreate the resource.
