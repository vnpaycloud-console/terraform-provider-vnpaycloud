---
page_title: "vnpaycloud_subnet Resource - VNPayCloud"
subcategory: "Networking"
description: |-
  Manages a subnet resource within a VNPayCloud VPC.
---

# vnpaycloud_subnet (Resource)

Manages a subnet resource within a VNPayCloud VPC. Subnets partition the IP address space of a VPC into smaller segments.

~> **Note:** All fields except `description` are immutable. Changing any required field will force creation of a new subnet.

## Example Usage

```hcl
resource "vnpaycloud_vpc" "main" {
  name = "my-vpc"
  cidr = "10.0.0.0/16"
}

resource "vnpaycloud_subnet" "example" {
  name        = "my-subnet"
  vpc_id      = vnpaycloud_vpc.main.id
  cidr        = "10.0.1.0/24"
  gateway_ip  = "10.0.1.1"
  enable_dhcp = true
  used_by_k8s = false
}
```

## Schema

### Required

- `name` (String, ForceNew) The name of the subnet. Changing this creates a new subnet.
- `vpc_id` (String, ForceNew) The ID of the VPC in which to create the subnet. Changing this creates a new subnet.
- `cidr` (String, ForceNew) The CIDR block for the subnet. Must be within the VPC CIDR range. Changing this creates a new subnet.

### Optional

- `gateway_ip` (String, ForceNew, Computed) The gateway IP address for the subnet. If not specified, the first usable IP in the CIDR range is used. Changing this creates a new subnet.
- `enable_dhcp` (Boolean, ForceNew) Whether to enable DHCP for the subnet. Defaults to `true`. Changing this creates a new subnet.
- `used_by_k8s` (Boolean, ForceNew) Whether this subnet is reserved for Kubernetes cluster use. Defaults to `false`. Changing this creates a new subnet.

### Read-Only

- `id` (String) The ID of the subnet.
- `status` (String) The current status of the subnet.
- `created_at` (String) The creation timestamp of the subnet.

## Timeouts

- `create` - (Default `10 minutes`) Used for creating the subnet.
- `delete` - (Default `10 minutes`) Used for deleting the subnet.

## Import

Subnets can be imported using the `id`:

```shell
terraform import vnpaycloud_subnet.example <subnet-id>
```
