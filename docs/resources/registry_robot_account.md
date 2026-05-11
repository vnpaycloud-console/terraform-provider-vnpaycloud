---
page_title: "vnpaycloud_registry_robot_account Resource - VNPayCloud"
subcategory: "Container Registry"
description: |-
  Manages a system-level robot account for container registry within VNPayCloud.
---

# vnpaycloud_registry_robot_account (Resource)

Manages a system-level robot account for container registry within VNPayCloud. Robot accounts provide automated, non-human access to registry projects for use in CI/CD pipelines, deployments, and other automation workflows. A single robot account can be granted permissions across multiple registry projects.

~> **Note:** The `secret` attribute is only available immediately after creation. It is not stored remotely and cannot be retrieved later. Ensure you save the secret from the Terraform state or output immediately after `terraform apply`.

~> **Note:** All attributes are ForceNew. Any change to robot account configuration will destroy the existing account and create a new one.

## Example Usage

### CI/CD robot account with access to multiple projects

```hcl
resource "vnpaycloud_registry_project" "app" {
  name = "my-application"
}

resource "vnpaycloud_registry_project" "backend" {
  name = "backend-services"
}

resource "vnpaycloud_registry_robot_account" "ci" {
  name            = "ci-pipeline-robot"
  expires_in_days = 365

  permission {
    registry_id = vnpaycloud_registry_project.app.id
    actions     = ["push", "pull"]
  }

  permission {
    registry_id = vnpaycloud_registry_project.backend.id
    actions     = ["pull"]
  }
}

output "robot_secret" {
  value     = vnpaycloud_registry_robot_account.ci.secret
  sensitive = true
}
```

### Read-only robot account for a single project

```hcl
resource "vnpaycloud_registry_robot_account" "readonly" {
  name = "readonly-robot"

  permission {
    registry_id = vnpaycloud_registry_project.app.id
    actions     = ["pull"]
  }
}
```

## Schema

### Required

- `name` (String, ForceNew) The name of the robot account. Changing this creates a new robot account.
- `permission` (Block List, ForceNew) One or more permission blocks granting access to registry projects. Each block specifies a registry project and the actions allowed. Changing this creates a new robot account.

#### permission

- `registry_id` (String, Required) The ID of the registry project to grant access to.
- `actions` (List of String, Required) A list of actions granted on this registry project (e.g., `push`, `pull`).

### Optional

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
