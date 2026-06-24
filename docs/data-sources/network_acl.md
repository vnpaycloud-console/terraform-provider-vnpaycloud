---
page_title: "vnpaycloud_network_acl Data Source - VNPayCloud"
subcategory: "Networking"
description: |-
  Get information about a Network ACL in VNPayCloud.
---

# vnpaycloud_network_acl (Data Source)

Use this data source to get an existing Network ACL by ID, or by filtering with `name` and optionally `vpc_id`.

## Example Usage

```hcl
data "vnpaycloud_network_acl" "app" {
  name   = "app-acl"
  vpc_id = vnpaycloud_vpc.main.id
}
```

## Schema

### Optional

- `id` (String) The ACL ID. If set, the provider reads this ACL directly.
- `name` (String) The ACL name used for lookup when `id` is omitted.
- `vpc_id` (String) The VPC ID used to narrow lookup when `id` is omitted.

### Read-Only

- `description` (String) The ACL description.
- `subnet_ids` (Set of String) Subnet IDs mapped to the ACL.
- `total_rules` (Number) The number of rules in the ACL.
- `status` (String) The current ACL status.
- `created_at` (String) The creation timestamp.
