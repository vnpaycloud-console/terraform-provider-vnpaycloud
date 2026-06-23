---
page_title: "vnpaycloud_vpn_gateway Data Source - VNPayCloud"
subcategory: "Networking"
description: |-
  Get information about a VPN gateway in VNPayCloud.
---

# vnpaycloud_vpn_gateway (Data Source)

Use this data source to get information about an existing VPN gateway by ID or name.

At least one of `id` or `name` must be specified. If both are specified, the VPN gateway ID must match the given name. Looking up by name requires a unique match.

## Example Usage

```hcl
data "vnpaycloud_vpn_gateway" "example" {
  name = "tf-vpngw-route"
}

output "vpn_gateway_id" {
  value = data.vnpaycloud_vpn_gateway.example.id
}
```

```hcl
data "vnpaycloud_vpn_gateway" "by_id" {
  id = "vpn-gateway-abc12345"
}
```

## Schema

### Optional (filter)

- `id` (String) The ID of the VPN gateway.
- `name` (String) The name of the VPN gateway.

### Read-Only

- `description` (String) A human-readable description of the VPN gateway.
- `vpn_type` (String) The VPN gateway type (`POLICY_BASED`, `ROUTE_BASED`).
- `status` (String) The current status of the VPN gateway.
- `attached_vpc_ids` (List of String) The IDs of VPCs currently attached to the VPN gateway.
- `created_at` (String) The timestamp when the VPN gateway was created, in ISO 8601 format.
- `zone_id` (String) The availability zone ID.
