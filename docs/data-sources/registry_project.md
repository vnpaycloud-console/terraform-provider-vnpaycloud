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

- `name` (String) The name of the registry project, used as the namespace for image repositories (e.g., `my-org/my-app`).
- `is_public` (Boolean) Whether the registry project is publicly accessible without authentication.
- `storage_limit` (Number) The maximum storage limit for this project in bytes. `-1` indicates unlimited.
- `storage_used` (Number) The current storage used by all repositories in this project, in bytes.
- `repo_count` (Number) The number of image repositories within this project.
- `status` (String) The current status of the registry project (e.g., `ACTIVE`, `DISABLED`).
- `created_at` (String) The timestamp when the registry project was created, in ISO 8601 format.
