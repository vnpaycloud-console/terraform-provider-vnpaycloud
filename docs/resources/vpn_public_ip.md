---
page_title: "vnpaycloud_vpn_public_ip Resource - VNPayCloud"
subcategory: "Networking"
description: |-
  Manages a VPN public IP within VNPayCloud.
---

# vnpaycloud_vpn_public_ip (Resource)

Manages a VPN public IP within VNPayCloud. A VPN public IP is allocated for use by VPN connections.

## Usage Notes

- The VPN public IP is allocated in the provider's configured zone and must be used with VPN resources from the same zone.
- The VPN public IP must be `active` before it can be used by a VPN connection.
- A VPN public IP can be used by only one blocking VPN connection.

## Example Usage

```hcl
resource "vnpaycloud_vpn_public_ip" "main" {
  name        = "tf-vpn-public-ip"
  description = "VPN public IP for site-to-site VPN"
}
```

## Schema

### Required

- `name` (String) The name of the VPN public IP. Length must be between `3` and `255`, and it may contain only letters, digits, `.`, `_`, `-` and spaces.

### Optional

- `description` (String) A human-readable description of the VPN public IP.

### Read-Only

- `id` (String) The ID of the VPN public IP.
- `address` (String) The allocated public IP address.
- `status` (String) The current status of the VPN public IP.
- `created_at` (String) The creation timestamp of the VPN public IP.

## Timeouts

- `create` - (Default `10 minutes`) Used for creating the VPN public IP.
- `update` - (Default `10 minutes`) Used for updating the VPN public IP.
- `delete` - (Default `10 minutes`) Used for deleting the VPN public IP.

## Import

VPN public IPs can be imported using the `id`:

```shell
terraform import vnpaycloud_vpn_public_ip.example <vpn-public-ip-id>
```
