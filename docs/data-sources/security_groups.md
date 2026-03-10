---
page_title: "vnpaycloud_security_groups Data Source - VNPayCloud"
subcategory: "Networking"
description: |-
  List all security groups in VNPayCloud.
---

# vnpaycloud_security_groups (Data Source)

Use this data source to list all security groups in the current project.

## Example Usage

```hcl
data "vnpaycloud_security_groups" "all" {}

output "all_security_group_names" {
  value = data.vnpaycloud_security_groups.all.security_groups[*].name
}

output "active_security_group_ids" {
  value = [
    for sg in data.vnpaycloud_security_groups.all.security_groups :
    sg.id if sg.status == "ACTIVE"
  ]
}
```

## Schema

### Read-Only

- `security_groups` (List of Object) List of security groups. Each element contains:
  - `id` (String) The unique identifier of the security group.
  - `name` (String) The name of the security group.
  - `description` (String) A human-readable description of the security group.
  - `status` (String) The current status of the security group (e.g., `ACTIVE`, `BUILD`, `ERROR`).
  - `created_at` (String) The timestamp when the security group was created, in ISO 8601 format.
