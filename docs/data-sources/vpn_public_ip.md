---
page_title: "vnpaycloud_vpn_public_ip Data Source - VNPayCloud"
subcategory: "Networking"
description: |-
  Get information about a VPN public IP in VNPayCloud.
---

# vnpaycloud_vpn_public_ip (Data Source)

Use this data source to get information about an existing VPN public IP by ID or name.

At least one of `id` or `name` must be specified. If both are specified, the VPN public IP ID must match the given name. Looking up by name requires a unique match.

## Example Usage

```hcl
data "vnpaycloud_vpn_public_ip" "example" {
  name = "tf-vpn-public-ip"
}

output "vpn_public_ip_address" {
  value = data.vnpaycloud_vpn_public_ip.example.address
}
```

```hcl
data "vnpaycloud_vpn_public_ip" "by_id" {
  id = "vpn-public-ip-abc12345"
}
```

## Schema

### Optional (filter)

- `id` (String) The ID of the VPN public IP.
- `name` (String) The name of the VPN public IP.

### Read-Only

- `description` (String) A human-readable description of the VPN public IP.
- `address` (String) The allocated public IP address.
- `status` (String) The current status of the VPN public IP.
- `created_at` (String) The timestamp when the VPN public IP was created, in ISO 8601 format.
- `zone_id` (String) The availability zone ID.
