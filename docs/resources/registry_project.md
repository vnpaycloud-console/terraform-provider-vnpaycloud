---
page_title: "vnpaycloud_registry_project Resource - VNPayCloud"
subcategory: "Container Registry"
description: |-
  Manages a container registry project within VNPayCloud.
---

# vnpaycloud_registry_project (Resource)

Manages a container registry project within VNPayCloud. A registry project is a namespace that organizes container image repositories. Projects can be configured as public or private and may have a storage quota applied.

~> **Note:** All attributes are ForceNew. Any change to project configuration will destroy the existing project and create a new one. Destroying a project will also remove all image repositories and tags within it.

## Example Usage

### Private registry project

```hcl
resource "vnpaycloud_registry_project" "app" {
  name          = "my-application"
  is_public     = false
  storage_limit = 10737418240  # 10 GiB in bytes
}
```

### Public registry project

```hcl
resource "vnpaycloud_registry_project" "public_images" {
  name      = "public-base-images"
  is_public = true
}
```

### Using with a robot account

```hcl
resource "vnpaycloud_registry_project" "app" {
  name      = "my-application"
  is_public = false
}

resource "vnpaycloud_registry_robot_account" "ci" {
  registry_id = vnpaycloud_registry_project.app.id
  name        = "ci-robot"
  permissions = ["push", "pull"]
}
```

## Schema

### Required

- `name` (String, ForceNew) The name of the registry project. Must be unique within the registry. Changing this creates a new project.

### Optional

- `is_public` (Boolean, ForceNew) Whether the project is publicly accessible for pulling images without authentication. Changing this creates a new project. Defaults to `false`.
- `storage_limit` (Number, ForceNew, Computed) The maximum storage quota for the project in bytes. If not specified, the platform default is applied. Changing this creates a new project.

### Read-Only

- `id` (String) The ID of the registry project.
- `storage_used` (Number) The current storage usage of the project in bytes.
- `repo_count` (Number) The number of image repositories within the project.
- `status` (String) The current status of the project (e.g., `active`, `disabled`).
- `created_at` (String) The creation timestamp of the project in ISO 8601 format.

## Timeouts

- `create` - (Default `10 minutes`) Used for creating the registry project.
- `delete` - (Default `10 minutes`) Used for deleting the registry project.

## Import

Registry projects can be imported using the `id`:

```shell
terraform import vnpaycloud_registry_project.example <project-id>
```
