---
page_title: "vnpaycloud_security_group Data Source - VNPayCloud"
subcategory: "Networking"
description: |-
  Get information about a security group in VNPayCloud.
---

# vnpaycloud_security_group (Data Source)

Use this data source to get information about an existing security group, including all of its inbound and outbound rules.

## Example Usage

```hcl
data "vnpaycloud_security_group" "example" {
  name = "allow-web-traffic"
}

output "sg_rules" {
  value = data.vnpaycloud_security_group.example.rules
}
```

```hcl
data "vnpaycloud_security_group" "by_id" {
  id = "sg-def67890"
}
```

## Schema

### Optional (filter)

- `id` (String) The ID of the security group.
- `name` (String) The name of the security group.

### Read-Only

- `description` (String) A human-readable description of the security group.
- `status` (String) The current status of the security group (e.g., `ACTIVE`).
- `rules` (List of Object) A list of security group rules. Each object contains:
  - `id` (String) The ID of the security group rule.
  - `security_group_id` (String) The ID of the security group this rule belongs to.
  - `direction` (String) The direction of traffic the rule applies to (`ingress` or `egress`).
  - `protocol` (String) The IP protocol (e.g., `tcp`, `udp`, `icmp`). Empty string means all protocols.
  - `ethertype` (String) The Ethernet frame type (`IPv4` or `IPv6`).
  - `port_range_min` (Number) The minimum port number in the range. `0` means all ports (for `icmp`, this is the type).
  - `port_range_max` (Number) The maximum port number in the range. `0` means all ports (for `icmp`, this is the code).
  - `remote_ip_prefix` (String) The remote CIDR IP prefix that traffic is allowed from/to.
  - `remote_group_id` (String) The ID of the remote security group that traffic is allowed from/to.
- `created_at` (String) The timestamp when the security group was created, in ISO 8601 format.
