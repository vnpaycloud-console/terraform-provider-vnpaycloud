---
page_title: "vnpaycloud_registry_project Data Source - VNPayCloud"
subcategory: "Container Registry"
description: |-
  Get information about a container registry project in VNPayCloud.
---

# vnpaycloud_registry_project (Data Source)

Use this data source to get information about an existing container registry project, including its storage usage and repository count. A registry project is a namespace that groups related container image repositories.

## Example Usage

```hcl
data "vnpaycloud_registry_project" "example" {
  id = "rp-klm23456"
}

output "project_name" {
  value = data.vnpaycloud_registry_project.example.name
}

output "repo_count" {
  value = data.vnpaycloud_registry_project.example.repo_count
}

output "storage_used_bytes" {
  value = data.vnpaycloud_registry_project.example.storage_used
}
```

## Schema

### Required (filter)

- `id` (String) The ID of the registry project.

### Read-Only

- `name` (String) User-friendly project name (the label shown in the console UI).
- `is_public` (Boolean) Whether the project is publicly accessible (anonymous `docker pull`).
- `storage_limit` (String) Maximum storage quota in bytes (as a string because the value exceeds 32-bit int).
- `storage_used` (Number) Current storage used by all repositories, in bytes.
- `repo_count` (Number) Number of image repositories inside the project.
- `status` (String) Project status: `active`, `creating`, `deleting`, `disabled`, `error`, `deleted`, `unknown`.
- `created_at` (String) Creation timestamp (RFC 3339 nanosecond).
- `namespace` (String) Full registry namespace `"{org_id_short}-{name}"`. **Use this** (not `name`) when tagging images: `vcr.vnpaycloud.vn/<namespace>/<repo>:<tag>`.
