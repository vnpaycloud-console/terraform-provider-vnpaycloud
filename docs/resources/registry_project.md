---
page_title: "vnpaycloud_registry_project Resource - VNPayCloud"
subcategory: "Container Registry"
description: |-
  Manages a container registry project within VNPayCloud.
---

# vnpaycloud_registry_project (Resource)

Manages a container registry project within VNPayCloud. A registry project is a namespace that organizes container image repositories. Projects can be configured as public or private and have a storage quota applied. `is_public` and `storage_limit` are editable in place; `name` is fixed at creation time.

## Example Usage

### Private registry project (1 GiB quota)

```hcl
resource "vnpaycloud_registry_project" "app" {
  name          = "my-application"
  is_public     = false
  storage_limit = "1073741824" # 1 GiB
}
```

### Public registry project

```hcl
resource "vnpaycloud_registry_project" "public" {
  name          = "public-base-images"
  is_public     = true
  storage_limit = "5368709120" # 5 GiB
}
```

### Used with a robot account for CI

```hcl
resource "vnpaycloud_registry_project" "app" {
  name          = "my-application"
  is_public     = false
  storage_limit = "10737418240" # 10 GiB
}

resource "vnpaycloud_registry_robot_account" "ci" {
  name            = "ci-pipeline"
  expires_in_days = 365

  permission {
    registry_id = vnpaycloud_registry_project.app.id
    actions     = ["repository:push", "repository:pull"]
  }
}
```

## Pushing & pulling images with Docker

Use the `namespace` attribute to tag images correctly, and the robot account `username` / `secret` to authenticate:

```hcl
output "registry_namespace" {
  value = vnpaycloud_registry_project.app.namespace
}

output "robot_username" {
  value = vnpaycloud_registry_robot_account.ci.username
}

output "robot_secret" {
  value     = vnpaycloud_registry_robot_account.ci.secret
  sensitive = true
}
```

Then from a shell:

```bash
# 1. Login (run once per machine)
docker login vcr.vnpaycloud.vn \
  -u "$(terraform output -raw robot_username)" \
  -p "$(terraform output -raw robot_secret)"

# 2. Tag and push
NS=$(terraform output -raw registry_namespace)
docker tag myapp:v1 vcr.vnpaycloud.vn/$NS/myapp:v1
docker push           vcr.vnpaycloud.vn/$NS/myapp:v1
```

The `namespace` is `"{org_id_short}-{project_name}"` — the provider computes it for you so you never need to look up the organization ID manually.

## Schema

### Required

- `name` (String, ForceNew) Project name. Must contain only letters, digits, `.`, `_`, `-` (no spaces). Length 3–250. Unique within the organization.
- `storage_limit` (String) Storage quota in bytes (as a string because the value exceeds 32-bit int). Must be `> 0`. Editable in-place. Cannot be reduced below the currently used storage.

### Optional

- `is_public` (Boolean) Whether the project is publicly accessible (anonymous `docker pull`). Editable in-place. Defaults to `false`.

### Read-Only

- `id` (String) Project ID.
- `namespace` (String) **Registry namespace** in the form `"{org_id_short}-{name}"`. Use this as the second path segment when tagging images, e.g. `vcr.vnpaycloud.vn/<namespace>/<repo>:<tag>`.
- `storage_used` (Number) Current storage usage in bytes.
- `repo_count` (Number) Number of repositories inside the project.
- `status` (String) Project status (`active`, `deleted`, …).
- `created_at` (String) RFC 3339 (nanosecond) timestamp.

## Update behaviour

- Editing `is_public` or `storage_limit` triggers an **in-place update** (no recreation, no secret rotation for robot accounts attached to the project).
- Editing `name` is **ForceNew** — the project is destroyed and recreated, which also deletes any images inside it.
- The new `storage_limit` must be strictly greater than `storage_used` and must fit inside the caller's available user-level storage quota; otherwise the update returns `InvalidArgument` with the specific bytes available.

## Destruction

- Deleting a project with one or more repositories returns `FailedPrecondition` (`"project still contains N repositor(ies); remove all images before deleting"`). Before running `terraform destroy`, delete every repository belonging to this project from the **VNPayCloud console UI** (registry section). `terraform destroy` does **not** cascade into repositories.

## Timeouts

- `create` (Default `10 minutes`)
- `delete` (Default `10 minutes`)

## Import

```shell
terraform import vnpaycloud_registry_project.app <project-id>
```

After import, run `terraform plan` to verify state matches the configuration. Imported projects retain their `namespace` value — no extra lookup needed.
