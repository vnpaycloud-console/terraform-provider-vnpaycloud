---
page_title: "vnpaycloud_lb_health_monitors Data Source - VNPayCloud"
subcategory: "Load Balancer"
description: |-
  List all load balancer health monitors in VNPayCloud.
---

# vnpaycloud_lb_health_monitors (Data Source)

Use this data source to list all load balancer health monitors in the current project.

## Example Usage

```hcl
data "vnpaycloud_lb_health_monitors" "all" {}

output "all_health_monitor_names" {
  value = data.vnpaycloud_lb_health_monitors.all.health_monitors[*].name
}

output "tcp_health_monitor_ids" {
  value = [
    for m in data.vnpaycloud_lb_health_monitors.all.health_monitors :
    m.id if m.type == "TCP"
  ]
}
```

## Schema

### Read-Only

- `health_monitors` (List of Object) List of health monitors. Each element contains:
  - `id` (String) The unique identifier of the health monitor.
  - `name` (String) Name of the health monitor.
  - `pool_id` (String) The ID of the pool this health monitor is associated with.
  - `type` (String) The type of health check to perform (e.g., `HTTP`, `HTTPS`, `TCP`, `PING`, `TLS-HELLO`, `UDP-CONNECT`, `SCTP`).
  - `delay` (Number) The time in seconds between consecutive health checks.
  - `timeout` (Number) The maximum time in seconds to wait for a health check response before marking the check as failed.
  - `max_retries` (Number) Consecutive successful probes (rise) before a previously-unhealthy member is reinstated as healthy.
  - `max_retries_down` (Number) Consecutive failed probes (fall) before a member is marked unhealthy and removed from the pool.
  - `http_method` (String) The HTTP method used for HTTP/HTTPS health checks (e.g., `GET`, `HEAD`). Only applicable when `type` is `HTTP` or `HTTPS`.
  - `url_path` (String) The URL path to request during HTTP/HTTPS health checks (e.g., `/health`). Only applicable when `type` is `HTTP` or `HTTPS`.
  - `expected_codes` (String) The expected HTTP response status codes for a successful health check (e.g., `200`, `200-299`, `200,201`). Only applicable when `type` is `HTTP` or `HTTPS`.
  - `status` (String) Lifecycle status: `active`, `creating`, `pending_create`, `pending_update`, `pending_delete`, `deleting`, `disabled`, `error`, `unknown`.
