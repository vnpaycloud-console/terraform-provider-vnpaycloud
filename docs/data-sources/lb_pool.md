---
page_title: "vnpaycloud_lb_pool Data Source - VNPayCloud"
subcategory: "Load Balancer"
description: |-
  Get information about a load balancer pool in VNPayCloud.
---

# vnpaycloud_lb_pool (Data Source)

Use this data source to get information about an existing load balancer pool (backend server group), including all of its current members and their health statuses.

## Example Usage

```hcl
data "vnpaycloud_lb_pool" "example" {
  id = "pool-efg88990"
}

output "pool_algorithm" {
  value = data.vnpaycloud_lb_pool.example.lb_algorithm
}

output "pool_members" {
  value = data.vnpaycloud_lb_pool.example.member
}
```

## Schema

### Required (filter)

- `id` (String) The ID of the load balancer pool.

### Read-Only

- `name` (String) The name of the pool.
- `listener_id` (String) The ID of the listener this pool is the default pool for.
- `lb_algorithm` (String) The load balancing algorithm used by the pool (e.g., `ROUND_ROBIN`, `LEAST_CONNECTIONS`, `SOURCE_IP`).
- `protocol` (String) The protocol used by the pool's members (e.g., `HTTP`, `HTTPS`, `TCP`).
- `member` (List of Object) A list of backend members in this pool. Each object contains:
  - `id` (String) The ID of the pool member.
  - `name` (String) The name of the pool member.
  - `address` (String) The IP address of the backend server.
  - `protocol_port` (Number) The port on the backend server that receives traffic.
  - `weight` (Number) The weight of this member relative to others when using weighted algorithms. Defaults to `1`.
  - `status` (String) The operating status of the member (e.g., `ONLINE`, `OFFLINE`, `NO_MONITOR`).
- `status` (String) The current provisioning status of the pool (e.g., `ACTIVE`, `PENDING_CREATE`, `ERROR`).
- `created_at` (String) The timestamp when the pool was created, in ISO 8601 format.
