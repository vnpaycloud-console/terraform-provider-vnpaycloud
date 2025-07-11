---
# generated by https://github.com/hashicorp/terraform-plugin-docs
page_title: "vnpaycloud_compute_server Resource - terraform-provider-vnpaycloud"
subcategory: ""
description: |-
  
---

# vnpaycloud_compute_server (Resource)





<!-- schema generated by tfplugindocs -->
## Schema

### Required

- `name` (String)

### Optional

- `access_ip_v4` (String)
- `access_ip_v6` (String)
- `admin_pass` (String, Sensitive)
- `availability_zone` (String)
- `availability_zone_hints` (String)
- `block_device` (Block List) (see [below for nested schema](#nestedblock--block_device))
- `config_drive` (Boolean)
- `flavor_id` (String)
- `flavor_name` (String)
- `force_delete` (Boolean)
- `hypervisor_hostname` (String)
- `image_id` (String)
- `image_name` (String)
- `key_pair` (String)
- `metadata` (Map of String)
- `network` (Block List) (see [below for nested schema](#nestedblock--network))
- `network_mode` (String)
- `personality` (Block Set) (see [below for nested schema](#nestedblock--personality))
- `power_state` (String)
- `region` (String)
- `scheduler_hints` (Block Set) (see [below for nested schema](#nestedblock--scheduler_hints))
- `security_groups` (Set of String)
- `stop_before_destroy` (Boolean)
- `tags` (Set of String)
- `timeouts` (Block, Optional) (see [below for nested schema](#nestedblock--timeouts))
- `user_data` (String)
- `vendor_options` (Block Set, Max: 1) (see [below for nested schema](#nestedblock--vendor_options))

### Read-Only

- `all_metadata` (Map of String)
- `all_tags` (Set of String)
- `created` (String)
- `id` (String) The ID of this resource.
- `updated` (String)

<a id="nestedblock--block_device"></a>
### Nested Schema for `block_device`

Required:

- `source_type` (String)

Optional:

- `boot_index` (Number)
- `destination_type` (String)
- `device_type` (String)
- `disk_bus` (String)
- `guest_format` (String)
- `multiattach` (Boolean)
- `uuid` (String)
- `volume_size` (Number)
- `volume_type` (String)


<a id="nestedblock--network"></a>
### Nested Schema for `network`

Optional:

- `access_network` (Boolean)
- `fixed_ip_v4` (String)
- `fixed_ip_v6` (String)
- `name` (String)
- `port` (String)
- `uuid` (String)

Read-Only:

- `mac` (String)


<a id="nestedblock--personality"></a>
### Nested Schema for `personality`

Required:

- `content` (String)
- `file` (String)


<a id="nestedblock--scheduler_hints"></a>
### Nested Schema for `scheduler_hints`

Optional:

- `additional_properties` (Map of String)
- `build_near_host_ip` (String)
- `different_cell` (List of String)
- `different_host` (List of String)
- `group` (String)
- `query` (List of String)
- `same_host` (List of String)
- `target_cell` (String)


<a id="nestedblock--timeouts"></a>
### Nested Schema for `timeouts`

Optional:

- `create` (String)
- `delete` (String)
- `update` (String)


<a id="nestedblock--vendor_options"></a>
### Nested Schema for `vendor_options`

Optional:

- `detach_ports_before_destroy` (Boolean)
- `ignore_resize_confirmation` (Boolean)
