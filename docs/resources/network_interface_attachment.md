---
page_title: "vnpaycloud_network_interface_attachment Resource - VNPayCloud"
subcategory: "Networking"
description: |-
  Attaches a network interface to a server instance within VNPayCloud.
---

# vnpaycloud_network_interface_attachment (Resource)

Attaches a network interface (port) to a server instance within VNPayCloud. This resource manages the attachment lifecycle — creating it attaches the interface to the server, destroying it detaches it.

~> **Note:** This resource does not support import. Both `network_interface_id` and `server_id` are immutable; changing either will force creation of a new attachment.

## Example Usage

```hcl
resource "vnpaycloud_subnet" "app" {
  name   = "app-subnet"
  vpc_id = vnpaycloud_vpc.main.id
  cidr   = "10.0.1.0/24"
}

resource "vnpaycloud_network_interface" "extra" {
  name      = "extra-nic"
  subnet_id = vnpaycloud_subnet.app.id
}

resource "vnpaycloud_network_interface_attachment" "example" {
  network_interface_id = vnpaycloud_network_interface.extra.id
  server_id            = vnpaycloud_server.app.id
}
```

## Schema

### Required

- `network_interface_id` (String, ForceNew) The ID of the network interface to attach. Changing this creates a new attachment.
- `server_id` (String, ForceNew) The ID of the server instance to which the network interface is attached. Changing this creates a new attachment.

### Read-Only

- `status` (String) The current attachment status of the network interface.
- `ip_address` (String) The IP address assigned to the network interface after attachment.

## Timeouts

- `create` - (Default `10 minutes`) Used for attaching the network interface.
- `delete` - (Default `10 minutes`) Used for detaching the network interface.
