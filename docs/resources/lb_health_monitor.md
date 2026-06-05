---
page_title: "vnpaycloud_lb_health_monitor Resource - VNPayCloud"
subcategory: "Load Balancer"
description: |-
  Manages a health monitor for a load balancer pool within VNPayCloud.
---

# vnpaycloud_lb_health_monitor (Resource)

Manages a health monitor for a load balancer pool within VNPayCloud. Health monitors periodically check the health of pool members and automatically remove unhealthy members from the rotation until they recover.

Health monitor type must be compatible with the pool's protocol — for example a `UDP` pool only accepts `UDP-CONNECT`, `SCTP`, `TCP`, or `HTTP` monitor types. The backend rejects incompatible combinations and returns a clear error.

## Example Usage

### HTTP health check

```hcl
resource "vnpaycloud_lb_pool" "app_pool" {
  name             = "app-backend-pool"
  load_balancer_id = "lb-xyz98765"
  listener_id      = "listener-abc12345"
  lb_algorithm     = "ROUND_ROBIN"
  protocol         = "HTTP"
}

resource "vnpaycloud_lb_health_monitor" "http_check" {
  name             = "http-health-check"
  pool_id          = vnpaycloud_lb_pool.app_pool.id
  type             = "HTTP"
  delay            = 10
  timeout          = 5
  max_retries      = 3
  max_retries_down = 3
  http_method      = "GET"
  url_path         = "/healthz"
  expected_codes   = "200,201"
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

## Schema

### Required

- `pool_id` (String, ForceNew) The ID of the pool to attach this health monitor to.
- `type` (String, ForceNew) The type of health check. Valid values:
  - `HTTP`, `HTTPS` — application-layer probe using an HTTP/HTTPS request.
  - `PING` — ICMP echo.
  - `TCP` — TCP socket open.
  - `TLS-HELLO` — TLS handshake probe.
  - `UDP-CONNECT` — UDP connect probe (UDP/SCTP pools).
  - `SCTP` — SCTP probe.

  Use the hyphenated values shown here; the provider translates `TLS-HELLO` / `UDP-CONNECT` to the backend's underscore form automatically.

  **Pool ↔ Monitor type compatibility** (incompatible combinations are rejected by the backend):

  | Pool `protocol` \ Monitor `type` | `HTTP` | `HTTPS` | `TCP` | `PING` | `TLS-HELLO` | `UDP-CONNECT` | `SCTP` |
  |---|:---:|:---:|:---:|:---:|:---:|:---:|:---:|
  | `HTTP` | ✓ | ✓ | ✓ | ✓ | ✓ |  |  |
  | `HTTPS` | ✓ | ✓ | ✓ | ✓ | ✓ |  |  |
  | `TCP` | ✓ | ✓ | ✓ | ✓ | ✓ |  |  |
  | `PROXY` | ✓ | ✓ | ✓ | ✓ | ✓ |  |  |
  | `UDP` | ✓ |  | ✓ |  |  | ✓ | ✓ |

- `delay` (Number) Interval in seconds between consecutive probes. Must be `>= 1`.
- `timeout` (Number) Maximum seconds to wait for a probe response. Must be `>= 1` **and** `<= delay` (enforced at plan time).
- `max_retries` (Number) **Rise threshold** — consecutive *successful* probes before a previously-unhealthy member is marked healthy again. Range `1`–`10`. The name reads like "retries on failure" but counts successes. See `max_retries_down` (in *Optional* below) for the corresponding **fall threshold** that demotes a healthy member after consecutive failures.

### Optional

- `name` (String, Optional, Computed) Name of the monitor. Auto-generated when empty. If set: length `0`–`250`, no leading/trailing whitespace.
- `max_retries_down` (Number, Computed) **Fall threshold** — consecutive *failed* probes before a healthy member is marked unhealthy. Range `1`–`10`. Counterpart to `max_retries` (the rise threshold above).
- `http_method` (String, Computed) HTTP method used for HTTP/HTTPS probes. One of `GET`, `POST`, `PUT`, `DELETE`, `HEAD`, `OPTIONS`, `PATCH`, `CONNECT`, `TRACE`. **Only valid when `type` is `HTTP` or `HTTPS` — the server rejects this field for other types.**
- `url_path` (String, Computed) URL path for HTTP/HTTPS probes. Must start with `/`. **Only valid when `type` is `HTTP` or `HTTPS`.**
- `expected_codes` (String, Computed) HTTP status codes that indicate a healthy response. Formats: single code (`200`), comma list (`200,201,302`), or range (`200-299`). **Only valid when `type` is `HTTP` or `HTTPS`.**

### Read-Only

- `id` (String) The ID of the health monitor.
- `status` (String) Lifecycle status: `active`, `creating`, `pending_create`, `pending_update`, `pending_delete`, `deleting`, `disabled`, `error`, `unknown`.

## In-place updates

The following attributes can be updated without recreating the monitor:
`name`, `delay`, `timeout`, `max_retries`, `max_retries_down`, `http_method`, `url_path`, `expected_codes`.

`pool_id` and `type` are `ForceNew` — changing them destroys and recreates the monitor.

## Timeouts

- `create` - (Default `10 minutes`)
- `update` - (Default `10 minutes`)
- `delete` - (Default `10 minutes`)

~> **Rate limit:** see [Rate limits](../index.md#rate-limits) — applies to all create/update/delete on this resource type.

## Import

```shell
terraform import vnpaycloud_lb_health_monitor.example <health-monitor-id>
```
