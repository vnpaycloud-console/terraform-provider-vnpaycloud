---
page_title: "vnpaycloud_vpn_gateway Resource - VNPayCloud"
subcategory: "Networking"
description: |-
  Manages a VPN gateway within VNPayCloud.
---

# vnpaycloud_vpn_gateway (Resource)

Manages a VPN gateway within VNPayCloud. A VPN gateway is the cloud-side endpoint used by VPN connections. VPN gateways can be created independently and attached to VPCs using `vnpaycloud_vpn_gateway_vpc_attachment`, enabling separate lifecycle management for the gateway and its VPC attachments.

~> **Note:** The `vpn_type` field is immutable. Changing it will force creation of a new VPN gateway.

## Usage Notes

- `vpn_type` must match the customer gateway and VPN connection that will use this gateway.
- At least one VPC must be attached with `vnpaycloud_vpn_gateway_vpc_attachment` before creating a VPN connection.
- Policy-based VPN gateways can attach to at most one VPC.
- Policy-based VPN gateways can have only one blocking VPN connection.
- Route-based static and route-based BGP connections cannot be mixed on the same VPN gateway while blocking connections exist.

## Example Usage

### Route-Based VPN Gateway

```hcl
resource "vnpaycloud_vpn_gateway" "route_based" {
  name        = "tf-vpngw-route"
  description = "Route-based VPN gateway"
  vpn_type    = "ROUTE_BASED"
}
```

### Policy-Based VPN Gateway

```hcl
resource "vnpaycloud_vpn_gateway" "policy_based" {
  name        = "tf-vpngw-policy"
  description = "Policy-based VPN gateway"
  vpn_type    = "POLICY_BASED"
}
```

## Schema

### Required

- `name` (String) The name of the VPN gateway. Length must be between `3` and `255`, and it may contain only letters, digits, `.`, `_`, `-` and spaces.
- `vpn_type` (String, ForceNew) The VPN gateway type. Valid values are `POLICY_BASED` and `ROUTE_BASED`. Changing this creates a new VPN gateway.

### Optional

- `description` (String) A human-readable description of the VPN gateway.

### Read-Only

- `id` (String) The ID of the VPN gateway.
- `status` (String) The current status of the VPN gateway.
- `attached_vpc_ids` (List of String) The IDs of VPCs currently attached to the VPN gateway.
- `created_at` (String) The creation timestamp of the VPN gateway.

## Timeouts

- `create` - (Default `40 minutes`) Used for creating the VPN gateway.
- `update` - (Default `10 minutes`) Used for updating the VPN gateway.
- `delete` - (Default `10 minutes`) Used for deleting the VPN gateway.

## Import

VPN gateways can be imported using the `id`:

```shell
terraform import vnpaycloud_vpn_gateway.example <vpn-gateway-id>
```
