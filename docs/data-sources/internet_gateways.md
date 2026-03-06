---
page_title: "vnpaycloud_internet_gateways Data Source - VNPayCloud"
subcategory: "Networking"
description: |-
  List all internet gateways in VNPayCloud.
---

# vnpaycloud_internet_gateways (Data Source)

Use this data source to list all internet gateways in the current project.

## Example Usage

```hcl
data "vnpaycloud_internet_gateways" "all" {}

output "all_gateway_names" {
  value = data.vnpaycloud_internet_gateways.all.internet_gateways[*].name
}

output "active_gateway_ids" {
  value = [
    for igw in data.vnpaycloud_internet_gateways.all.internet_gateways :
    igw.id if igw.status == "ACTIVE"
  ]
}

output "gateways_by_vpc" {
  value = {
    for igw in data.vnpaycloud_internet_gateways.all.internet_gateways :
    igw.vpc_id => igw.id...
  }
}
```

## Schema

### Read-Only

- `internet_gateways` (List of Object) List of internet gateways. Each element contains:
  - `id` (String) The unique identifier of the internet gateway.
  - `name` (String) The name of the internet gateway.
  - `description` (String) A human-readable description of the internet gateway.
  - `vpc_id` (String) The ID of the VPC this internet gateway is attached to.
  - `status` (String) The current status of the internet gateway (e.g., `ACTIVE`, `BUILD`, `ERROR`).
  - `created_at` (String) The timestamp when the internet gateway was created, in ISO 8601 format.
