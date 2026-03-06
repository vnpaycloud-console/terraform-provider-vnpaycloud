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

## Schema

### Optional

- `port_id` (String) The ID of the network interface (port) to associate this floating IP with. Conflicts with `vpc_id`.
- `vpc_id` (String) The ID of the VPC to associate this floating IP with (for VPC-level SNAT). Conflicts with `port_id`.

### Read-Only

- `id` (String) The ID of the floating IP.
- `address` (String) The public IP address allocated.
- `status` (String) The current status of the floating IP.
- `instance_id` (String) The ID of the server instance this floating IP is associated with (if any).
- `instance_name` (String) The name of the server instance this floating IP is associated with (if any).
- `created_at` (String) The creation timestamp of the floating IP.

## Timeouts

- `create` - (Default `10 minutes`) Used for creating the floating IP.
- `delete` - (Default `10 minutes`) Used for deleting the floating IP.

## Import

Floating IPs can be imported using the `id`:

```shell
terraform import vnpaycloud_floating_ip.example <floating-ip-id>
```
