---
page_title: "vnpaycloud_vpc Resource - VNPayCloud"
subcategory: "Networking"
description: |-
  Manages a VPC resource within VNPayCloud.
---

# vnpaycloud_vpc (Resource)

Manages a VPC (Virtual Private Cloud) resource within VNPayCloud. A VPC provides an isolated network environment where you can launch resources.

## Example Usage

```hcl
resource "vnpaycloud_vpc" "example" {
  name        = "my-vpc"
  cidr        = "10.0.0.0/16"
  description = "My application VPC"
  enable_snat = true
}
```

## Schema

### Required

- `name` (String) The name of the VPC.
- `cidr` (String, ForceNew) The CIDR block for the VPC. Changing this creates a new VPC.

### Optional

- `description` (String) A description of the VPC.
- `enable_snat` (Boolean) Whether to enable SNAT (Source Network Address Translation) for the VPC. Defaults to `false`.

### Read-Only

- `id` (String) The ID of the VPC.
- `status` (String) The current status of the VPC.
- `snat_address` (String) The SNAT address assigned to the VPC when SNAT is enabled.
- `subnet_ids` (List of String) List of subnet IDs belonging to this VPC.
- `created_at` (String) The creation timestamp of the VPC.

## Timeouts

- `create` - (Default `10 minutes`) Used for creating the VPC.
- `delete` - (Default `10 minutes`) Used for deleting the VPC.

## Import

VPCs can be imported using the `id`:

```shell
terraform import vnpaycloud_vpc.example <vpc-id>
```
