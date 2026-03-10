---
page_title: "vnpaycloud_lb_loadbalancer Resource - VNPayCloud"
subcategory: "Load Balancer"
description: |-
  Manages a load balancer within VNPayCloud.
---

# vnpaycloud_lb_loadbalancer (Resource)

Manages a load balancer within VNPayCloud. A load balancer distributes incoming network traffic across multiple backend servers to ensure high availability and reliability. It is created within a subnet and can optionally be associated with a floating IP for public access.

## Example Usage

```hcl
resource "vnpaycloud_lb_loadbalancer" "app" {
  name           = "app-loadbalancer"
  subnet_id      = "subnet-abc12345"
  flavor         = "lb.small"
  description    = "Load balancer for the application tier"
  floating_ip_id = "fip-xyz98765"
}

output "lb_vip" {
  value = vnpaycloud_lb_loadbalancer.app.vip_address
}
```

### Internal load balancer (no floating IP)

```hcl
resource "vnpaycloud_lb_loadbalancer" "internal" {
  name        = "internal-lb"
  subnet_id   = "subnet-internal123"
  flavor      = "lb.medium"
  description = "Internal load balancer for microservices"
}
```

## Schema

### Required

- `name` (String) The name of the load balancer.
- `subnet_id` (String, ForceNew) The ID of the subnet where the load balancer's virtual IP will be allocated. Changing this creates a new load balancer.
- `flavor` (String, ForceNew) The flavor of the load balancer, defining its capacity and performance tier (e.g., `lb.small`, `lb.medium`, `lb.large`). Changing this creates a new load balancer.

### Optional

- `description` (String) A human-readable description of the load balancer.
- `floating_ip_id` (String, ForceNew) The ID of the floating IP to associate with the load balancer for public access. Changing this creates a new load balancer.

### Read-Only

- `id` (String) The ID of the load balancer.
- `vip_address` (String) The virtual IP address assigned to the load balancer within the subnet.
- `vip_subnet_id` (String) The subnet ID associated with the VIP address.
- `status` (String) The current status of the load balancer (e.g., `ACTIVE`, `PENDING_CREATE`, `ERROR`).
- `listener_ids` (List of String) List of listener IDs attached to this load balancer.
- `created_at` (String) The creation timestamp of the load balancer in ISO 8601 format.

## Timeouts

- `create` - (Default `10 minutes`) Used for creating the load balancer.
- `update` - (Default `10 minutes`) Used for updating the load balancer.
- `delete` - (Default `10 minutes`) Used for deleting the load balancer.

## Import

Load balancers can be imported using the `id`:

```shell
terraform import vnpaycloud_lb_loadbalancer.example <loadbalancer-id>
```
