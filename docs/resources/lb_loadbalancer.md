---
page_title: "vnpaycloud_lb_loadbalancer Resource - VNPayCloud"
subcategory: "Load Balancer"
description: |-
  Manages a load balancer within VNPayCloud.
---

# vnpaycloud_lb_loadbalancer (Resource)

Manages a load balancer within VNPayCloud. A load balancer distributes incoming network traffic across multiple backend servers to ensure high availability and reliability. It is created within a subnet and can optionally be associated with a floating IP for public access.

## Example Usage

### Internal load balancer

```hcl
resource "vnpaycloud_lb_loadbalancer" "internal" {
  name        = "internal-lb"
  description = "Internal load balancer for microservices"
  subnet_id   = "sb_iaas_portal_subnet_..."
  flavor      = "t1-small"
}
```

### External load balancer with floating IP

```hcl
resource "vnpaycloud_floating_ip" "lb_ip" {
  vpc_id = "vpc-abc"
}

resource "vnpaycloud_lb_loadbalancer" "external" {
  name           = "public-lb"
  subnet_id      = "sb_iaas_portal_subnet_..."
  flavor         = "t1-small"
  floating_ip_id = vnpaycloud_floating_ip.lb_ip.id
}

output "lb_public_ip" {
  value = vnpaycloud_floating_ip.lb_ip.address
}
```

## Schema

### Required

- `name` (String) The name of the load balancer. Length `3`–`250`, no leading/trailing whitespace. Unique per zone.
- `subnet_id` (String, ForceNew) The ID of the subnet where the load balancer's VIP will be allocated. The subnet must be in the same zone as the provider `zone_id`.
- `flavor` (String) The flavor name (e.g., `t1-small`, `t1-medium`, `t1-large`). The provider resolves this name to a flavor ID. Case-insensitive. Changing the flavor is applied **in-place** (the load balancer keeps its ID, VIP, listeners, and pools), but causes a brief data-plane interruption while the load balancer is rebuilt with the new flavor.

### Optional

- `description` (String) A human-readable description. Length `0`–`255`. Allowed characters: ASCII letters, digits, spaces, and `-` `_` `.` (must match `^[a-zA-Z0-9-_. ]*$`); other characters are rejected at create and update.
- `floating_ip_id` (String, Computed) The ID of a floating IP to associate with the load balancer for public access, applied **at create time only**. The FIP must exist in the same project and must not already be attached to a port. After the load balancer is created this argument is **read-only** — changing it has no effect and does not recreate the load balancer. To attach, detach, or switch the floating IP after creation, manage it from the `vnpaycloud_floating_ip` resource (`port_id = vnpaycloud_lb_loadbalancer.<name>.vip_port_id`). Omit it to create an internal-only load balancer.

### Read-Only

- `id` (String) The load balancer ID.
- `vip_address` (String) The virtual IP address assigned to the load balancer.
- `vip_port_id` (String) The network interface port backing the load balancer VIP.
- `vip_subnet_id` (String) The subnet ID associated with the VIP.
- `status` (String) Lifecycle status: `active`, `creating`, `pending_create`, `pending_update`, `pending_delete`, `deleting`, `disabled`, `error`, `unknown`.
- `created_at` (String) Creation timestamp.

## In-place updates

`name`, `description`, and `flavor` are updatable in place. Changing `flavor` triggers an in-place rebuild with a brief interruption — see the field note above. `subnet_id` is `ForceNew` and changing it recreates the resource. `floating_ip_id` is **create-time only** (see below).

## Floating IP behavior

`floating_ip_id` attaches a floating IP to the load balancer **at create time only** — the backend associates the FIP to the load balancer VIP port as part of provisioning an external load balancer. After the load balancer exists, this argument is read-only: changing it in the configuration has no effect (no drift, no recreate).

To attach, detach, or switch the floating IP **after** creation, manage the association from the `vnpaycloud_floating_ip` resource, pointing it at the load balancer VIP port:

```hcl
resource "vnpaycloud_lb_loadbalancer" "lb" {
  name      = "internal-lb"
  subnet_id = "sb_iaas_portal_subnet_..."
  flavor    = "t1-small"
}

# attach / switch / detach the public IP here — set port_id = "" to detach
resource "vnpaycloud_floating_ip" "lb_ip" {
  port_id = vnpaycloud_lb_loadbalancer.lb.vip_port_id
}
```

This keeps the floating IP association under a single owner (`vnpaycloud_floating_ip`), so the two resources never fight over it. The floating IP must be in the same project and must not already be attached to another port.

## Timeouts

- `create` - (Default `10 minutes`)
- `update` - (Default `10 minutes`)
- `delete` - (Default `10 minutes`)

~> **Rate limit:** see [Rate limits](../index.md#rate-limits) — applies to all create/update/delete on this resource type.

## Import

```shell
terraform import vnpaycloud_lb_loadbalancer.example <loadbalancer-id>
```

After import you only need to declare the **required** fields (`name`, `subnet_id`, `flavor`). Add `floating_ip_id` when you want Terraform to manage the load balancer public IP association.
