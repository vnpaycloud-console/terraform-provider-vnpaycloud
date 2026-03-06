---
page_title: "vnpaycloud_network_interface Resource - VNPayCloud"
subcategory: "Networking"
description: |-
  Manages a network interface (port) resource within VNPayCloud.
---

# vnpaycloud_network_interface (Resource)

Manages a network interface (virtual NIC/port) resource within VNPayCloud. Network interfaces can be created independently and attached to server instances, enabling flexible network topology management.

~> **Note:** The `name`, `subnet_id`, `ip_address`, and `description` fields are all immutable. Changing them will force creation of a new network interface.

## Example Usage

```hcl
resource "vnpaycloud_vpc" "main" {
  name = "my-vpc"
  cidr = "10.0.0.0/16"
}

resource "vnpaycloud_subnet" "app" {
  name   = "app-subnet"
  vpc_id = vnpaycloud_vpc.main.id
  cidr   = "10.0.1.0/24"
}

resource "vnpaycloud_network_interface" "example" {
  name        = "my-network-interface"
  subnet_id   = vnpaycloud_subnet.app.id
  ip_address  = "10.0.1.10"
  description = "Primary NIC for app server"
}
```

### With Dynamic IP Assignment

```hcl
resource "vnpaycloud_network_interface" "dynamic" {
  name      = "dynamic-nic"
  subnet_id = vnpaycloud_subnet.app.id
}
```

## Schema

### Required

- `name` (String, ForceNew) The name of the network interface. Changing this creates a new network interface.
- `subnet_id` (String, ForceNew) The ID of the subnet in which to create the network interface. Changing this creates a new network interface.

### Optional

- `ip_address` (String, ForceNew, Computed) The IP address to assign to the network interface. If not specified, an IP address is automatically assigned from the subnet. Changing this creates a new network interface.
- `description` (String, ForceNew) A description of the network interface. Changing this creates a new network interface.

### Read-Only

- `id` (String) The ID of the network interface.
- `network_id` (String) The ID of the network associated with the subnet.
- `mac_address` (String) The MAC address of the network interface.
- `status` (String) The current status of the network interface.
- `security_groups` (List of String) The list of security group IDs associated with the network interface.
- `port_security_enabled` (Boolean) Whether port security is enabled on the network interface.
- `network_type` (String) The type of the underlying network.
- `created_at` (String) The creation timestamp of the network interface.

## Timeouts

- `create` - (Default `10 minutes`) Used for creating the network interface.
- `delete` - (Default `10 minutes`) Used for deleting the network interface.

## Import

Network interfaces can be imported using the `id`:

```shell
terraform import vnpaycloud_network_interface.example <network-interface-id>
```
