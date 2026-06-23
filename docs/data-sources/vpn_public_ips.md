---
page_title: "vnpaycloud_vpn_public_ips Data Source - VNPayCloud"
subcategory: "Networking"
description: |-
  List all VPN public IPs in VNPayCloud.
---

# vnpaycloud_vpn_public_ips (Data Source)

Use this data source to list all VPN public IPs in the current project.

## Example Usage

```hcl
data "vnpaycloud_vpn_public_ips" "all" {}

output "active_vpn_public_ip_addresses" {
  value = [
    for ip in data.vnpaycloud_vpn_public_ips.all.vpn_public_ips :
    ip.address if ip.status == "active"
  ]
}
```

## Schema

### Read-Only

- `vpn_public_ips` (List of Object) List of VPN public IPs. Each element contains:
  - `id` (String) The unique identifier of the VPN public IP.
  - `name` (String) The name of the VPN public IP.
  - `description` (String) A human-readable description of the VPN public IP.
  - `address` (String) The allocated public IP address.
  - `status` (String) The current status of the VPN public IP.
  - `created_at` (String) The timestamp when the VPN public IP was created, in ISO 8601 format.
  - `zone_id` (String) The availability zone ID.
