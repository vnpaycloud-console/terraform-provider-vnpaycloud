---
page_title: "vnpaycloud_subnet Data Source - VNPayCloud"
subcategory: "Networking"
description: |-
  Get information about a subnet in VNPayCloud.
---

# vnpaycloud_subnet (Data Source)

Use this data source to get information about an existing subnet.

## Example Usage

```hcl
data "vnpaycloud_subnet" "example" {
  name = "my-subnet"
}

output "subnet_cidr" {
  value = data.vnpaycloud_subnet.example.cidr
}
```

```hcl
data "vnpaycloud_subnet" "by_vpc" {
  vpc_id = "vpc-abc12345"
}
```

## Schema

### Optional (filter)

- `id` (String) The ID of the subnet.
- `name` (String) The name of the subnet.
- `vpc_id` (String) The ID of the VPC to which the subnet belongs.

### Read-Only

- `cidr` (String) The CIDR block of the subnet (e.g., `10.0.1.0/24`).
- `gateway_ip` (String) The IP address of the default gateway for the subnet.
- `enable_dhcp` (Boolean) Whether DHCP is enabled for this subnet.
- `enable_snat` (Boolean) Whether source NAT is enabled for this subnet.
- `floating_ip_id` (String) The ID of the floating IP associated with this subnet's gateway, if any.
- `status` (String) The current status of the subnet (e.g., `ACTIVE`, `BUILD`, `ERROR`).
- `created_at` (String) The timestamp when the subnet was created, in ISO 8601 format.
