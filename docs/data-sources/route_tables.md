---
page_title: "vnpaycloud_route_tables Data Source - VNPayCloud"
subcategory: "Networking"
description: |-
  List route table entries in VNPayCloud, optionally filtered by VPC.
---

# vnpaycloud_route_tables (Data Source)

Use this data source to list route table entries in the current project, optionally filtered by VPC.

## Example Usage

```hcl
data "vnpaycloud_route_tables" "by_vpc" {
  vpc_id = vnpaycloud_vpc.main.id
}

output "route_dest_cidrs" {
  value = data.vnpaycloud_route_tables.by_vpc.route_tables[*].dest_cidr
}
```

## Schema

### Optional

- `vpc_id` (String) Filter the routes by VPC ID. If omitted, lists route tables across the project.

### Read-Only

- `route_tables` (List of Object) List of route table entries. Each element contains:
  - `id` (String) The ID of the route table entry.
  - `vpc_id` (String) The ID of the VPC the route belongs to.
  - `dest_cidr` (String) The destination CIDR block of the route.
  - `target_id` (String) The ID of the route target.
  - `target_type` (String) The type of the route target (`internet_gateway`, `peering_connection`, `service_instance`, `vpn_gateway`).
  - `target_name` (String) The name of the route target resource. Populated only for `peering_connection` targets; empty for other target types.
  - `name` (String) The system-assigned name of the route.
  - `status` (String) The current status of the route.
  - `created_at` (String) The creation timestamp of the route, in ISO 8601 format.
