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

resource "vnpaycloud_floating_ip" "snat_ip" {}

resource "vnpaycloud_subnet_snat" "example" {
  subnet_id      = vnpaycloud_subnet.private.id
  floating_ip_id = vnpaycloud_floating_ip.snat_ip.id
}
```

## Schema

### Required

- `subnet_id` (String, ForceNew) The ID of the subnet on which to enable SNAT. Changing this creates a new resource.
- `floating_ip_id` (String, ForceNew) The ID of the floating IP to use as the SNAT address. Changing this creates a new resource.

### Read-Only

- `id` (String) The resource ID, set to `{subnet_id}/snat`.
