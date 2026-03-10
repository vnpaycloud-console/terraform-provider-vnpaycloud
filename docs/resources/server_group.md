---
page_title: "vnpaycloud_server_group Resource - VNPayCloud"
subcategory: "Compute"
description: |-
  Manages a server group within VNPayCloud.
---

# vnpaycloud_server_group (Resource)

Manages a server group within VNPayCloud. Server groups define scheduling policies that control how instances are placed on physical hosts. For example, an `anti-affinity` policy ensures instances are placed on different hosts for high availability.

## Example Usage

### Creating an anti-affinity server group

```hcl
resource "vnpaycloud_server_group" "ha_group" {
  name   = "ha-web-servers"
  policy = "anti-affinity"
}
```

### Using a server group with an instance

```hcl
resource "vnpaycloud_server_group" "app_group" {
  name   = "app-servers"
  policy = "anti-affinity"
}

resource "vnpaycloud_instance" "web" {
  name           = "web-01"
  flavor_id      = "flavor-abc123"
  image_id       = "image-abc123"
  server_group_id = vnpaycloud_server_group.app_group.id
}
```

## Schema

### Required

- `name` (String, ForceNew) The name of the server group. Changing this creates a new server group.
- `policy` (String, ForceNew) The scheduling policy of the server group (e.g., `anti-affinity`, `affinity`). Changing this creates a new server group.

### Read-Only

- `id` (String) The ID of the server group.
- `member_ids` (List of String) The list of instance IDs that are members of this server group.
- `created_at` (String) The creation timestamp of the server group.

## Timeouts

- `create` - (Default `10 minutes`) Used for creating the server group.
- `delete` - (Default `10 minutes`) Used for deleting the server group.

## Import

Server groups can be imported using the `id`:

```shell
terraform import vnpaycloud_server_group.example <server-group-id>
```
