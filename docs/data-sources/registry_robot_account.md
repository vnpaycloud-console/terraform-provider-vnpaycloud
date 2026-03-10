---
page_title: "vnpaycloud_registry_robot_account Data Source - VNPayCloud"
subcategory: "Container Registry"
description: |-
  Get information about a container registry robot account in VNPayCloud.
---

# vnpaycloud_registry_robot_account (Data Source)

Use this data source to get information about an existing container registry robot account. Robot accounts are service accounts used for automated access to a registry, typically in CI/CD pipelines.

~> **Note:** The robot account secret (password) is only available at creation time and cannot be retrieved via this data source. If you need to rotate the secret, you must create a new robot account.

## Example Usage

```hcl
data "vnpaycloud_registry_robot_account" "ci_robot" {
  id          = "ra-nop45678"
  registry_id = "rp-klm23456"
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
- `registry_id` (String) The ID of the registry project this robot account belongs to.

### Read-Only

- `name` (String) The name of the robot account.
- `permissions` (List of String) A list of permissions granted to the robot account (e.g., `pull`, `push`, `delete`).
- `expires_at` (String) The expiration timestamp of the robot account in ISO 8601 format. Empty string if the account does not expire.
- `enabled` (Boolean) Whether the robot account is currently enabled and able to authenticate.
- `created_at` (String) The timestamp when the robot account was created, in ISO 8601 format.
