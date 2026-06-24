---
page_title: "vnpaycloud_vpc_peering Resource - VNPayCloud"
subcategory: "Networking"
description: |-
  Manages a VPC peering connection within VNPayCloud.
---

# vnpaycloud_vpc_peering (Resource)

Manages a VPC peering connection within VNPayCloud. A VPC peering connection establishes a direct network route between two VPCs, allowing resources in each VPC to communicate with each other as if they are within the same network.

~> **Note:** VPC peering is bidirectional. The provider automatically manages both directions of the peering connection — creating the resource creates both directions, and destroying it removes both. The `src_vpc_id` and `dest_vpc_id` fields are immutable; changing either forces creation of a new peering connection.

~> **Note:** Each VPC must already contain at least one network (subnet) before it can be peered. Creating a peering against an empty VPC fails with `Please init at least 1 network with your vpc`. The two VPCs must also have non-overlapping CIDR ranges, and a given pair of VPCs can only be peered once.

~> **Note:** Peering is supported only between VPCs in the same organization.

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

# Each VPC must have at least one network (subnet) before it can be peered.
resource "vnpaycloud_subnet" "subnet_a" {
  name   = "subnet-a"
  vpc_id = vnpaycloud_vpc.vpc_a.id
  cidr   = "10.1.1.0/24"
}

resource "vnpaycloud_subnet" "subnet_b" {
  name   = "subnet-b"
  vpc_id = vnpaycloud_vpc.vpc_b.id
  cidr   = "10.2.1.0/24"
}

resource "vnpaycloud_vpc_peering" "example" {
  name        = "vpc-a-to-vpc-b"
  src_vpc_id  = vnpaycloud_vpc.vpc_a.id
  dest_vpc_id = vnpaycloud_vpc.vpc_b.id
  description = "Peering between VPC A and VPC B"

  depends_on = [vnpaycloud_subnet.subnet_a, vnpaycloud_subnet.subnet_b]
}

# Add routes so instances can use the peering connection. A peering has two
# directional connections; a VPC's route must target the connection whose
# source is that VPC. The source VPC uses `id`; the destination VPC uses
# `reverse_peering_id`.
resource "vnpaycloud_route_table" "a_to_b" {
  vpc_id      = vnpaycloud_vpc.vpc_a.id
  dest_cidr   = vnpaycloud_vpc.vpc_b.cidr
  target_id   = vnpaycloud_vpc_peering.example.id
  target_type = "peering_connection"
}

resource "vnpaycloud_route_table" "b_to_a" {
  vpc_id      = vnpaycloud_vpc.vpc_b.id
  dest_cidr   = vnpaycloud_vpc.vpc_a.cidr
  target_id   = vnpaycloud_vpc_peering.example.reverse_peering_id
  target_type = "peering_connection"
}
```

### Without a name (auto-generated)

`name` is optional. When omitted, the backend assigns `<src-vpc>_to_<dest-vpc>`
and that value is read back into state, so subsequent plans show no drift.

```hcl
resource "vnpaycloud_vpc_peering" "auto" {
  src_vpc_id  = vnpaycloud_vpc.vpc_a.id
  dest_vpc_id = vnpaycloud_vpc.vpc_b.id

  depends_on = [vnpaycloud_subnet.subnet_a, vnpaycloud_subnet.subnet_b]
}

# e.g. name becomes "vpc-a_to_vpc-b"
output "peering_name" {
  value = vnpaycloud_vpc_peering.auto.name
}
```

## Schema

### Required

- `src_vpc_id` (String, ForceNew) The ID of the source VPC initiating the peering request. Changing this creates a new peering connection.
- `dest_vpc_id` (String, ForceNew) The ID of the destination VPC accepting the peering request. Changing this creates a new peering connection.

### Optional

- `name` (String, Optional) The name of the peering connection. Optional — you may set it or leave it out. If left out, the backend auto-generates one (`<src-vpc>_to_<dest-vpc>`) and that value is recorded in state. If set, it is applied at create and can be updated in place afterwards.
- `description` (String) A description of the peering connection, applied at create time only. The backend does not return `description` and has no update path for it, so changes after creation are ignored (no drift, no recreate), and it is left empty when the resource is imported.

### Read-Only

- `id` (String) The ID of the VPC peering connection.
- `status` (String) The current provisioning status of the peering connection.
- `peering_status` (String) The peering-specific status indicating whether the connection is established.
- `src_vpc_cidr` (String) The CIDR block of the source VPC.
- `dest_vpc_cidr` (String) The CIDR block of the destination VPC.
- `created_at` (String) The creation timestamp of the peering connection.
- `reverse_peering_id` (String) The ID of the reverse-direction peering connection (destination → source). The provider populates this so it can clean up both directions on destroy; use it as the `target_id` for a route table on the destination VPC.

## Timeouts

- `create` - (Default `10 minutes`) Used for creating the VPC peering connection.
- `delete` - (Default `10 minutes`) Used for deleting the VPC peering connection.

## Import

VPC peering connections can be imported using the `id`:

```shell
terraform import vnpaycloud_vpc_peering.example <peering-id>
```
