---
page_title: "vnpaycloud_lb_listener Data Source - VNPayCloud"
subcategory: "Load Balancer"
description: |-
  Get information about a load balancer listener in VNPayCloud.
---

# vnpaycloud_lb_listener (Data Source)

Use this data source to get information about an existing load balancer listener. A listener defines the protocol and port on which the load balancer accepts incoming connections.

## Example Usage

```hcl
data "vnpaycloud_lb_listener" "example" {
  id = "lst-bcd66778"
}

output "listener_protocol" {
  value = data.vnpaycloud_lb_listener.example.protocol
}

output "listener_default_pool" {
  value = data.vnpaycloud_lb_listener.example.default_pool_id
}
```

## Schema

### Required (filter)

- `id` (String) The ID of the load balancer listener.

### Read-Only

- `name` (String) The name of the listener.
- `load_balancer_id` (String) The ID of the load balancer this listener belongs to.
- `protocol` (String) The protocol the listener accepts (e.g., `HTTP`, `HTTPS`, `TCP`, `UDP`).
- `protocol_port` (Number) The port number on which the listener accepts connections (e.g., `80`, `443`).
- `default_pool_id` (String) The ID of the default backend pool to which traffic is forwarded.
- `status` (String) The current provisioning status of the listener (e.g., `ACTIVE`, `PENDING_CREATE`, `ERROR`).
- `created_at` (String) The timestamp when the listener was created, in ISO 8601 format.
