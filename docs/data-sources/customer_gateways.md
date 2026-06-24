---
page_title: "vnpaycloud_customer_gateways Data Source - VNPayCloud"
subcategory: "Networking"
description: |-
  List all customer gateways in VNPayCloud.
---

# vnpaycloud_customer_gateways (Data Source)

Use this data source to list all customer gateways in the current project.

## Mode-Specific Fields

- Policy-based customer gateways use `vpn_type = "POLICY_BASED"` and `routing_mode = "NONE"`. `local_tunnel_ip`, `remote_tunnel_ip`, and `bgp_config` are empty.
- Route-based static customer gateways use `vpn_type = "ROUTE_BASED"` and `routing_mode = "STATIC"`. `local_tunnel_ip` and `remote_tunnel_ip` are populated; `bgp_config` is empty.
- Route-based BGP customer gateways use `vpn_type = "ROUTE_BASED"` and `routing_mode = "DYNAMIC"`. `local_tunnel_ip`, `remote_tunnel_ip`, and `bgp_config` are populated.

## Example Usage

```hcl
data "vnpaycloud_customer_gateways" "all" {}

output "route_based_customer_gateway_ids" {
  value = [
    for cgw in data.vnpaycloud_customer_gateways.all.customer_gateways :
    cgw.id if cgw.vpn_type == "ROUTE_BASED"
  ]
}
```

## Schema

### Read-Only

- `customer_gateways` (List of Object) List of customer gateways. Each element contains:
  - `id` (String) The unique identifier of the customer gateway.
  - `name` (String) The name of the customer gateway.
  - `description` (String) A human-readable description of the customer gateway.
  - `public_ip` (String) The public IPv4 address of the customer-side VPN device.
  - `vpn_type` (String) The VPN type (`POLICY_BASED`, `ROUTE_BASED`).
  - `status` (String) The current status of the customer gateway.
  - `remote_prefixes` (Set of String) The remote network CIDR prefixes behind the customer gateway.
  - `remote_tunnel_ip` (String) The tunnel IP address on the customer gateway side.
  - `local_tunnel_ip` (String) The tunnel IP address on the VNPayCloud side.
  - `routing_mode` (String) The routing mode.
  - `bgp_config` (List of Object) The BGP configuration for route-based VPN with dynamic routing.
  - `created_at` (String) The timestamp when the customer gateway was created, in ISO 8601 format.
  - `zone_id` (String) The availability zone ID.
