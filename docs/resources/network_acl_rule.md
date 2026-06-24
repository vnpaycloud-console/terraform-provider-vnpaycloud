---
page_title: "vnpaycloud_network_acl_rule Resource - VNPayCloud"
subcategory: "Networking"
description: |-
  Manages a Network ACL rule within VNPayCloud.
---

# vnpaycloud_network_acl_rule (Resource)

Manages a rule in a Network ACL. The `type` field selects the protocol and preset ports.

Each Network ACL is created with two default rules at priority `1` (drop all) and priority `100` (allow all); choose a `priority` other than these for your own rules. `source` and `destination` must be set to valid CIDRs.

## Example Usage

```hcl
resource "vnpaycloud_network_acl_rule" "allow_https" {
  nacl_id     = vnpaycloud_network_acl.app.id
  name        = "allow-https"
  priority    = 200
  type        = "HTTPS"
  action      = "allow"
  source      = "0.0.0.0/0"
  destination = "10.0.1.0/24"
}

resource "vnpaycloud_network_acl_rule" "allow_custom_tcp" {
  nacl_id     = vnpaycloud_network_acl.app.id
  name        = "allow-app"
  priority    = 210
  type        = "CUSTOM_TCP"
  action      = "allow"
  port_start  = 8080
  port_end    = 8080
  source      = "10.0.0.0/16"
  destination = "10.0.1.0/24"
}

resource "vnpaycloud_network_acl_rule" "allow_icmp" {
  nacl_id     = vnpaycloud_network_acl.app.id
  name        = "allow-ping"
  priority    = 220
  type        = "ICMP"
  action      = "allow"
  icmp_type   = "Echo"
  source      = "0.0.0.0/0"
  destination = "10.0.1.0/24"
}
```

## Schema

### Required

- `nacl_id` (String, ForceNew) The Network ACL ID.
- `name` (String, ForceNew) The rule name.
- `priority` (Number, ForceNew) The priority from 1 to 1000.
- `type` (String, ForceNew) One of `ALL_TRAFFIC`, `CUSTOM_TCP`, `CUSTOM_UDP`, `ICMP`, `SSH`, `TELNET`, `SMTP`, `DNS_TCP`, `DNS_UDP`, `HTTP`, `HTTPS`.
- `action` (String, ForceNew) Either `allow` or `drop`.
- `source` (String, ForceNew) Source CIDR (e.g. `0.0.0.0/0`).
- `destination` (String, ForceNew) Destination CIDR (e.g. `10.0.1.0/24`).

### Optional

- `port_start` (Number, ForceNew) Required for `CUSTOM_TCP` and `CUSTOM_UDP`; computed for preset types.
- `port_end` (Number, ForceNew) Required for `CUSTOM_TCP` and `CUSTOM_UDP`; computed for preset types.
- `icmp_type` (String, ForceNew) ICMP subtype, only valid when `type` is `ICMP`. Examples: `Echo`, `Echo_Reply`, `Destination_Unreachable`, `Time_Exceeded`, `Redirect`.
- `description` (String, ForceNew) Rule description.

### Read-Only

- `id` (String) The rule ID.
- `status` (String) The current rule status.

## Import

Network ACL rules can be imported using the `id`:

```shell
terraform import vnpaycloud_network_acl_rule.example <network-acl-rule-id>
```
