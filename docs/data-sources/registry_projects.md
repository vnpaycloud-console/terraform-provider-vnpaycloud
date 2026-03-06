---
page_title: "vnpaycloud_registry_projects Data Source - VNPayCloud"
subcategory: "Container Registry"
description: |-
  List all container registry projects in VNPayCloud.
---

# vnpaycloud_registry_projects (Data Source)

Use this data source to list all container registry projects in the current project.

## Example Usage

```hcl
data "vnpaycloud_registry_projects" "all" {}

output "all_registry_names" {
  value = data.vnpaycloud_registry_projects.all.registries[*].name
}

output "public_registry_names" {
  value = [
    for reg in data.vnpaycloud_registry_projects.all.registries :
    reg.name if reg.is_public == true
  ]
}

output "registry_usage_summary" {
  value = {
    for reg in data.vnpaycloud_registry_projects.all.registries :
    reg.name => {
      used_gb    = reg.storage_used
      limit_gb   = reg.storage_limit
      repo_count = reg.repo_count
    }
  }
}
```

## Schema

### Read-Only

- `registries` (List of Object) List of container registry projects. Each element contains:
  - `id` (String) The unique identifier of the registry project.
  - `name` (String) The name of the registry project.
  - `is_public` (Boolean) Whether the registry project is publicly accessible.
  - `storage_limit` (Number) The maximum storage limit for the registry in GB. `0` means unlimited.
  - `storage_used` (Number) The amount of storage currently used by the registry in GB.
  - `repo_count` (Number) The number of repositories within this registry project.
  - `status` (String) The current status of the registry project (e.g., `ACTIVE`, `ERROR`).
  - `created_at` (String) The timestamp when the registry project was created, in ISO 8601 format.
