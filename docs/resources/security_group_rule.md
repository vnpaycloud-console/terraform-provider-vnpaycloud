---
page_title: "vnpaycloud_security_group_rule Resource - VNPayCloud"
subcategory: "Networking"
description: |-
  Manages a security group rule within VNPayCloud.
---

# vnpaycloud_security_group_rule (Resource)

Manages a security group rule within VNPayCloud. Rules define the allowed inbound or outbound traffic for a security group.

~> **Note:** `remote_ip_prefix` (CIDR) and `description` can be updated in place. All other fields (`security_group_id`, `direction`, `protocol`, `ethertype`, `port_range_min`, `port_range_max`) are immutable — changing any of them forces creation of a new rule.

~> **Note:** A newly created security group already includes a default egress rule that allows all outbound traffic (all protocols, `0.0.0.0/0`). Do not add another egress rule with the same `direction`/`ethertype`/`protocol`/port range/`remote_ip_prefix`, as a duplicate rule is rejected by the backend. A rule is considered a duplicate only when this whole combination matches an existing rule.

## Example Usage

### Allow HTTP/HTTPS Ingress

```hcl
resource "vnpaycloud_security_group" "web" {
  name        = "web-sg"
  description = "Web server security group"
}

resource "vnpaycloud_security_group_rule" "allow_http" {
  security_group_id = vnpaycloud_security_group.web.id
  direction         = "ingress"
  protocol          = "tcp"
  ethertype         = "IPv4"
  port_range_min    = 80
  port_range_max    = 80
  remote_ip_prefix  = "0.0.0.0/0"
}

resource "vnpaycloud_security_group_rule" "allow_https" {
  security_group_id = vnpaycloud_security_group.web.id
  direction         = "ingress"
  protocol          = "tcp"
  ethertype         = "IPv4"
  port_range_min    = 443
  port_range_max    = 443
  remote_ip_prefix  = "0.0.0.0/0"
}
```

## Schema

### Required

- `security_group_id` (String, ForceNew) The ID of the security group to which this rule belongs. Changing this creates a new rule.
- `direction` (String, ForceNew) The direction of the rule. Valid values are `ingress` or `egress`. Changing this creates a new rule.

### Optional

- `protocol` (String, ForceNew) The IP protocol of the rule. Valid values are `tcp`, `udp`, or `icmp`. If omitted, the rule applies to all protocols. Changing this creates a new rule.
- `ethertype` (String, ForceNew) The Ethernet type. Valid values are `IPv4` or `IPv6`. Defaults to `IPv4`. Changing this creates a new rule.
- `port_range_min` (Number, ForceNew) The minimum port number in the port range. If omitted for `tcp`/`udp`, the rule applies to all ports. For `icmp`, this is the ICMP type. Changing this creates a new rule.
- `port_range_max` (Number, ForceNew) The maximum port number in the port range. If omitted for `tcp`/`udp`, the rule applies to all ports. For `icmp`, this is the ICMP code. Changing this creates a new rule.
- `remote_ip_prefix` (String) The remote CIDR block the rule applies to. Can be updated in place.
- `description` (String) A description of the rule. May contain letters, digits, spaces, hyphens (`-`), underscores (`_`), and periods (`.`). Can be updated in place.

### Read-Only

- `id` (String) The ID of the security group rule.

## Import

Security group rules can be imported using the `id`:

```shell
terraform import vnpaycloud_security_group_rule.example <rule-id>
```
