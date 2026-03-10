---
page_title: "vnpaycloud_vpcs Data Source - VNPayCloud"
subcategory: "Networking"
description: |-
  List all VPCs in VNPayCloud.
---

# vnpaycloud_vpcs (Data Source)

Use this data source to list all VPCs in the current project.

## Example Usage

```hcl
data "vnpaycloud_vpcs" "all" {}

output "all_vpc_names" {
  value = data.vnpaycloud_vpcs.all.vpcs[*].name
}

output "active_vpc_ids" {
  value = [
    for vpc in data.vnpaycloud_vpcs.all.vpcs :
    vpc.id if vpc.status == "ACTIVE"
  ]
}
```

## Schema

### Read-Only

- `vpcs` (List of Object) List of VPCs. Each element contains:
  - `id` (String) The unique identifier of the VPC.
  - `name` (String) The name of the VPC.
  - `description` (String) A human-readable description of the VPC.
  - `cidr` (String) The CIDR block of the VPC (e.g., `10.0.0.0/16`).
  - `status` (String) The current status of the VPC (e.g., `ACTIVE`, `BUILD`, `ERROR`).
  - `enable_snat` (Boolean) Whether source NAT is enabled for this VPC.
  - `snat_address` (String) The SNAT IP address assigned to this VPC, if SNAT is enabled.
  - `subnet_ids` (List of String) A list of subnet IDs that belong to this VPC.
  - `created_at` (String) The timestamp when the VPC was created, in ISO 8601 format.
