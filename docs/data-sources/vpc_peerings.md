---
page_title: "vnpaycloud_vpc_peerings Data Source - VNPayCloud"
subcategory: "Networking"
description: |-
  List all VPC peering connections in VNPayCloud.
---

# vnpaycloud_vpc_peerings (Data Source)

Use this data source to list all VPC peering connections in the current project.

## Example Usage

```hcl
data "vnpaycloud_vpc_peerings" "all" {}

output "all_peering_names" {
  value = data.vnpaycloud_vpc_peerings.all.vpc_peerings[*].name
}

output "active_peerings" {
  value = [
    for p in data.vnpaycloud_vpc_peerings.all.vpc_peerings :
    {
      name          = p.name
      src_vpc_cidr  = p.src_vpc_cidr
      dest_vpc_cidr = p.dest_vpc_cidr
    }
    if p.peering_status == "ACTIVE"
  ]
}

output "peering_cidr_map" {
  value = {
    for p in data.vnpaycloud_vpc_peerings.all.vpc_peerings :
    p.name => "${p.src_vpc_cidr} <-> ${p.dest_vpc_cidr}"
  }
}
```

## Schema

### Read-Only

- `vpc_peerings` (List of Object) List of VPC peering connections. Each element contains:
  - `id` (String) The unique identifier of the VPC peering connection.
  - `name` (String) The name of the VPC peering connection.
  - `src_vpc_id` (String) The ID of the source VPC in the peering connection.
  - `dest_vpc_id` (String) The ID of the destination VPC in the peering connection.
  - `status` (String) The provisioning status of the peering connection (e.g., `ACTIVE`, `PENDING`, `ERROR`).
  - `peering_status` (String) The negotiation status of the peering connection (e.g., `ACTIVE`, `PENDING_ACCEPTANCE`, `REJECTED`).
  - `src_vpc_cidr` (String) The CIDR block of the source VPC.
  - `dest_vpc_cidr` (String) The CIDR block of the destination VPC.
  - `created_at` (String) The timestamp when the peering connection was created, in ISO 8601 format.
