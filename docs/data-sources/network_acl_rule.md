---
page_title: "vnpaycloud_network_acl_rule Data Source - VNPayCloud"
subcategory: "Networking"
description: |-
  Get information about a Network ACL rule in VNPayCloud.
---

# vnpaycloud_network_acl_rule (Data Source)

Use this data source to get an existing Network ACL rule by ID, or by `name` within a specific ACL.

## Example Usage

```hcl
data "vnpaycloud_network_acl_rule" "https" {
  nacl_id = vnpaycloud_network_acl.app.id
  name    = "allow-https"
}
```

## Schema

### Optional

- `id` (String) The rule ID. If set, the provider reads this rule directly.
- `nacl_id` (String) The Network ACL ID. Required when looking up by `name`.
- `name` (String) The rule name used for lookup when `id` is omitted.

### Read-Only

- `priority` (Number) The rule priority.
- `type` (String) The rule type preset.
- `action` (String) Either `allow` or `drop`.
- `port_start` (Number) The effective starting port.
- `port_end` (Number) The effective ending port.
- `source` (String) Source CIDR.
- `destination` (String) Destination CIDR.
- `icmp_type` (String) ICMP subtype when applicable.
- `description` (String) Rule description.
- `status` (String) The current rule status.
