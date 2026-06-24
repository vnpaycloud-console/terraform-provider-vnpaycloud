---
page_title: "vnpaycloud_vpc Resource - VNPayCloud"
subcategory: "Networking"
description: |-
  Manages a VPC resource within VNPayCloud.
---

# vnpaycloud_vpc (Resource)

Manages a VPC (Virtual Private Cloud) resource within VNPayCloud. A VPC provides an isolated network environment where you can launch resources.

## Example Usage

### Auto-Generated CIDR

If `cidr` is omitted, VNPayCloud allocates an available `/16` private CIDR for the VPC and the provider stores the generated value in state.

```hcl
resource "vnpaycloud_vpc" "example" {
  name        = "my-vpc"
  description = "My application VPC"
}
```

### Explicit CIDR

```hcl
resource "vnpaycloud_vpc" "example" {
  name        = "my-vpc"
  cidr        = "10.0.0.0/16"
  description = "My application VPC"
}
```

## Schema

### Required

- `name` (String) The name of the VPC. Length 3–255; may only contain ASCII letters, digits, spaces, and the characters `- _ .`. Can be updated in place.

### Optional

- `cidr` (String, ForceNew) The CIDR block for the VPC. If omitted, VNPayCloud automatically allocates an available `/16` private CIDR and returns it during read. When provided, it must be a `/16` IPv4 network address in a private range (`10.0.0.0/8`, `172.16.0.0/12`, `192.168.0.0/16`). Changing this creates a new VPC.
- `description` (String) A description of the VPC. Set at creation only; changes after creation are ignored (description cannot be updated via the API — only from the console Network page).

### Read-Only

- `id` (String) The ID of the VPC.
- `status` (String) The current status of the VPC.
- `enable_snat` (Boolean) Whether SNAT (Source Network Address Translation) is enabled for the VPC. Read-only — SNAT is managed from the console Network page, not via Terraform.
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
