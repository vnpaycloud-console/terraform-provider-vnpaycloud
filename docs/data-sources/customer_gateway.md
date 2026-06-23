---
page_title: "vnpaycloud_customer_gateway Data Source - VNPayCloud"
subcategory: "Networking"
description: |-
  Get information about a customer gateway in VNPayCloud.
---

# vnpaycloud_customer_gateway (Data Source)

Use this data source to get information about an existing customer gateway by ID or name.

At least one of `id` or `name` must be specified. If both are specified, the customer gateway ID must match the given name. Looking up by name requires a unique match.

## Mode-Specific Fields

- Policy-based customer gateways use `vpn_type = "POLICY_BASED"` and `routing_mode = "NONE"`. `local_tunnel_ip`, `remote_tunnel_ip`, and `bgp_config` are empty.
- Route-based static customer gateways use `vpn_type = "ROUTE_BASED"` and `routing_mode = "STATIC"`. `local_tunnel_ip` and `remote_tunnel_ip` are populated; `bgp_config` is empty.
- Route-based BGP customer gateways use `vpn_type = "ROUTE_BASED"` and `routing_mode = "DYNAMIC"`. `local_tunnel_ip`, `remote_tunnel_ip`, and `bgp_config` are populated.

## Example Usage

```hcl
data "vnpaycloud_customer_gateway" "example" {
  name = "tf-cgw-static"
}

output "customer_gateway_public_ip" {
  value = data.vnpaycloud_customer_gateway.example.public_ip
}
```

```hcl
data "vnpaycloud_customer_gateway" "by_id" {
  id = "customer-gateway-abc12345"
}
```

## Schema

### Optional (filter)

- `id` (String) The ID of the customer gateway.
- `name` (String) The name of the customer gateway.

### Read-Only

- `description` (String) A human-readable description of the customer gateway.
- `public_ip` (String) The public IPv4 address of the customer-side VPN device.
- `vpn_type` (String) The VPN type (`POLICY_BASED`, `ROUTE_BASED`).
- `status` (String) The current status of the customer gateway.
- `remote_prefixes` (Set of String) The remote network CIDR prefixes behind the customer gateway.
- `remote_tunnel_ip` (String) The tunnel IP address on the customer gateway side.
- `local_tunnel_ip` (String) The tunnel IP address on the VNPayCloud side.
- `routing_mode` (String) The routing mode.
- `bgp_config` (List of Object) The BGP configuration for route-based VPN with dynamic routing. The block contains:
  - `local_as` (Number) The local BGP autonomous system number.
  - `peer_as` (Number) The peer BGP autonomous system number.
  - `as_path` (String) The configured BGP AS path.
- `created_at` (String) The timestamp when the customer gateway was created, in ISO 8601 format.
- `zone_id` (String) The availability zone ID.
