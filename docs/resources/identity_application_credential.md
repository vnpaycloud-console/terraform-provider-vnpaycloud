---
# generated by https://github.com/hashicorp/terraform-plugin-docs
page_title: "vnpaycloud_identity_application_credential Resource - terraform-provider-vnpaycloud"
subcategory: ""
description: |-
  
---

# vnpaycloud_identity_application_credential (Resource)





<!-- schema generated by tfplugindocs -->
## Schema

### Required

- `name` (String)
- `project_id` (String)

### Optional

- `access_rules` (Block List) (see [below for nested schema](#nestedblock--access_rules))
- `description` (String)
- `expires_at` (String)
- `region` (String)
- `roles` (Set of String)
- `secret` (String, Sensitive)
- `unrestricted` (Boolean)

### Read-Only

- `id` (String) The ID of this resource.

<a id="nestedblock--access_rules"></a>
### Nested Schema for `access_rules`

Required:

- `service` (String)

Optional:

- `method` (String)
- `path` (String)

Read-Only:

- `id` (String)
