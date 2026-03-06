---
page_title: "vnpaycloud_lb_health_monitor Resource - VNPayCloud"
subcategory: "Load Balancer"
description: |-
  Manages a health monitor for a load balancer pool within VNPayCloud.
---

# vnpaycloud_lb_health_monitor (Resource)

Manages a health monitor for a load balancer pool within VNPayCloud. Health monitors periodically check the health of pool members and automatically remove unhealthy members from the rotation until they recover.

~> **Note:** All attributes are ForceNew. Any change to health monitor configuration will destroy the existing monitor and create a new one.

## Example Usage

### HTTP health check

```hcl
resource "vnpaycloud_lb_pool" "app_pool" {
  name         = "app-backend-pool"
  listener_id  = "listener-abc12345"
  lb_algorithm = "ROUND_ROBIN"
  protocol     = "HTTP"
}

resource "vnpaycloud_lb_health_monitor" "http_check" {
  pool_id        = vnpaycloud_lb_pool.app_pool.id
  type           = "HTTP"
  delay          = 10
  timeout        = 5
  max_retries    = 3
  http_method    = "GET"
  url_path       = "/health"
  expected_codes = "200,201"
}
```

### TCP health check

```hcl
resource "vnpaycloud_lb_health_monitor" "tcp_check" {
  pool_id     = vnpaycloud_lb_pool.app_pool.id
  type        = "TCP"
  delay       = 15
  timeout     = 10
  max_retries = 3
}
```

### PING health check

```hcl
resource "vnpaycloud_lb_health_monitor" "ping_check" {
  pool_id     = "pool-xyz98765"
  type        = "PING"
  delay       = 5
  timeout     = 3
  max_retries = 5
}
```

## Schema

### Required

- `pool_id` (String, ForceNew) The ID of the pool to attach this health monitor to. Changing this creates a new health monitor.
- `type` (String, ForceNew) The type of health check to perform. Valid values are `HTTP`, `HTTPS`, `TCP`, `PING`. Changing this creates a new health monitor.
- `delay` (Number, ForceNew) The interval in seconds between consecutive health checks. Must be greater than or equal to `1`. Changing this creates a new health monitor.
- `timeout` (Number, ForceNew) The maximum number of seconds to wait for a health check response before declaring the check a failure. Must be greater than or equal to `1`. Changing this creates a new health monitor.
- `max_retries` (Number, ForceNew) The number of consecutive failed health checks before a member is marked as unhealthy. Must be between `1` and `10`. Changing this creates a new health monitor.

### Optional

- `http_method` (String, ForceNew, Computed) The HTTP method to use for the health check request. Applicable only when `type` is `HTTP` or `HTTPS`. Common values: `GET`, `HEAD`. If not specified, defaults to `GET`. Changing this creates a new health monitor.
- `url_path` (String, ForceNew, Computed) The URL path to request during the health check. Applicable only when `type` is `HTTP` or `HTTPS`. If not specified, defaults to `/`. Changing this creates a new health monitor.
- `expected_codes` (String, ForceNew, Computed) A comma-separated list of HTTP status codes that indicate a healthy response (e.g., `200`, `200,201`, `200-204`). Applicable only when `type` is `HTTP` or `HTTPS`. If not specified, defaults to `200`. Changing this creates a new health monitor.

### Read-Only

- `id` (String) The ID of the health monitor.
- `status` (String) The current status of the health monitor (e.g., `ACTIVE`, `PENDING_CREATE`, `ERROR`).

## Timeouts

- `create` - (Default `10 minutes`) Used for creating the health monitor.
- `delete` - (Default `10 minutes`) Used for deleting the health monitor.

## Import

Health monitors can be imported using the `id`:

```shell
terraform import vnpaycloud_lb_health_monitor.example <health-monitor-id>
```
