---
page_title: "vnpaycloud_lb_pool Resource - VNPayCloud"
subcategory: "Load Balancer"
description: |-
  Manages a load balancer pool within VNPayCloud.
---

# vnpaycloud_lb_pool (Resource)

Manages a backend pool for a load balancer within VNPayCloud. A pool groups together a set of backend members (servers) and defines the algorithm used to distribute traffic among them.

## Relationships

```
Load Balancer (1) ───── (N) Pool          # parent: pool.load_balancer_id (Required)
Load Balancer (1) ───── (N) Listener
Listener      (1) ───── (0..1) Pool       # listener.default_pool_id (Optional)
```

- A **pool belongs to a load balancer**, not to a listener — set `load_balancer_id` (Required).
- A **listener may have at most one default pool**. Attach via `listener_id` on the pool or `default_pool_id` on `vnpaycloud_lb_listener` — both converge and neither recreates. Via the pool's `listener_id` the first `plan` after create is clean (no drift); declaring `default_pool_id` on the listener instead shows one benign in-place sync that settles after a second `apply` (no recreate).
- A pool can exist **without any listener attachment** (e.g. as an L7-policy `redirect_pool_id` target, or held in reserve to swap into a listener later).
- Attempting to create a second pool with the same `listener_id` is **rejected server-side** with a clear error. Swap default via `vnpaycloud_lb_listener.default_pool_id` instead.

## Example Usage

### Pool attached to a listener as its default

```hcl
resource "vnpaycloud_lb_loadbalancer" "lb" {
  name      = "app-lb"
  subnet_id = "subnet-abc12345"
  flavor    = "t1-small"
}

resource "vnpaycloud_lb_listener" "http" {
  name             = "http-listener"
  load_balancer_id = vnpaycloud_lb_loadbalancer.lb.id
  protocol         = "HTTP"
  protocol_port    = 80
}

resource "vnpaycloud_lb_pool" "app_pool" {
  name             = "app-backend-pool"
  description      = "Backend pool for HTTP traffic"
  load_balancer_id = vnpaycloud_lb_loadbalancer.lb.id
  listener_id      = vnpaycloud_lb_listener.http.id   # auto-set listener.default_pool_id at create
  lb_algorithm     = "ROUND_ROBIN"
  protocol         = "HTTP"

  session_persistence {
    type = "SOURCE_IP"
  }

  # Replace these placeholder IPs with real backend addresses in your subnet.
  member {
    address       = "10.0.1.10"
    protocol_port = 8080
    weight        = 1
  }

  member {
    address       = "10.0.1.11"
    protocol_port = 8080
    weight        = 2
  }
}
```

### Standalone pool — not attached to any listener

A pool that exists under the load balancer but is not the default of any listener. Typical uses: an L7 policy `redirect_pool_id` target, or a spare pool you'll later swap into a listener.

```hcl
resource "vnpaycloud_lb_pool" "spare_pool" {
  name             = "spare-pool"
  load_balancer_id = vnpaycloud_lb_loadbalancer.lb.id
  # no listener_id — pool stays unattached
  lb_algorithm     = "ROUND_ROBIN"
  protocol         = "HTTP"
}
```

### APP_COOKIE session persistence

```hcl
resource "vnpaycloud_lb_pool" "stateful_pool" {
  name             = "stateful-pool"
  load_balancer_id = vnpaycloud_lb_loadbalancer.lb.id
  listener_id      = vnpaycloud_lb_listener.http.id
  lb_algorithm     = "LEAST_CONNECTIONS"
  protocol         = "HTTP"

  session_persistence {
    type        = "APP_COOKIE"
    cookie_name = "JSESSIONID"
  }
}
```

### TCP pool

```hcl
resource "vnpaycloud_lb_pool" "tcp_pool" {
  name             = "tcp-backend-pool"
  load_balancer_id = vnpaycloud_lb_loadbalancer.lb.id
  listener_id      = vnpaycloud_lb_listener.tcp.id
  lb_algorithm     = "LEAST_CONNECTIONS"
  protocol         = "TCP"

  member {
    address       = "192.168.1.100"
    protocol_port = 3000
  }
}
```

## Schema

### Required

