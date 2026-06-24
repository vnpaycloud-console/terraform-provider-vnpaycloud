---
page_title: "vnpaycloud_subnet Resource - VNPayCloud"
subcategory: "Networking"
description: |-
  Manages a subnet resource within a VNPayCloud VPC.
---

# vnpaycloud_subnet (Resource)

Manages a subnet resource within a VNPayCloud VPC. Subnets partition the IP address space of a VPC into smaller segments.

~> **Note:** `name` and `dns_nameservers` can be updated in place. All other configurable fields (`vpc_id`, `cidr`, `used_by_k8s`, `used_by_si`) are immutable — changing them forces creation of a new subnet.

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
  used_by_k8s = false
}
```

## Schema

### Required

- `name` (String) The name of the subnet. Length 3–255; may only contain ASCII letters, digits, spaces, and the characters `- _ .`. Can be updated in place.
- `vpc_id` (String, ForceNew) The ID of the VPC in which to create the subnet. Changing this creates a new subnet.

### Optional

- `cidr` (String, ForceNew) The CIDR block for the subnet. If omitted, the backend auto-allocates an available `/24` within the VPC CIDR range. When provided, it must be a `/24` IPv4 network address within the VPC CIDR. Changing this creates a new subnet.
- `dns_nameservers` (List of String) Override the DNS nameservers for the subnet. If omitted, the backend assigns default nameservers (returned during read). Can be updated in place. Do not set this on a subnet with `used_by_k8s = true` — Kubernetes manages its own DNS and the value will be overridden.
- `route` (Block List) Static host routes for the subnet. Can be updated in place. Each block contains:
  - `destination` (String) Destination CIDR (e.g. `192.168.1.0/24`).
  - `nexthop` (String) Next-hop IP address. Must be a valid IP within one of your subnets.
- `used_by_k8s` (Boolean, ForceNew) Whether this subnet is reserved for Kubernetes cluster use. Defaults to `false`. Changing this creates a new subnet.
- `used_by_si` (Boolean, ForceNew) Whether this subnet is reserved for Service Instance use. Defaults to `false`. Write-only — not returned during read. Changing this creates a new subnet.

### Read-Only

- `id` (String) The ID of the subnet.
- `gateway_ip` (String) The gateway IP address of the subnet, auto-assigned by the backend (the first usable IP in the CIDR range). Cannot be set via Terraform.
- `enable_dhcp` (Boolean) Whether DHCP is enabled for the subnet. Managed by the backend; cannot be set via Terraform.
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
