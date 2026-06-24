---
page_title: "vnpaycloud_network_acl Resource - VNPayCloud"
subcategory: "Networking"
description: |-
  Manages a Network ACL and its subnet mappings within VNPayCloud.
---

# vnpaycloud_network_acl (Resource)

Manages a Network ACL in a VPC. The `subnet_ids` set maps or unmaps the ACL to subnets.

~> **Note:** A newly created Network ACL includes default rules at priorities `1` and `100`. Use other priorities for custom rules.

## Example Usage

```hcl
resource "vnpaycloud_network_acl" "app" {
  name        = "app-acl"
  vpc_id      = vnpaycloud_vpc.main.id
  description = "Application ACL"
  subnet_ids = [
    vnpaycloud_subnet.app.id,
  ]
}
```

## Schema

### Required

- `name` (String, ForceNew) The ACL name.
- `vpc_id` (String, ForceNew) The VPC ID.

### Optional

- `description` (String, ForceNew) The ACL description.
- `subnet_ids` (Set of String) Subnet IDs mapped to the ACL. Updating this set maps and unmaps the backing networks for those subnets.

### Read-Only

- `id` (String) The ACL ID.
- `total_rules` (Number) The number of rules in the ACL.
- `status` (String) The current ACL status.
- `created_at` (String) The creation timestamp.

## Import

Network ACLs can be imported using the `id`:

```shell
terraform import vnpaycloud_network_acl.example <network-acl-id>
```
