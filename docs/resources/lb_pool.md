---
page_title: "vnpaycloud_lb_pool Resource - VNPayCloud"
subcategory: "Load Balancer"
description: |-
  Manages a load balancer pool within VNPayCloud.
---

# vnpaycloud_lb_pool (Resource)

Manages a backend pool for a load balancer listener within VNPayCloud. A pool groups together a set of backend members (servers) and defines the algorithm used to distribute traffic among them.

## Example Usage

### Round-robin HTTP pool with members

```hcl
resource "vnpaycloud_lb_listener" "http" {
  name             = "http-listener"
  load_balancer_id = "lb-abc12345"
  protocol         = "HTTP"
  protocol_port    = 80
}

resource "vnpaycloud_lb_pool" "app_pool" {
  name         = "app-backend-pool"
  listener_id  = vnpaycloud_lb_listener.http.id
  lb_algorithm = "ROUND_ROBIN"
  protocol     = "HTTP"

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

### Least-connections TCP pool

```hcl
resource "vnpaycloud_lb_pool" "tcp_pool" {
  name         = "tcp-backend-pool"
  listener_id  = "listener-xyz98765"
  lb_algorithm = "LEAST_CONNECTIONS"
  protocol     = "TCP"

  member {
    address       = "192.168.1.100"
    protocol_port = 3000
  }
}
```

## Schema

### Required

- `name` (String) The name of the pool.
- `listener_id` (String, ForceNew) The ID of the listener to associate this pool with. Changing this creates a new pool.
- `lb_algorithm` (String) The load balancing algorithm used to distribute traffic across pool members. Valid values are `ROUND_ROBIN`, `LEAST_CONNECTIONS`, `SOURCE_IP`.
- `protocol` (String, ForceNew) The protocol used for communication between the load balancer and the pool members. Valid values are `HTTP`, `HTTPS`, `TCP`, `UDP`, `PROXY`. Changing this creates a new pool.

### Optional

- `member` (Block List) A list of backend members to include in the pool. Each member block supports the following:
  - `address` (String, Required) The IP address of the backend member.
  - `protocol_port` (Number, Required) The port on the backend member that receives traffic. Must be between `1` and `65535`.
  - `weight` (Number, Optional) The relative weight of the member for load distribution. Must be between `0` and `256`. A weight of `0` prevents new connections from being sent to this member. Defaults to `1`.
  - `id` (String, Read-Only) The ID of the pool member.
  - `name` (String, Read-Only) The system-assigned name of the pool member.
  - `status` (String, Read-Only) The current status of the pool member (e.g., `ACTIVE`, `ERROR`, `NO_MONITOR`).

### Read-Only

- `id` (String) The ID of the pool.
- `status` (String) The current status of the pool (e.g., `ACTIVE`, `PENDING_CREATE`, `ERROR`).
- `created_at` (String) The creation timestamp of the pool in ISO 8601 format.

## Timeouts

- `create` - (Default `10 minutes`) Used for creating the pool.
- `delete` - (Default `10 minutes`) Used for deleting the pool.

## Import

Pools can be imported using the `id`:

```shell
terraform import vnpaycloud_lb_pool.example <pool-id>
```
