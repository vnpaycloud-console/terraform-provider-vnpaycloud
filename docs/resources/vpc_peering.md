---
page_title: "vnpaycloud_vpc_peering Resource - VNPayCloud"
subcategory: "Networking"
description: |-
  Manages a VPC peering connection within VNPayCloud.
---

# vnpaycloud_vpc_peering (Resource)

Manages a VPC peering connection within VNPayCloud. A VPC peering connection establishes a direct network route between two VPCs, allowing resources in each VPC to communicate with each other as if they are within the same network.

~> **Note:** VPC peering is bidirectional. The provider automatically manages both directions of the peering connection. The `src_vpc_id`, `dest_vpc_id`, and `description` fields are immutable. Changing them will force creation of a new peering connection.

## Example Usage

```hcl
resource "vnpaycloud_vpc" "vpc_a" {
  name = "vpc-a"
  cidr = "10.1.0.0/16"
}

resource "vnpaycloud_vpc" "vpc_b" {
  name = "vpc-b"
  cidr = "10.2.0.0/16"
}

resource "vnpaycloud_vpc_peering" "example" {
  name        = "vpc-a-to-vpc-b"
  src_vpc_id  = vnpaycloud_vpc.vpc_a.id
  dest_vpc_id = vnpaycloud_vpc.vpc_b.id
  description = "Peering between VPC A and VPC B"
}

# Add routes so instances can use the peering connection
resource "vnpaycloud_route_table" "a_to_b" {
  vpc_id      = vnpaycloud_vpc.vpc_a.id
  dest_cidr   = vnpaycloud_vpc.vpc_b.cidr
  target_id   = vnpaycloud_vpc_peering.example.id
  target_type = "peering_connection"
}

resource "vnpaycloud_route_table" "b_to_a" {
  vpc_id      = vnpaycloud_vpc.vpc_b.id
  dest_cidr   = vnpaycloud_vpc.vpc_a.cidr
  target_id   = vnpaycloud_vpc_peering.example.id
  target_type = "peering_connection"
}
```

## Schema

### Required

- `src_vpc_id` (String, ForceNew) The ID of the source VPC initiating the peering request. Changing this creates a new peering connection.
- `dest_vpc_id` (String, ForceNew) The ID of the destination VPC accepting the peering request. Changing this creates a new peering connection.

### Optional

- `name` (String, Computed) The name of the peering connection. If not set, a name is auto-generated. Can be updated after creation.
- `description` (String, ForceNew) A description of the peering connection. Changing this creates a new peering connection.

### Read-Only

- `id` (String) The ID of the VPC peering connection.
- `status` (String) The current provisioning status of the peering connection.
- `peering_status` (String) The peering-specific status indicating whether the connection is established.
- `src_vpc_cidr` (String) The CIDR block of the source VPC.
- `dest_vpc_cidr` (String) The CIDR block of the destination VPC.
- `created_at` (String) The creation timestamp of the peering connection.

## Timeouts

- `create` - (Default `10 minutes`) Used for creating the VPC peering connection.
- `delete` - (Default `10 minutes`) Used for deleting the VPC peering connection.

## Import

VPC peering connections can be imported using the `id`:

```shell
terraform import vnpaycloud_vpc_peering.example <peering-id>
```
