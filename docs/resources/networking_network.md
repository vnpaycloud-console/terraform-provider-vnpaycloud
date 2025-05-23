---
# generated by https://github.com/hashicorp/terraform-plugin-docs
page_title: "vnpaycloud_networking_network Resource - terraform-provider-vnpaycloud"
subcategory: ""
description: |-
  
---

# vnpaycloud_networking_network (Resource)





<!-- schema generated by tfplugindocs -->
## Schema

### Optional

- `admin_state_up` (Boolean)
- `availability_zone_hints` (Set of String)
- `description` (String)
- `dns_domain` (String)
- `external` (Boolean)
- `mtu` (Number)
- `name` (String)
- `port_security_enabled` (Boolean)
- `qos_policy_id` (String)
- `region` (String)
- `segments` (Block Set) (see [below for nested schema](#nestedblock--segments))
- `shared` (Boolean)
- `tags` (Set of String)
- `tenant_id` (String)
- `timeouts` (Block, Optional) (see [below for nested schema](#nestedblock--timeouts))
- `transparent_vlan` (Boolean)
- `value_specs` (Map of String)

### Read-Only

- `all_tags` (Set of String)
- `id` (String) The ID of this resource.

<a id="nestedblock--segments"></a>
### Nested Schema for `segments`

Optional:

- `network_type` (String)
- `physical_network` (String)
- `segmentation_id` (Number)


<a id="nestedblock--timeouts"></a>
### Nested Schema for `timeouts`

Optional:

- `create` (String)
- `delete` (String)
