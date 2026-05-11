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

- `name` (String) The name of the robot account.
- `permission` (Block List) The permissions granted to the robot account across registry projects.
  - `registry_id` (String) The ID of the registry project.
  - `actions` (List of String) The actions granted on this registry project (e.g., `pull`, `push`).
- `expires_at` (String) The expiration timestamp of the robot account in ISO 8601 format. Empty string if the account does not expire.
- `enabled` (Boolean) Whether the robot account is currently enabled and able to authenticate.
- `created_at` (String) The timestamp when the robot account was created, in ISO 8601 format.