- `name` (String) The name of the pool. Length `3`–`250`, no leading/trailing whitespace. Unique per load balancer.
- `load_balancer_id` (String, ForceNew) The ID of the parent load balancer. Pools belong to a load balancer (1 LB → many pools); a pool is independent of any listener unless you opt in via `listener_id`.
- `lb_algorithm` (String) Load balancing algorithm. One of `ROUND_ROBIN`, `LEAST_CONNECTIONS`, `SOURCE_IP`.
- `protocol` (String, ForceNew) Backend protocol. One of `HTTP`, `HTTPS`, `TCP`, `UDP`, `PROXY`. Must be compatible with the listener's protocol.

  **Listener ↔ Pool protocol compatibility** (incompatible combinations are rejected by the backend):

  | Listener `protocol` \ Pool `protocol` | `HTTP` | `HTTPS` | `TCP` | `UDP` | `PROXY` |
  |---|:---:|:---:|:---:|:---:|:---:|
  | `HTTP` | ✓ |  |  |  | ✓ |
  | `HTTPS` | ✓ |  |  |  | ✓ |
  | `TCP` | ✓ | ✓ | ✓ |  | ✓ |
  | `UDP` |  |  |  | ✓ |  |

  **Pool ↔ Health monitor type compatibility** (relevant when attaching a `vnpaycloud_lb_health_monitor`):

  | Pool `protocol` \ Monitor `type` | `HTTP` | `HTTPS` | `TCP` | `PING` | `TLS-HELLO` | `UDP-CONNECT` | `SCTP` |
  |---|:---:|:---:|:---:|:---:|:---:|:---:|:---:|
  | `HTTP` | ✓ | ✓ | ✓ | ✓ | ✓ |  |  |
  | `HTTPS` | ✓ | ✓ | ✓ | ✓ | ✓ |  |  |
  | `TCP` | ✓ | ✓ | ✓ | ✓ | ✓ |  |  |
  | `PROXY` | ✓ | ✓ | ✓ | ✓ | ✓ |  |  |
  | `UDP` | ✓ |  | ✓ |  |  | ✓ | ✓ |

### Optional

- `listener_id` (String, Optional, Computed, ForceNew) Convenience for setting this pool as the listener's default at create time.

  **TL;DR:** shortcut for "this pool is the default of this listener"; the target listener must have no default yet.

  When set, the provider issues `UpdateListener default_pool_id=<this pool>` after the pool is active. The listener must currently have no default pool — a listener accepts **at most one** default; if the listener already has one, create will fail with a clear error. To swap an existing default, update `vnpaycloud_lb_listener.default_pool_id` instead.
  - `Computed` back-pointer: leaving `listener_id` unset keeps the backend value (no spurious recreate). A `ForceNew` recreate happens only if you explicitly set `listener_id` to a value differing from the backend (e.g. moving the pool to another listener).
- `description` (String) A human-readable description. Length `0`–`255`.
- `session_persistence` (Block, Max 1) Configure sticky sessions:
  - `type` (String, Required) `SOURCE_IP`, `HTTP_COOKIE`, or `APP_COOKIE`.
  - `cookie_name` (String) Required when `type = APP_COOKIE` (enforced at plan time via `CustomizeDiff`).
- `tls_enabled` (Boolean, Optional, Computed) Whether TLS is enabled for backend member connections.
- `member` (Block Set) Backend members. Each member is keyed by `(address, protocol_port, weight)` — no ordering noise on plan diffs. **Note**: changing `weight` is detected as remove+add of the member (full set PUT to backend, no real connection drain), not update-in-place. Same address+port with different weight = different element in the Set. Per-member fields:
  - `address` (String, Required) Member IP. Must be a valid IP (validated at plan time).
  - `protocol_port` (Number, Required) Backend port. Range `1`–`65535`.
  - `weight` (Number, Optional, Default `1`) Relative weight, `0`–`256`. `0` drains traffic from the member.
  - `id` (String, Read-Only) Server-assigned member ID.
  - `name` (String, Read-Only) Member name.
  - `status` (String, Read-Only) Member lifecycle status: `active`, `creating`, `pending_create`, `pending_update`, `pending_delete`, `deleting`, `disabled`, `error`, `unknown`.

### Read-Only

- `id` (String) The pool ID.
- `status` (String) Pool lifecycle status: `active`, `creating`, `pending_create`, `pending_update`, `pending_delete`, `deleting`, `disabled`, `error`, `unknown`.
- `created_at` (String) Creation timestamp.

## In-place updates

`name`, `description`, `lb_algorithm`, `session_persistence`, `tls_enabled`, and the entire `member` set support in-place updates.

`load_balancer_id`, `listener_id`, and `protocol` are `ForceNew` — to move the pool to a different load balancer or attach it to a different listener, destroy and recreate. To **change which pool is a listener's default** without recreating the pool, update `vnpaycloud_lb_listener.default_pool_id` instead.

## Timeouts

- `create` - (Default `10 minutes`)
- `update` - (Default `10 minutes`)
- `delete` - (Default `10 minutes`)

~> **Rate limit:** see [Rate limits](../index.md#rate-limits) — applies to all create/update/delete on this resource type.

## Import

```shell
terraform import vnpaycloud_lb_pool.example <pool-id>
```

Existing `member` set is restored from the live pool — to keep them managed declaratively, mirror each as a `member { ... }` block in your config. Without them, `terraform plan` will treat the live members as drift and try to remove them on the next apply.
