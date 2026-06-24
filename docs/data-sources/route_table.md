---
page_title: "vnpaycloud_route_table Data Source - VNPayCloud"
subcategory: "Networking"
description: |-
  Get information about a route table entry in VNPayCloud.
---

# vnpaycloud_route_table (Data Source)

Use this data source to get information about an existing route table entry by its ID.

## Example Usage

```hcl
data "vnpaycloud_route_table" "example" {
  id = "rtb-bcd66778"
}

output "route_target" {
  value = data.vnpaycloud_route_table.example.target_id
}
```

## Schema

### Required (filter)

- `id` (String) The ID of the route table entry.

### Read-Only

- `vpc_id` (String) The ID of the VPC the route belongs to.
- `dest_cidr` (String) The destination CIDR block of the route.
- `target_id` (String) The ID of the route target.
- `target_type` (String) The type of the route target (`internet_gateway`, `peering_connection`, `service_instance`, `vpn_gateway`).
- `target_name` (String) The name of the route target resource. Populated only for `peering_connection` targets; empty for other target types.
- `name` (String) The system-assigned name of the route.
- `status` (String) The current status of the route.
- `created_at` (String) The creation timestamp of the route, in ISO 8601 format.
