---
page_title: "vnpaycloud_registry_permissions Data Source - VNPayCloud"
subcategory: "Container Registry"
description: |-
  Returns the catalogue of (resource, action) pairs the registry accepts as robot account permissions.
---

# vnpaycloud_registry_permissions (Data Source)

Returns the full catalogue of permissions the registry accepts when building a [`vnpaycloud_registry_robot_account`](../resources/registry_robot_account.md). Each entry is a `(resource, action)` pair and the convenience field `key` already in the `resource:action` form that the resource expects.

Use it to:

- Discover valid actions without leaving Terraform.
- Build permission sets dynamically (e.g. "all repository actions" or "all read-only actions").
- Avoid hard-coding the action list — the catalogue auto-updates when the registry adds new actions.

## Example Usage

### Print every valid permission

```hcl
data "vnpaycloud_registry_permissions" "all" {}

output "valid_permissions" {
  value = [for p in data.vnpaycloud_registry_permissions.all.permissions : p.key]
}
```

The output is something like:

```
[
  "repository:list",
  "repository:pull",
  "repository:push",
  "repository:delete",
  "artifact:read",
  "artifact:list",
  "artifact:delete",
  "tag:create",
  "tag:delete",
  "tag:list",
  "scan:create",
  "scan:stop",
]
```

### Use it to build a CI robot with all repository actions plus pull-only artifact

```hcl
data "vnpaycloud_registry_permissions" "all" {}

resource "vnpaycloud_registry_robot_account" "ci" {
  name            = "ci-builder"
  expires_in_days = 365

  permission {
    registry_id = vnpaycloud_registry_project.app.id
    actions = concat(
      [for p in data.vnpaycloud_registry_permissions.all.permissions :
         p.key if p.resource == "repository"],
      ["artifact:read"],
    )
  }
}
```

## Schema

### Read-Only

- `permissions` (List of Object) The catalogue of permissions.
  - `resource` (String) Resource the permission applies to (`repository`, `artifact`, `tag`, `scan`, …).
  - `action` (String) Action allowed on that resource (`push`, `pull`, `list`, …).
  - `key` (String) Convenience field `"<resource>:<action>"` — paste it directly into `permission.actions` of a robot account.
