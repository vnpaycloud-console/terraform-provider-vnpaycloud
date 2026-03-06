---
page_title: "vnpaycloud_lb_health_monitor Data Source - VNPayCloud"
subcategory: "Load Balancer"
description: |-
  Get information about a load balancer health monitor in VNPayCloud.
---

# vnpaycloud_lb_health_monitor (Data Source)

Use this data source to get information about an existing load balancer health monitor. Health monitors periodically check the health of pool members and automatically remove unhealthy ones from the rotation.

## Example Usage

```hcl
data "vnpaycloud_lb_health_monitor" "example" {
  id = "hm-hij01123"
}

output "health_check_type" {
  value = data.vnpaycloud_lb_health_monitor.example.type
}

output "health_check_interval" {
  value = data.vnpaycloud_lb_health_monitor.example.delay
}
```

## Schema

### Required (filter)

- `id` (String) The ID of the health monitor.

### Read-Only

- `pool_id` (String) The ID of the pool this health monitor is associated with.
- `type` (String) The type of health check to perform (e.g., `HTTP`, `HTTPS`, `TCP`, `PING`, `UDP-CONNECT`).
- `delay` (Number) The time in seconds between consecutive health checks.
- `timeout` (Number) The maximum time in seconds to wait for a health check response before marking the check as failed.
- `max_retries` (Number) The number of consecutive health check failures before a member is considered unhealthy and removed from the pool.
- `http_method` (String) The HTTP method used for HTTP/HTTPS health checks (e.g., `GET`, `HEAD`). Only applicable when `type` is `HTTP` or `HTTPS`.
- `url_path` (String) The URL path to request during HTTP/HTTPS health checks (e.g., `/health`). Only applicable when `type` is `HTTP` or `HTTPS`.
- `expected_codes` (String) The expected HTTP response status codes for a successful health check (e.g., `200`, `200-299`, `200,201`). Only applicable when `type` is `HTTP` or `HTTPS`.
- `status` (String) The current provisioning status of the health monitor (e.g., `ACTIVE`, `PENDING_CREATE`, `ERROR`).
