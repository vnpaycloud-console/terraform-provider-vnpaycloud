---
page_title: "vnpaycloud_vpn_connection Data Source - VNPayCloud"
subcategory: "Networking"
description: |-
  Get information about a VPN connection in VNPayCloud.
---

# vnpaycloud_vpn_connection (Data Source)

Use this data source to get information about an existing VPN connection by ID or name.

At least one of `id` or `name` must be specified. If both are specified, the VPN connection ID must match the given name. Looking up by name requires a unique match.

## Mode-Specific Fields

- Policy-based VPN connections use `vpn_type = "POLICY_BASED"`. `route_base_config` and `connection_bgp_config` are empty.
- Route-based static VPN connections use `vpn_type = "ROUTE_BASED"` with a customer gateway whose `routing_mode` is `STATIC`. `route_base_config` is populated; `connection_bgp_config` is empty.
- Route-based BGP VPN connections use `vpn_type = "ROUTE_BASED"` with a customer gateway whose `routing_mode` is `DYNAMIC`. Both `route_base_config` and `connection_bgp_config` are populated.
- `ipsec_auth_config` is not exposed by this data source because the pre-shared key is sensitive.

## Example Usage

```hcl
data "vnpaycloud_vpn_connection" "example" {
  name = "tf-vpn-conn-static"
}

output "vpn_connection_status" {
  value = data.vnpaycloud_vpn_connection.example.status
}
```

```hcl
data "vnpaycloud_vpn_connection" "by_id" {
  id = "vpn-connection-abc12345"
}
```

## Schema

### Optional (filter)

- `id` (String) The ID of the VPN connection.
- `name` (String) The name of the VPN connection.

### Read-Only

- `description` (String) A human-readable description of the VPN connection.
- `vpn_gateway_id` (String) The ID of the VPN gateway used by this VPN connection.
- `customer_gateway_id` (String) The ID of the customer gateway used by this VPN connection.
- `vpn_type` (String) The VPN type (`POLICY_BASED`, `ROUTE_BASED`).
- `vpn_public_ip_id` (String) The ID of the VPN public IP associated with this VPN connection.
- `ike_profile_config` (List of Object) The IKE profile configuration.
- `ipsec_profile_config` (List of Object) The IPSec profile configuration.
- `route_base_config` (List of Object) The route-based VPN configuration.
- `connection_bgp_config` (List of Object) The BGP timer configuration.
- `status` (String) The current status of the VPN connection.
- `created_at` (String) The timestamp when the VPN connection was created, in ISO 8601 format.
- `zone_id` (String) The availability zone ID.
