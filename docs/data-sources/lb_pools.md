---
page_title: "vnpaycloud_lb_pools Data Source - VNPayCloud"
subcategory: "Load Balancer"
description: |-
  List all load balancer pools in VNPayCloud.
---

# vnpaycloud_lb_pools (Data Source)

Use this data source to list all load balancer pools in the current project.

## Example Usage

```hcl
data "vnpaycloud_lb_pools" "all" {}

output "all_pool_names" {
  value = data.vnpaycloud_lb_pools.all.pools[*].name
}

output "round_robin_pool_ids" {
  value = [
    for p in data.vnpaycloud_lb_pools.all.pools :
    p.id if p.lb_algorithm == "ROUND_ROBIN"
  ]
}
```

## Schema

### Read-Only

- `pools` (List of Object) List of pools. Each element contains:
  - `id` (String) The unique identifier of the pool.
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
  - `member` (List of Object) A list of backend members in this pool. The order is server-defined and may change between reads — key by `(address, protocol_port)` if you need to look up a specific member. Each object contains:
    - `id` (String) The ID of the pool member.
    - `name` (String) The name of the pool member.
    - `address` (String) The IP address of the backend server.
    - `protocol_port` (Number) The port on the backend server that receives traffic.
    - `weight` (Number) The weight of this member relative to others when using weighted algorithms.
    - `status` (String) Member lifecycle status: `active`, `creating`, `pending_create`, `pending_update`, `pending_delete`, `deleting`, `disabled`, `error`, `unknown`.
  - `status` (String) Pool lifecycle status: `active`, `creating`, `pending_create`, `pending_update`, `pending_delete`, `deleting`, `disabled`, `error`, `unknown`.
  - `created_at` (String) The timestamp when the pool was created, in ISO 8601 format.
