---
page_title: "vnpaycloud_vpc Data Source - VNPayCloud"
subcategory: "Networking"
description: |-
  Get information about a VPC in VNPayCloud.
---

# vnpaycloud_vpc (Data Source)

Use this data source to get information about an existing VPC.

## Example Usage

```hcl
data "vnpaycloud_vpc" "example" {
  name = "my-vpc"
}

output "vpc_cidr" {
  value = data.vnpaycloud_vpc.example.cidr
}
```

```hcl
data "vnpaycloud_vpc" "by_id" {
  id = "vpc-abc12345"
}
```

## Schema

### Optional (filter)

- `id` (String) The ID of the VPC.
- `name` (String) The name of the VPC.

### Read-Only

- `description` (String) A human-readable description of the VPC.
- `cidr` (String) The CIDR block of the VPC (e.g., `10.0.0.0/16`).
- `status` (String) The current status of the VPC (e.g., `ACTIVE`, `BUILD`, `ERROR`).
- `enable_snat` (Boolean) Whether source NAT is enabled for this VPC.
- `snat_address` (String) The SNAT IP address assigned to this VPC, if SNAT is enabled.
- `subnet_ids` (List of String) A list of subnet IDs that belong to this VPC.
- `created_at` (String) The timestamp when the VPC was created, in ISO 8601 format.
