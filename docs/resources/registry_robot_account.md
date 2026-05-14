---
page_title: "vnpaycloud_registry_robot_account Resource - VNPayCloud"
subcategory: "Container Registry"
description: |-
  Manages a system-level robot account for container registry within VNPayCloud.
---

# vnpaycloud_registry_robot_account (Resource)

Manages a system-level robot account for the VNPayCloud container registry. Robot accounts provide automated, non-human access to registry projects for use in CI/CD pipelines, deployments, and other automation workflows. A single robot account can be granted permissions across multiple registry projects.

~> **Secret retention** — `secret` is only returned at creation time. It is not stored remotely; if you lose it the only recovery path is to recreate the robot account (which rotates the secret). Save it via `output { sensitive = true }` and pipe to a secret manager.

~> **Docker login uses `username`, not `name`** — the registry injects a prefix into the principal it accepts, so the field you pass to `docker login` is `username` (e.g. `bot$260513-nokjb3-ci`), not the friendly `name` you wrote in the HCL.

## Example Usage

### CI robot for one project (push & pull)

```hcl
resource "vnpaycloud_registry_project" "app" {
  name          = "my-application"
  is_public     = false
  storage_limit = "10737418240" # 10 GiB
}

resource "vnpaycloud_registry_robot_account" "ci" {
  name            = "ci-pipeline"
  description     = "Used by the GitHub Actions release workflow"
  expires_in_days = 365

  permission {
    registry_id = vnpaycloud_registry_project.app.id
    actions     = ["repository:push", "repository:pull"]
  }
}
```

### Read-only robot with no expiry

```hcl
resource "vnpaycloud_registry_robot_account" "readonly" {
  name            = "deploy-puller"
  expires_in_days = -1 # never expire

  permission {
    registry_id = vnpaycloud_registry_project.app.id
    actions     = ["repository:pull", "repository:list"]
  }
}
```

### Robot for multi-project scanning

```hcl
resource "vnpaycloud_registry_robot_account" "scanner" {
  name            = "vuln-scanner"
  expires_in_days = 30

  permission {
    registry_id = vnpaycloud_registry_project.app.id
    actions     = ["repository:pull", "artifact:read", "scan:create"]
  }

  permission {
    registry_id = vnpaycloud_registry_project.backend.id
    actions     = ["repository:pull", "artifact:read", "scan:create"]
  }
}
```

### Discover valid actions

```hcl
data "vnpaycloud_registry_permissions" "all" {}

# Use the catalogue to build dynamic permission sets
resource "vnpaycloud_registry_robot_account" "all_repo" {
  name            = "ci"
  expires_in_days = 30

  permission {
    registry_id = vnpaycloud_registry_project.app.id
    actions     = [
      for p in data.vnpaycloud_registry_permissions.all.permissions :
        p.key if p.resource == "repository"
    ]
  }
}
```

## Pushing & pulling images with Docker

The robot account exposes the two fields needed to authenticate to the registry, and the registry project exposes the namespace required to tag images:

| Step | Use |
|---|---|
| **Authenticate** | `username` (full registry principal `bot$...`) + `secret` |
| **Tag image** | `vnpaycloud_registry_project.<label>.namespace` |
| **Endpoint** | `vcr.vnpaycloud.vn` |

End-to-end example:

```hcl
output "registry_endpoint"  { value = "vcr.vnpaycloud.vn" }
output "robot_username"     { value = vnpaycloud_registry_robot_account.ci.username }
output "robot_secret"       { value = vnpaycloud_registry_robot_account.ci.secret; sensitive = true }
output "registry_namespace" { value = vnpaycloud_registry_project.app.namespace }
```

Then in a shell or CI step:

```bash
docker login vcr.vnpaycloud.vn \
  -u "$(terraform output -raw robot_username)" \
  -p "$(terraform output -raw robot_secret)"

NS=$(terraform output -raw registry_namespace)
docker tag myapp:v1 vcr.vnpaycloud.vn/$NS/myapp:v1
docker push           vcr.vnpaycloud.vn/$NS/myapp:v1
```

Or in the [Docker Terraform provider](https://registry.terraform.io/providers/kreuzwerker/docker):

```hcl
provider "docker" {
  registry_auth {
    address  = "vcr.vnpaycloud.vn"
    username = vnpaycloud_registry_robot_account.ci.username
    password = vnpaycloud_registry_robot_account.ci.secret
  }
}
```

## Schema

### Required

- `name` (String, ForceNew) Robot account name. Must contain only letters, digits, `.`, `_`, `-` (no spaces). Length 3–100. Must be unique. The full registry principal is exposed via [`username`](#username).
- `permission` (Block List, `MinItems: 1`) One or more permission blocks. Each block grants a set of actions on one registry project. Editable in-place — changes do not recreate the robot account.

#### permission block

- `registry_id` (String, Required) The ID of the registry project to grant access to. Must belong to the caller (foreign projects return `NotFound`).
- `actions` (List of String, `MinItems: 1`, Required) Each entry must match `^[a-z]+:[a-z-]+$` (e.g. `repository:push`, `artifact:read`). Use [`vnpaycloud_registry_permissions`](../data-sources/registry_permissions.md) to discover the valid list.

### Optional

- `description` (String) Free-form label. **Editable in-place** (changing it does not rotate the secret).
- `expires_in_days` (Number) Days until expiry. Must be `-1` (never expire) or a positive integer. Editable in-place. If omitted on import, the value is read back from the backend.

### Read-Only

- `id` (String) Robot account ID.
- `username` (String) **Full registry principal** in the form `bot$<YYMMDD>-<random>-<name>`. Pass this to `docker login -u`.
- `secret` (String, Sensitive) The registry secret. Only set on create — pulled back into state when the resource is first applied; subsequent `terraform refresh` does not change it. Importing a robot account leaves this empty.
- `expires_at` (String) Expiration time (RFC 3339 nanosecond). Empty when `expires_in_days = -1`.
- `enabled` (Boolean) Whether the account is currently active.
- `created_at` (String) Creation timestamp (RFC 3339 nanosecond).

## Update behaviour

| Field | Updateable in place? |
|---|---|
| `description` | ✅ |
| `expires_in_days` | ✅ — backend recomputes `expires_at` (pass `0` to keep the current window) |
| `permission` (list of blocks) | ✅ — full permission set is replaced |
| `name` | ❌ ForceNew — destroys & recreates (rotates `secret`) |

In-place updates **do not rotate** the secret. Any CI/CD job using the previous credentials keeps working.

## Validation errors

- `expires_in_days` outside `{-1, >0}` is rejected at `terraform plan` time.
- Actions not matching `<resource>:<action>` format are rejected at plan time.
- Empty `permission` list, or empty `actions` list inside a permission, are rejected at plan time.
- Actions whose `<resource>:<action>` pair is not in the registry catalogue are rejected by the proxy and include the full valid list in the error message.

## Timeouts

- `create` (Default `10 minutes`)
- `delete` (Default `10 minutes`)

## Import

```shell
terraform import vnpaycloud_registry_robot_account.ci <robot-account-id>
```

After import:

- `secret` will be **empty** — it cannot be retrieved from the registry backend. To rotate it, delete and recreate the resource (or refresh via the VNPayCloud console UI).
- All other fields including `username`, `description`, `expires_in_days`, and the full `permission` list are populated from the backend.
- Run `terraform plan` after import; it should report **No changes** when the HCL matches the imported state.
