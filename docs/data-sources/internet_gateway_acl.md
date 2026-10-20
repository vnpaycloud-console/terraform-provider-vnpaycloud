---
page_title: "vnpaycloud_internet_gateway_acl Data Source - VNPayCloud"
subcategory: "Networking"
description: |-
  Get the ACL (CIDR blacklist) of an internet gateway in VNPayCloud.
---

# vnpaycloud_internet_gateway_acl (Data Source)

Use this data source to read the ACL of an existing internet gateway. The ACL is a blacklist of public IPv4 CIDR blocks whose traffic is dropped at the gateway.

An internet gateway with no blacklisted CIDR block returns an empty `match` and `total_match` of `0`.

## Example Usage

```hcl
data "vnpaycloud_internet_gateway" "main" {
  name = "my-internet-gateway"
}

data "vnpaycloud_internet_gateway_acl" "main" {
  internet_gateway_id = data.vnpaycloud_internet_gateway.main.id
}

output "blocked_cidrs" {
  value = data.vnpaycloud_internet_gateway_acl.main.match
}

output "blocked_cidr_count" {
  value = data.vnpaycloud_internet_gateway_acl.main.total_match
}
```

## Schema

### Required

- `internet_gateway_id` (String) The ID of the internet gateway whose ACL is read.

### Read-Only

- `id` (String) The ID of the internet gateway the ACL belongs to.
- `match` (Set of String) The blacklisted CIDR blocks.
- `total_match` (Number) The number of blacklisted CIDR blocks.
- `zone_id` (String) The availability zone ID where the internet gateway is deployed.
