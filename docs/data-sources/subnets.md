---
page_title: "vnpaycloud_subnets Data Source - VNPayCloud"
subcategory: "Networking"
description: |-
  List all subnets in VNPayCloud.
---

# vnpaycloud_subnets (Data Source)

Use this data source to list all subnets in the current project, optionally filtered by VPC.

## Example Usage

```hcl
data "vnpaycloud_subnets" "all" {}

output "all_subnet_cidrs" {
  value = data.vnpaycloud_subnets.all.subnets[*].cidr
}
```

```hcl
data "vnpaycloud_vpcs" "all" {}

data "vnpaycloud_subnets" "in_vpc" {
  vpc_id = data.vnpaycloud_vpcs.all.vpcs[0].id
}

output "subnet_ids_in_vpc" {
  value = data.vnpaycloud_subnets.in_vpc.subnets[*].id
}
```

## Schema

### Optional (filter)

- `vpc_id` (String) Filter subnets by the ID of the parent VPC.

### Read-Only

- `subnets` (List of Object) List of subnets. Each element contains:
  - `id` (String) The unique identifier of the subnet.
  - `name` (String) The name of the subnet.
  - `vpc_id` (String) The ID of the VPC this subnet belongs to.
  - `cidr` (String) The CIDR block of the subnet (e.g., `10.0.1.0/24`).
  - `gateway_ip` (String) The gateway IP address of the subnet.
  - `enable_dhcp` (Boolean) Whether DHCP is enabled on this subnet.
  - `status` (String) The current status of the subnet (e.g., `ACTIVE`, `BUILD`, `ERROR`).
  - `created_at` (String) The timestamp when the subnet was created, in ISO 8601 format.
