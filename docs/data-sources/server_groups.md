---
page_title: "vnpaycloud_server_groups Data Source - VNPayCloud"
subcategory: "Compute"
description: |-
  List all server groups in VNPayCloud.
---

# vnpaycloud_server_groups (Data Source)

Use this data source to list all server groups in the current project.

## Example Usage

```hcl
data "vnpaycloud_server_groups" "all" {}

output "all_server_group_names" {
  value = data.vnpaycloud_server_groups.all.server_groups[*].name
}

output "anti_affinity_groups" {
  value = [
    for sg in data.vnpaycloud_server_groups.all.server_groups :
    sg.name if sg.policy == "anti-affinity"
  ]
}
```

## Schema

### Read-Only

- `server_groups` (List of Object) List of server groups. Each element contains:
  - `id` (String) The unique identifier of the server group.
  - `name` (String) The name of the server group.
  - `policy` (String) The scheduling policy of the server group (e.g., `anti-affinity`, `affinity`).
  - `member_ids` (List of String) The list of instance IDs that are members of this server group.
  - `created_at` (String) The creation timestamp of the server group.
