---
page_title: "vnpaycloud_security_group_rule Resource - VNPayCloud"
subcategory: "Networking"
description: |-
  Manages a security group rule within VNPayCloud.
---

# vnpaycloud_security_group_rule (Resource)

Manages a security group rule within VNPayCloud. Rules define the allowed inbound or outbound traffic for a security group.

~> **Note:** All fields are immutable. Changing any field will force creation of a new rule.

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

### Allow Traffic from Another Security Group

```hcl
resource "vnpaycloud_security_group_rule" "allow_from_app" {
  security_group_id = vnpaycloud_security_group.db.id
  direction         = "ingress"
  protocol          = "tcp"
  port_range_min    = 5432
  port_range_max    = 5432
  remote_group_id   = vnpaycloud_security_group.app.id
}
```

## Schema

### Required

- `security_group_id` (String, ForceNew) The ID of the security group to which this rule belongs. Changing this creates a new rule.
- `direction` (String, ForceNew) The direction of the rule. Valid values are `ingress` or `egress`. Changing this creates a new rule.

### Optional

- `protocol` (String, ForceNew) The IP protocol of the rule. Common values are `tcp`, `udp`, `icmp`. If omitted, the rule applies to all protocols. Changing this creates a new rule.
- `ethertype` (String, ForceNew) The Ethernet type. Valid values are `IPv4` or `IPv6`. Defaults to `IPv4`. Changing this creates a new rule.
- `port_range_min` (Number, ForceNew) The minimum port number in the port range. Required when `protocol` is `tcp` or `udp`. Changing this creates a new rule.
- `port_range_max` (Number, ForceNew) The maximum port number in the port range. Required when `protocol` is `tcp` or `udp`. Changing this creates a new rule.
- `remote_ip_prefix` (String, ForceNew) The remote CIDR block the rule applies to. Conflicts with `remote_group_id`. Changing this creates a new rule.
- `remote_group_id` (String, ForceNew) The ID of the remote security group the rule applies to. Conflicts with `remote_ip_prefix`. Changing this creates a new rule.

### Read-Only

- `id` (String) The ID of the security group rule.

## Import

Security group rules can be imported using the `id`:

```shell
terraform import vnpaycloud_security_group_rule.example <rule-id>
```
