---
page_title: "vnpaycloud_floating_ip Resource - VNPayCloud"
subcategory: "Networking"
description: |-
  Manages a floating IP resource within VNPayCloud.
---

# vnpaycloud_floating_ip (Resource)

Manages a floating IP (public IP) resource within VNPayCloud. Floating IPs can be associated with a server port or a VPC for SNAT purposes.

## Example Usage

### Floating IP Associated with a Server Port

```hcl
resource "vnpaycloud_floating_ip" "example" {
  port_id = vnpaycloud_network_interface.main.id
}
```

~> **Note:** Associating a floating IP with a server port requires the port's VPC to have an Internet Gateway attached (the floating IP routes out through it). Without one, the association is rejected with `VPC is not associated with an Internet Gateway`. Create a `vnpaycloud_internet_gateway` for the VPC first (and depend on it). The same precondition applies to VPC-level SNAT association.

### Floating IP Associated with a VPC

```hcl
resource "vnpaycloud_floating_ip" "vpc_snat" {
  vpc_id = vnpaycloud_vpc.main.id
}
```

### Unassociated Floating IP

```hcl
resource "vnpaycloud_floating_ip" "spare" {}
```

### Explicitly Disassociate a Floating IP

`port_id` and `vpc_id` are computed so backend-managed attachments, such as a load balancer using `floating_ip_id`, do not create Terraform drift. Because of that, omitting the field or setting it to `null` keeps the current backend value.

To actively detach a floating IP managed by this resource, set the currently managed association field to an empty string:

```hcl
resource "vnpaycloud_floating_ip" "vpc_snat" {
  vpc_id = ""
}
```

```hcl
resource "vnpaycloud_floating_ip" "server_ip" {
  port_id = ""
}
```

~> **Load balancer public IPs are managed here.** `vnpaycloud_lb_loadbalancer.floating_ip_id` only attaches a floating IP **at create time** (it is read-only afterward). To attach, switch, or detach a load balancer's public IP **after** creation, set `port_id = vnpaycloud_lb_loadbalancer.<name>.vip_port_id` on this resource (and `port_id = ""` to detach). This keeps the association under a single owner — the floating IP resource — so nothing drifts.

## Schema

### Optional

- `port_id` (String, Computed) The ID of the network interface (port) to associate this floating IP with. Conflicts with `vpc_id`. **Computed because external resources may set it server-side** — e.g. attaching this FIP to a load balancer via `vnpaycloud_lb_loadbalancer.floating_ip_id` causes the backend to set `port_id` to the LB's port. Omit it from HCL when the attachment is managed by another resource; only declare it when you want to actively manage the attachment from this `vnpaycloud_floating_ip`. Set `port_id = ""` to explicitly disassociate a port-managed attachment.
- `vpc_id` (String, Computed) The ID of the VPC to associate this floating IP with (for VPC-level SNAT). Conflicts with `port_id`. Computed for the same reason as `port_id`. Set `vpc_id = ""` to explicitly disassociate a VPC-managed attachment.

### Read-Only

- `id` (String) The ID of the floating IP.
- `address` (String) The public IP address allocated.
- `fixed_ip` (String) The private (fixed) IP address that this floating IP maps to when associated.
- `status` (String) The current status of the floating IP.
- `instance_id` (String) The ID of the server instance this floating IP is associated with (if any).
- `instance_name` (String) The name of the server instance this floating IP is associated with (if any).
- `created_at` (String) The creation timestamp of the floating IP.

## Timeouts

- `create` - (Default `10 minutes`) Used for creating the floating IP.
- `update` - (Default `10 minutes`) Used for associating or disassociating the floating IP.
- `delete` - (Default `10 minutes`) Used for deleting the floating IP.

## Import

Floating IPs can be imported using the `id`:

```shell
terraform import vnpaycloud_floating_ip.example <floating-ip-id>
```
