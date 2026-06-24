---
page_title: "vnpaycloud_customer_gateway Resource - VNPayCloud"
subcategory: "Networking"
description: |-
  Manages a customer gateway within VNPayCloud.
---

# vnpaycloud_customer_gateway (Resource)

Manages a customer gateway within VNPayCloud. A customer gateway represents the customer-side VPN device and the remote networks behind it.

## Mode-Specific Requirements

The Terraform schema marks some fields as optional because the valid set depends on `vpn_type` and `routing_mode`. The backend validates the following combinations:

| Customer gateway mode | Required fields | Must not be set |
|---|---|---|
| Policy-based | `vpn_type = "POLICY_BASED"`, `remote_prefixes`, `public_ip`, `routing_mode = "NONE"` or omitted | `local_tunnel_ip`, `remote_tunnel_ip`, `bgp_config` |
| Route-based static | `vpn_type = "ROUTE_BASED"`, `routing_mode = "STATIC"`, `remote_prefixes`, `public_ip`, `local_tunnel_ip`, `remote_tunnel_ip` | `bgp_config` |
| Route-based BGP | `vpn_type = "ROUTE_BASED"`, `routing_mode = "DYNAMIC"`, `remote_prefixes`, `public_ip`, `local_tunnel_ip`, `remote_tunnel_ip`, `bgp_config` | none |

Additional validation:

- `public_ip` must be a public IPv4 address.
- `local_tunnel_ip` and `remote_tunnel_ip` are required for all route-based customer gateways, must use valid tunnel CIDR format such as `169.254.0.1/30`, and must be different.
- `remote_prefixes` must contain at least one valid CIDR. Duplicate prefixes are rejected.
- For route-based customer gateways, remote prefixes in the same request must not overlap.
- For policy-based customer gateways, overlapping remote prefixes are allowed.
- For route-based BGP, `bgp_config.as_path` is required and each ASN must be valid for the configured VPNaaS ASN range. The default range is `64512`–`65534`, `local_as` and `peer_as` must differ, and `as_path` accepts at most 10 space-separated ASNs.
- While the customer gateway is in use by a VPN connection, only `name`, `description`, and `remote_prefixes` may be updated. Changing `public_ip`, `local_tunnel_ip`, `remote_tunnel_ip`, `routing_mode`, or any `bgp_config` field is rejected until the connection is deleted. `vpn_type` is immutable in all cases.

## Example Usage

### Policy-Based Customer Gateway

```hcl
resource "vnpaycloud_customer_gateway" "policy" {
  name        = "tf-cgw-policy"
  description = "Policy-based customer gateway"
  public_ip       = "203.0.113.10"
  vpn_type        = "POLICY_BASED"
  remote_prefixes = ["10.10.0.0/16"]
}
```

### Route-Based Static Customer Gateway

```hcl
resource "vnpaycloud_customer_gateway" "static" {
  name        = "tf-cgw-static"
  description = "Route-based static customer gateway"
  public_ip        = "203.0.113.11"
  vpn_type         = "ROUTE_BASED"
  routing_mode     = "STATIC"
  remote_prefixes  = ["10.20.0.0/16"]
  local_tunnel_ip  = "169.254.0.1/30"
  remote_tunnel_ip = "169.254.0.2/30"
}
```

### Route-Based BGP Customer Gateway

```hcl
resource "vnpaycloud_customer_gateway" "bgp" {
  name        = "tf-cgw-bgp"
  description = "Route-based BGP customer gateway"
  public_ip        = "203.0.113.12"
  vpn_type         = "ROUTE_BASED"
  routing_mode     = "DYNAMIC"
  remote_prefixes  = ["10.30.0.0/16"]
  local_tunnel_ip  = "169.254.1.1/30"
  remote_tunnel_ip = "169.254.1.2/30"

  bgp_config {
    local_as = 65534
    peer_as  = 65000
    as_path  = "65000"
  }
}
```

## Schema

### Required

- `name` (String) The name of the customer gateway. Length must be between `3` and `255`, and it may contain only letters, digits, `.`, `_`, `-` and spaces.
- `public_ip` (String) The public IPv4 address of the customer-side VPN device.
- `vpn_type` (String, ForceNew) The VPN type. Valid values are `POLICY_BASED` and `ROUTE_BASED`.
- `remote_prefixes` (Set of String) The remote network CIDR prefixes behind the customer gateway. At least one prefix is required.

### Optional

- `description` (String) A human-readable description of the customer gateway.
- `remote_tunnel_ip` (String) The tunnel IP address on the customer gateway side, in tunnel CIDR format such as `169.254.0.2/30`. Required when `vpn_type = "ROUTE_BASED"` and not allowed when `vpn_type = "POLICY_BASED"`.
- `local_tunnel_ip` (String) The tunnel IP address on the VNPayCloud side, in tunnel CIDR format such as `169.254.0.1/30`. Required when `vpn_type = "ROUTE_BASED"` and not allowed when `vpn_type = "POLICY_BASED"`.
- `routing_mode` (String) The routing mode. Valid values are `NONE`, `STATIC`, and `DYNAMIC`. Defaults to `NONE`. Use `NONE` for policy-based VPN, `STATIC` for route-based static VPN, and `DYNAMIC` for route-based BGP VPN.
- `bgp_config` (Block List, Max: 1) The BGP configuration. Required when `vpn_type = "ROUTE_BASED"` and `routing_mode = "DYNAMIC"`. Not allowed for policy-based or route-based static customer gateways.

### Nested Schema for `bgp_config`

#### Required

- `local_as` (Number) The local BGP autonomous system number.
- `peer_as` (Number) The peer BGP autonomous system number.
- `as_path` (String) The BGP AS path to advertise for this customer gateway.

### Read-Only

- `id` (String) The ID of the customer gateway.
- `status` (String) The current status of the customer gateway.
- `created_at` (String) The creation timestamp of the customer gateway.

## Timeouts

- `create` - (Default `10 minutes`) Used for creating the customer gateway.
- `update` - (Default `10 minutes`) Used for updating the customer gateway.
- `delete` - (Default `10 minutes`) Used for deleting the customer gateway.

## Import

Customer gateways can be imported using the `id`:

```shell
terraform import vnpaycloud_customer_gateway.example <customer-gateway-id>
```
