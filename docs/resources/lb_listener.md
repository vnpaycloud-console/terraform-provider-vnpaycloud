---
page_title: "vnpaycloud_lb_listener Resource - VNPayCloud"
subcategory: "Load Balancer"
description: |-
  Manages a load balancer listener within VNPayCloud.
---

# vnpaycloud_lb_listener (Resource)

Manages a listener for a load balancer within VNPayCloud. A listener defines the protocol and port on which the load balancer accepts incoming connections and forwards them to a backend pool.

## Example Usage

### HTTP listener

```hcl
resource "vnpaycloud_lb_loadbalancer" "app" {
  name      = "app-loadbalancer"
  subnet_id = "subnet-abc12345"
  flavor    = "lb.small"
}

resource "vnpaycloud_lb_listener" "http" {
  name             = "http-listener"
  load_balancer_id = vnpaycloud_lb_loadbalancer.app.id
  protocol         = "HTTP"
  protocol_port    = 80
}
```

### HTTPS listener

```hcl
resource "vnpaycloud_lb_listener" "https" {
  name             = "https-listener"
  load_balancer_id = vnpaycloud_lb_loadbalancer.app.id
  protocol         = "HTTPS"
  protocol_port    = 443
}
```

### TCP listener on custom port

```hcl
resource "vnpaycloud_lb_listener" "tcp_custom" {
  name             = "tcp-8080-listener"
  load_balancer_id = vnpaycloud_lb_loadbalancer.app.id
  protocol         = "TCP"
  protocol_port    = 8080
}
```

## Schema

### Required

- `name` (String) The name of the listener.
- `load_balancer_id` (String, ForceNew) The ID of the load balancer to attach this listener to. Changing this creates a new listener.
- `protocol` (String, ForceNew) The protocol the listener accepts. Valid values are `HTTP`, `HTTPS`, `TCP`, `UDP`. Changing this creates a new listener.
- `protocol_port` (Number, ForceNew) The port on which the listener accepts traffic. Must be between `1` and `65535`. Changing this creates a new listener.

### Optional

- `default_pool_id` (String, Computed) The ID of the default pool to route traffic to. This may be computed after associating a pool with this listener via `vnpaycloud_lb_pool`.

### Read-Only

- `id` (String) The ID of the listener.
- `status` (String) The current status of the listener (e.g., `ACTIVE`, `PENDING_CREATE`, `ERROR`).
- `created_at` (String) The creation timestamp of the listener in ISO 8601 format.

## Timeouts

- `create` - (Default `10 minutes`) Used for creating the listener.
- `delete` - (Default `10 minutes`) Used for deleting the listener.

## Import

Listeners can be imported using the `id`:

```shell
terraform import vnpaycloud_lb_listener.example <listener-id>
```
