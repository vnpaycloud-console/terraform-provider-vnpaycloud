---
page_title: "vnpaycloud_vpc_peering Data Source - VNPayCloud"
subcategory: "Networking"
description: |-
  Get information about a VPC peering connection in VNPayCloud.
---

# vnpaycloud_vpc_peering (Data Source)

Use this data source to get information about an existing VPC peering connection. VPC peering enables private network connectivity between two VPCs without routing traffic through the public internet.

## Example Usage

```hcl
data "vnpaycloud_vpc_peering" "example" {
  name = "my-vpc-peering"
}

output "peering_status" {
  value = data.vnpaycloud_vpc_peering.example.peering_status
}

output "source_vpc_cidr" {
  value = data.vnpaycloud_vpc_peering.example.src_vpc_cidr
}
```

```hcl
data "vnpaycloud_vpc_peering" "by_id" {
  id = "pcx-wxy89012"
}
```

## Schema

### Optional (filter)

- `id` (String) The ID of the VPC peering connection.
- `name` (String) The name of the VPC peering connection.

### Read-Only

- `src_vpc_id` (String) The ID of the source (requester) VPC in the peering connection.
- `dest_vpc_id` (String) The ID of the destination (accepter) VPC in the peering connection.
- `description` (String) A human-readable description of the peering connection.
- `status` (String) The provisioning status of the peering connection (e.g., `ACTIVE`, `BUILD`, `ERROR`).
- `peering_status` (String) The negotiation status of the peering connection (e.g., `pending-acceptance`, `active`, `rejected`, `expired`, `deleted`).
- `src_vpc_cidr` (String) The CIDR block of the source VPC.
- `dest_vpc_cidr` (String) The CIDR block of the destination VPC.
- `created_at` (String) The timestamp when the VPC peering connection was created, in ISO 8601 format.
