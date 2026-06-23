---
page_title: "vnpaycloud_vpn_gateways Data Source - VNPayCloud"
subcategory: "Networking"
description: |-
  List all VPN gateways in VNPayCloud.
---

# vnpaycloud_vpn_gateways (Data Source)

Use this data source to list all VPN gateways in the current project.

## Example Usage

```hcl
data "vnpaycloud_vpn_gateways" "all" {}

output "route_based_vpn_gateway_ids" {
  value = [
    for gw in data.vnpaycloud_vpn_gateways.all.vpn_gateways :
    gw.id if gw.vpn_type == "ROUTE_BASED"
  ]
}
```

## Schema

### Read-Only

- `vpn_gateways` (List of Object) List of VPN gateways. Each element contains:
  - `id` (String) The unique identifier of the VPN gateway.
  - `name` (String) The name of the VPN gateway.
  - `description` (String) A human-readable description of the VPN gateway.
  - `vpn_type` (String) The VPN gateway type (`POLICY_BASED`, `ROUTE_BASED`).
  - `status` (String) The current status of the VPN gateway.
  - `attached_vpc_ids` (List of String) The IDs of VPCs currently attached to the VPN gateway.
  - `created_at` (String) The timestamp when the VPN gateway was created, in ISO 8601 format.
  - `zone_id` (String) The availability zone ID.
