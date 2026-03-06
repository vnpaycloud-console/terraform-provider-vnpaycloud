---
page_title: "vnpaycloud_registry_robot_account Resource - VNPayCloud"
subcategory: "Container Registry"
description: |-
  Manages a robot account for a container registry project within VNPayCloud.
---

# vnpaycloud_registry_robot_account (Resource)

Manages a robot account for a container registry project within VNPayCloud. Robot accounts provide automated, non-human access to registry projects for use in CI/CD pipelines, deployments, and other automation workflows.

~> **Note:** The `secret` attribute is only available immediately after creation. It is not stored remotely and cannot be retrieved later. Ensure you save the secret from the Terraform state or output immediately after `terraform apply`.

~> **Note:** All attributes are ForceNew. Any change to robot account configuration will destroy the existing account and create a new one.

## Example Usage

### CI/CD robot account with push and pull access

```hcl
resource "vnpaycloud_registry_project" "app" {
  name = "my-application"
}

resource "vnpaycloud_registry_robot_account" "ci" {
  registry_id     = vnpaycloud_registry_project.app.id
  name            = "ci-pipeline-robot"
  permissions     = ["push", "pull"]
  expires_in_days = 365
}

output "robot_secret" {
  value     = vnpaycloud_registry_robot_account.ci.secret
  sensitive = true
}
```

### Read-only robot account (no expiry)

```hcl
resource "vnpaycloud_registry_robot_account" "readonly" {
  registry_id = vnpaycloud_registry_project.app.id
  name        = "readonly-robot"
  permissions = ["pull"]
}
```

## Schema

### Required

- `registry_id` (String, ForceNew) The ID of the registry project this robot account belongs to. Changing this creates a new robot account.
- `name` (String, ForceNew) The name of the robot account. Changing this creates a new robot account.

### Optional

- `permissions` (List of String, ForceNew) A list of permissions granted to the robot account. Supported values include `push` and `pull`. If not specified, defaults to read-only access. Changing this creates a new robot account.
- `expires_in_days` (Number, ForceNew) The number of days until the robot account credentials expire. If not specified, the account does not expire. Changing this creates a new robot account.

### Read-Only

- `id` (String) The ID of the robot account.
- `secret` (String, Sensitive) The secret token for authenticating with the registry as this robot account. Only populated at creation time.
- `expires_at` (String) The expiration timestamp of the robot account credentials in ISO 8601 format. Empty if the account does not expire.
- `enabled` (Boolean) Whether the robot account is currently active and able to authenticate.
- `created_at` (String) The creation timestamp of the robot account in ISO 8601 format.

## Timeouts

- `create` - (Default `10 minutes`) Used for creating the robot account.
- `delete` - (Default `10 minutes`) Used for deleting the robot account.

## Import

Robot accounts can be imported using the `id`:

```shell
terraform import vnpaycloud_registry_robot_account.example <robot-account-id>
```

~> **Note:** Importing a robot account does not import the `secret`. The `secret` attribute will be empty after import.
