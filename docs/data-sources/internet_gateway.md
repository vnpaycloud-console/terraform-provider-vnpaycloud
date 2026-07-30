---
page_title: "vnpaycloud_internet_gateway Data Source - VNPayCloud"
subcategory: "Networking"
description: |-
  Get information about an internet gateway in VNPayCloud.
---

# vnpaycloud_internet_gateway (Data Source)

Use this data source to get information about an existing internet gateway. Internet gateways provide connectivity between a VPC and the public internet.

At least one of `id` or `name` must be specified (enforced at plan time). If both are given, the ID must resolve to an internet gateway whose name matches. Looking up by name requires a unique match.

## Example Usage

```hcl
data "vnpaycloud_internet_gateway" "example" {
  name = "my-internet-gateway"
}

output "gateway_vpc_id" {
  value = data.vnpaycloud_internet_gateway.example.vpc_id
}
```

```hcl
data "vnpaycloud_internet_gateway" "by_id" {
  id = "igw-vwx22334"
}
```

## Schema

### Optional (filter)

- `id` (String) The ID of the internet gateway.
- `name` (String) The name of the internet gateway.

### Read-Only

- `description` (String) A human-readable description of the internet gateway.
- `vpc_id` (String) The ID of the VPC this internet gateway is attached to.
- `status` (String) The current status of the internet gateway (e.g., `ACTIVE`, `BUILD`, `ERROR`).
- `zone_id` (String) The availability zone ID where the internet gateway is deployed.
- `created_at` (String) The timestamp when the internet gateway was created, in ISO 8601 format.
