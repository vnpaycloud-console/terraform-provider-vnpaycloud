---
page_title: "vnpaycloud_server_group Data Source - VNPayCloud"
subcategory: "Compute"
description: |-
  Get information about a server group in VNPayCloud.
---

# vnpaycloud_server_group (Data Source)

Use this data source to get information about an existing server group, including its scheduling policy and member instances.

## Example Usage

```hcl
data "vnpaycloud_server_group" "my_group" {
  name = "web-servers"
}

output "server_group_policy" {
  value = data.vnpaycloud_server_group.my_group.policy
}

output "server_group_members" {
  value = data.vnpaycloud_server_group.my_group.member_ids
}
```

```hcl
data "vnpaycloud_server_group" "by_id" {
  id = "sg-abc123"
}
```

## Schema

### Optional (filter)

- `id` (String) The ID of the server group.
- `name` (String) The name of the server group.

### Read-Only

- `policy` (String) The scheduling policy of the server group (e.g., `anti-affinity`, `affinity`).
- `member_ids` (List of String) The list of instance IDs that are members of this server group.
- `created_at` (String) The creation timestamp of the server group.
