---
# generated by https://github.com/hashicorp/terraform-plugin-docs
page_title: "vnpaycloud_networking_route Resource - terraform-provider-vnpaycloud"
subcategory: ""
description: |-
  
---

# vnpaycloud_networking_route (Resource)





<!-- schema generated by tfplugindocs -->
## Schema

### Required

- `cidr_block` (String)
- `vpc_id` (String)

### Optional

- `internet_gateway_id` (String)
- `peering_connection_id` (String)
- `timeouts` (Block, Optional) (see [below for nested schema](#nestedblock--timeouts))

### Read-Only

- `id` (String) The ID of this resource.
- `name` (String)
- `target_id` (String)
- `target_type` (String)

<a id="nestedblock--timeouts"></a>
### Nested Schema for `timeouts`

Optional:

- `create` (String)
- `delete` (String)
