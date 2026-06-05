---
page_title: "vnpaycloud_lb_pool Data Source - VNPayCloud"
subcategory: "Load Balancer"
description: |-
  Get information about a load balancer pool in VNPayCloud.
---

# vnpaycloud_lb_pool (Data Source)

Use this data source to get information about an existing load balancer pool (backend server group), including all of its current members and their lifecycle status.

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
- `description` (String) A human-readable description for the pool.
- `load_balancer_id` (String) The ID of the parent load balancer.
- `listener_id` (String) The ID of the listener this pool is the default of. Empty when the pool is standalone (not any listener's default).
- `lb_algorithm` (String) The load balancing algorithm used by the pool (e.g., `ROUND_ROBIN`, `LEAST_CONNECTIONS`, `SOURCE_IP`).
- `protocol` (String) The protocol used by the pool's members (e.g., `HTTP`, `HTTPS`, `TCP`, `UDP`, `PROXY`).
- `session_persistence` (List of Object) Session persistence configuration. Each object contains:
  - `type` (String) The type of session persistence (`SOURCE_IP`, `HTTP_COOKIE`, `APP_COOKIE`).
  - `cookie_name` (String) The cookie name (for `APP_COOKIE` type).
- `tls_enabled` (Boolean) Whether TLS is enabled for backend member connections.
- `member` (List of Object) A list of backend members in this pool. The order is server-defined and may change between reads — index access is not stable; key by `(address, protocol_port)` if you need to look up a specific member. Each object contains:
  - `id` (String) The ID of the pool member.
  - `name` (String) The name of the pool member.
  - `address` (String) The IP address of the backend server.
  - `protocol_port` (Number) The port on the backend server that receives traffic.
  - `weight` (Number) The weight of this member relative to others when using weighted algorithms. Defaults to `1`.
  - `status` (String) Member lifecycle status: `active`, `creating`, `pending_create`, `pending_update`, `pending_delete`, `deleting`, `disabled`, `error`, `unknown`.
- `status` (String) Pool lifecycle status: `active`, `creating`, `pending_create`, `pending_update`, `pending_delete`, `deleting`, `disabled`, `error`, `unknown`.
- `created_at` (String) The timestamp when the pool was created, in ISO 8601 format.
