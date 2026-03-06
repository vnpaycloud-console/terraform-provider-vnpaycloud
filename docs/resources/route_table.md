---
page_title: "vnpaycloud_route_table Resource - VNPayCloud"
subcategory: "Networking"
description: |-
  Manages a route table entry within a VNPayCloud VPC.
---

# vnpaycloud_route_table (Resource)

Manages a route table entry within a VNPayCloud VPC. Each resource represents a single route that directs traffic matching a destination CIDR to a specified target.

~> **Note:** All fields are immutable. Changing any field will force creation of a new route table entry. This resource does not support import.

## Example Usage

### Route to Internet Gateway

```hcl
resource "vnpaycloud_vpc" "main" {
  name = "my-vpc"
  cidr = "10.0.0.0/16"
}

resource "vnpaycloud_internet_gateway" "igw" {
  name   = "my-igw"
  vpc_id = vnpaycloud_vpc.main.id
}

resource "vnpaycloud_route_table" "internet" {
  vpc_id      = vnpaycloud_vpc.main.id
  dest_cidr   = "0.0.0.0/0"
  target_id   = vnpaycloud_internet_gateway.igw.id
  target_type = "internet_gateway"
}
```

### Route to VPC Peering Connection

```hcl
resource "vnpaycloud_route_table" "peering_route" {
  vpc_id      = vnpaycloud_vpc.main.id
  dest_cidr   = "192.168.0.0/16"
  target_id   = vnpaycloud_vpc_peering.peer.id
  target_type = "peering_connection"
}
```

## Schema

### Required

- `vpc_id` (String, ForceNew) The ID of the VPC to which this route belongs. Changing this creates a new route.
- `dest_cidr` (String, ForceNew) The destination CIDR block for the route. Traffic matching this CIDR is forwarded to the specified target. Changing this creates a new route.
- `target_id` (String, ForceNew) The ID of the route target (e.g., internet gateway ID, peering connection ID). Changing this creates a new route.
- `target_type` (String, ForceNew) The type of the route target (e.g., `internet_gateway`, `peering_connection`). Changing this creates a new route.

### Read-Only

- `id` (String) The ID of the route table entry.
- `name` (String) The system-assigned name of the route.
- `target_name` (String) The name of the route target resource.
- `status` (String) The current status of the route.
- `created_at` (String) The creation timestamp of the route.

## Timeouts

- `create` - (Default `10 minutes`) Used for creating the route.
- `delete` - (Default `10 minutes`) Used for deleting the route.
