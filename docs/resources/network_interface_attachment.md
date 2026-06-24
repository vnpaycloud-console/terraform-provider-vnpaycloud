---
page_title: "vnpaycloud_network_interface_attachment Resource - VNPayCloud"
subcategory: "Networking"
description: |-
  Attaches a network interface to a server instance within VNPayCloud.
---

# vnpaycloud_network_interface_attachment (Resource)

Attaches a network interface (port) to a server instance within VNPayCloud. This resource manages the attachment lifecycle — creating it attaches the interface to the server, destroying it detaches it.

~> **Note:** This resource does not support import. Both `network_interface_id` and `server_id` are immutable; changing either will force creation of a new attachment.

~> **Note:** The attached interface must belong to a **different subnet** than the server's existing interface(s). Attaching a second interface from the same subnet is rejected by the backend with a conflict error. (The example below places the extra NIC in a separate subnet.)

## Example Usage

```hcl
resource "vnpaycloud_vpc" "main" {
  name = "my-vpc"
  cidr = "10.0.0.0/16"
}

resource "vnpaycloud_subnet" "primary" {
  name   = "primary-subnet"
  vpc_id = vnpaycloud_vpc.main.id
  cidr   = "10.0.1.0/24"
}

resource "vnpaycloud_subnet" "extra" {
  name   = "extra-subnet"
  vpc_id = vnpaycloud_vpc.main.id
  cidr   = "10.0.2.0/24"
}

resource "vnpaycloud_network_interface" "primary" {
  name      = "primary-nic"
  subnet_id = vnpaycloud_subnet.primary.id
}

resource "vnpaycloud_instance" "app" {
  name                  = "app-server"
  image                 = "Ubuntu 22.04 LTS"
  flavor                = "a-pro-small.2x2"
  root_disk_gb          = 20
  root_disk_type        = "c1-standard"
  network_interface_ids = [vnpaycloud_network_interface.primary.id]

  # Interfaces added via vnpaycloud_network_interface_attachment also appear in
  # network_interface_ids; ignore them so the two resources don't fight — see
  # the note below.
  lifecycle {
    ignore_changes = [network_interface_ids]
  }
}

resource "vnpaycloud_network_interface" "extra" {
  name      = "extra-nic"
  subnet_id = vnpaycloud_subnet.extra.id
}

resource "vnpaycloud_network_interface_attachment" "example" {
  network_interface_id = vnpaycloud_network_interface.extra.id
  server_id            = vnpaycloud_instance.app.id
}
```

~> **Note:** An interface attached with this resource also shows up in the instance's `network_interface_ids`. Manage a given interface with **either** `network_interface_ids` on `vnpaycloud_instance` **or** this resource — not both; if you use both, add `lifecycle { ignore_changes = [network_interface_ids] }` to the instance.

## Schema

### Required

- `network_interface_id` (String, ForceNew) The ID of the network interface to attach. Changing this creates a new attachment.
- `server_id` (String, ForceNew) The ID of the server instance to which the network interface is attached. Changing this creates a new attachment.

### Read-Only

- `status` (String) The current status of the network interface after attachment (for example, `active`).
- `ip_address` (String) The IP address assigned to the network interface after attachment.

## Timeouts

- `create` - (Default `10 minutes`) Used for attaching the network interface.
- `delete` - (Default `10 minutes`) Used for detaching the network interface.

## Import

Network interface attachments do not support import.
