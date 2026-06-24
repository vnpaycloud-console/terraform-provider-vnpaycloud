---
page_title: "vnpaycloud_security_group Resource - VNPayCloud"
subcategory: "Networking"
description: |-
  Manages a security group resource within VNPayCloud.
---

# vnpaycloud_security_group (Resource)

Manages a security group resource within VNPayCloud. Security groups act as virtual firewalls that control inbound and outbound traffic for your instances.

## Example Usage

```hcl
resource "vnpaycloud_security_group" "example" {
  name        = "my-security-group"
  description = "Security group for web servers"
}

resource "vnpaycloud_security_group_rule" "allow_http" {
  security_group_id = vnpaycloud_security_group.example.id
  direction         = "ingress"
  protocol          = "tcp"
  port_range_min    = 80
  port_range_max    = 80
  remote_ip_prefix  = "0.0.0.0/0"
}
```

## Schema

### Required

- `name` (String) The name of the security group. Must be unique within the project.

### Optional

- `description` (String) A description of the security group. May contain only letters, digits, spaces, hyphens (`-`), underscores (`_`), and periods (`.`) (`^[a-zA-Z0-9-_. ]*$`).
- `enable_log` (Boolean) Whether ACCEPT network logging is enabled for this security group. When set to `true`, an ACCEPT log is created; setting it back to `false` removes the log. Can be set at create and updated in place. Network logging is only available in zones that support it — when `can_enable_log` is `false`, enabling or disabling `enable_log` is rejected at plan time, so only set this in a supporting zone.

### Read-Only

- `id` (String) The ID of the security group.
- `can_enable_log` (Boolean) Whether network logging can be enabled for this security group.
- `rules` (List of Object) The list of security group rules currently associated with this security group. Each object contains:
  - `id` (String) The ID of the rule.
  - `security_group_id` (String) The ID of the security group this rule belongs to.
  - `direction` (String) The direction of the rule (`ingress` or `egress`).
  - `protocol` (String) The IP protocol of the rule (e.g., `tcp`, `udp`, `icmp`).
  - `ethertype` (String) The Ethernet type (`IPv4` or `IPv6`).
  - `port_range_min` (Number) The minimum port number in the range.
  - `port_range_max` (Number) The maximum port number in the range.
  - `remote_ip_prefix` (String) The remote CIDR block the rule applies to.
- `created_at` (String) The creation timestamp of the security group.

## Import

Security groups can be imported using the `id`:

```shell
terraform import vnpaycloud_security_group.example <security-group-id>
```
