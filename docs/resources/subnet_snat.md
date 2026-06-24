---
page_title: "vnpaycloud_subnet_snat Resource - VNPayCloud"
subcategory: "Networking"
description: |-
  Enables SNAT on a VNPayCloud subnet using a floating IP.
---

# vnpaycloud_subnet_snat (Resource)

Enables SNAT (Source Network Address Translation) on a VNPayCloud subnet by associating it with a floating IP address. This allows instances in the subnet to access the internet using the floating IP as their source address.

~> **Note:** This resource does not support import. Destroying this resource disables SNAT on the subnet.

## Example Usage

```hcl
resource "vnpaycloud_vpc" "main" {
  name = "my-vpc"
  cidr = "10.0.0.0/16"
}

resource "vnpaycloud_subnet" "private" {
  name   = "my-private-subnet"
  vpc_id = vnpaycloud_vpc.main.id
  cidr   = "10.0.2.0/24"
}

resource "vnpaycloud_internet_gateway" "igw" {
  name       = "my-igw"
  vpc_id     = vnpaycloud_vpc.main.id
  depends_on = [vnpaycloud_subnet.private]
}

resource "vnpaycloud_route_table" "internet" {
  vpc_id      = vnpaycloud_vpc.main.id
  dest_cidr   = "0.0.0.0/0"
  target_id   = vnpaycloud_internet_gateway.igw.id
  target_type = "internet_gateway"
}

resource "vnpaycloud_floating_ip" "snat_ip" {
  vpc_id = vnpaycloud_vpc.main.id

  depends_on = [vnpaycloud_internet_gateway.igw]
}

resource "vnpaycloud_subnet_snat" "example" {
  subnet_id      = vnpaycloud_subnet.private.id
  floating_ip_id = vnpaycloud_floating_ip.snat_ip.id

  depends_on = [vnpaycloud_route_table.internet]
}
```

~> **Note:** The floating IP must be associated with the same VPC as the subnet before enabling subnet SNAT. An unassociated floating IP, or a floating IP associated with a different VPC, will be rejected by the API.

~> **Note:** Use exactly one `vnpaycloud_subnet_snat` per subnet — a subnet has a single SNAT slot, and all such resources share the id `{subnet_id}/snat`. A second resource for the same subnet is rejected on apply with `SNAT is already enabled ... disable it before switching to floating IP ...`.

## Schema

### Required

- `subnet_id` (String, ForceNew) The ID of the subnet on which to enable SNAT. Changing this creates a new resource.
- `floating_ip_id` (String, ForceNew) The ID of the floating IP to use as the SNAT address. Changing this creates a new resource.

### Read-Only

- `id` (String) The resource ID, set to `{subnet_id}/snat`.

## Import

Subnet SNAT resources do not support import.
