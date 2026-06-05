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
      used_bytes  = reg.storage_used
      limit_bytes = reg.storage_limit
      repo_count  = reg.repo_count
    }
  }
}

# Build docker tag prefixes for every project
output "docker_tag_prefixes" {
  value = {
    for reg in data.vnpaycloud_registry_projects.all.registries :
    reg.name => "vcr.vnpaycloud.vn/${reg.namespace}"
  }
}
```

## Schema

### Read-Only

- `registries` (List of Object) List of container registry projects. Each element contains:
  - `id` (String) Project ID.
  - `name` (String) Project name (used as the user-facing label).
  - `is_public` (Boolean) Whether the project is publicly accessible (anonymous `docker pull`).
  - `storage_limit` (String) Maximum storage quota in bytes (string because the value exceeds 32-bit int).
  - `storage_used` (Number) Current storage usage in bytes.
  - `repo_count` (Number) Number of image repositories inside the project.
  - `status` (String) Project status: `active`, `creating`, `deleting`, `disabled`, `error`, `deleted`, `unknown`.
  - `created_at` (String) Creation timestamp (RFC 3339 nanosecond).
  - `namespace` (String) Full registry namespace `"{org_id_short}-{name}"`. Use it to tag images: `vcr.vnpaycloud.vn/<namespace>/<repo>:<tag>`.
