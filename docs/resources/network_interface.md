---
page_title: "vnpaycloud_network_interface Resource - VNPayCloud"
subcategory: "Networking"
description: |-
  Manages a network interface (port) resource within VNPayCloud.
---

# vnpaycloud_network_interface (Resource)

Manages a network interface (virtual NIC/port) resource within VNPayCloud. Network interfaces can be created independently and attached to server instances, enabling flexible network topology management.

~> **Note:** The `name`, `subnet_id`, and `ip_address` fields are immutable — changing them forces creation of a new network interface. `description`, `reserved`, `virtual_ip`, `allowed_address_pairs`, `port_security_enabled`, and `security_groups` can be updated in place. `name` may be an empty string when you want to create an unnamed interface.

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
  ip_address  = "10.0.1.20"
  description = "Primary NIC for app server"
}
```

~> **Note:** When `ip_address` is set explicitly, the host portion of the address must be in the range `[16, 250]` (the first 15 and the last few host addresses of each subnet are reserved). For example, in a `/24` subnet, `.16`–`.250` are assignable; `.1`–`.15`, `.251`–`.255` are rejected. Leave `ip_address` unset to let the platform auto-assign a valid address.

~> **Note:** `description` may contain only letters, digits, spaces, hyphens (`-`), underscores (`_`), and periods (`.`). Other punctuation, such as `:`, is rejected by the backend.

~> **Note:** A reserved interface (`reserved = true`) cannot be deleted by the backend. Before running `terraform destroy` (or otherwise removing the resource), set `reserved = false` and `terraform apply` first; attempting to delete a reserved interface fails with `This is a reserved port. You cannot delete.`

~> **Note:** On create, when `port_security_enabled` is left enabled (its default) and `security_groups` is omitted, the interface is automatically assigned — and keeps — the project's **default security group** and the **system security group**; Terraform does not manage the list in this case. Do not set `security_groups = []`; an explicit empty set is not supported and is rejected at plan time. If `security_groups` is set, `port_security_enabled` must be enabled and the system security group must remain in the list. Setting `port_security_enabled = false` clears the security groups, so it cannot be combined with `security_groups`.

### With Dynamic IP Assignment

```hcl
resource "vnpaycloud_network_interface" "dynamic" {
  name      = "dynamic-nic"
  subnet_id = vnpaycloud_subnet.app.id
}
```

### With Empty Name and Dynamic IP

```hcl
resource "vnpaycloud_network_interface" "unnamed" {
  name      = ""
  subnet_id = vnpaycloud_subnet.app.id
}
```

### Reserved IP and Virtual IP

```hcl
resource "vnpaycloud_network_interface" "vip" {
  name       = "vip-nic"
  subnet_id  = vnpaycloud_subnet.app.id
  reserved   = true
  virtual_ip = true
}
```

### Allowed Address Pairs

```hcl
resource "vnpaycloud_network_interface" "ha" {
  name      = "ha-nic"
  subnet_id = vnpaycloud_subnet.app.id

  allowed_address_pairs {
    ip_address = "10.0.1.100"
  }
  allowed_address_pairs {
    ip_address = "10.0.2.0/24"
  }
}
```

### With Security Groups

When `security_groups` is set, the system security group must stay in the list.
Look it up with the `vnpaycloud_security_group` data source and include its ID
alongside your own groups:

```hcl
data "vnpaycloud_security_group" "system" {
  name = "System Security Group"
}

resource "vnpaycloud_security_group" "web" {
  name = "web-sg"
}

resource "vnpaycloud_network_interface" "with_sg" {
  name      = "sg-nic"
  subnet_id = vnpaycloud_subnet.app.id

  security_groups = [
    data.vnpaycloud_security_group.system.id,
    vnpaycloud_security_group.web.id,
  ]
}
```

## Schema

### Required

- `name` (String, ForceNew) The name of the network interface. May be an empty string. Changing this creates a new network interface.
- `subnet_id` (String, ForceNew) The ID of the subnet in which to create the network interface. Changing this creates a new network interface.

### Optional

- `ip_address` (String, ForceNew, Computed) The IP address to assign to the network interface. If set, the backend validates that it is a valid IP within the subnet CIDR with host ID in range `[16, 250]` (see note above). If not specified, an IP address is automatically assigned from the subnet. Changing this creates a new network interface.
- `description` (String) A description of the network interface. It may contain only letters, digits, spaces, hyphens (`-`), underscores (`_`), and periods (`.`). Can be updated in place.
- `reserved` (Boolean, Computed) Whether the interface (its IP) is reserved. Can be set at create and updated in place. A reserved interface cannot be deleted — set `reserved = false` and apply before destroying it (see note above).
- `virtual_ip` (Boolean, Computed) Whether the interface is marked as a virtual IP (VIP). Can be set at create and updated in place.
- `allowed_address_pairs` (Block List, Computed) Additional IP address (or CIDR) / MAC pairs allowed to pass through this interface — used for VIP/HA setups. Can be set at create and updated in place. Each block supports:
  - `ip_address` (String, Required) An IP address or CIDR allowed on the interface.
  - `mac_address` (String, Optional, Computed) The MAC address for the pair. Defaults to the interface's own MAC if omitted.
- `security_groups` (Set of String, Computed) The set of security group IDs associated with the interface. If omitted, the platform assigns and keeps the default and system security groups. An explicit empty set (`security_groups = []`) is not supported. If set, Terraform manages the list and the system security group must remain attached. Requires `port_security_enabled` to be enabled; cannot be set together with `port_security_enabled = false`. Can be set at create and updated in place.
- `port_security_enabled` (Boolean, Computed) Whether port security (anti-spoof) is enabled on the interface. When set to `false`, `security_groups` must be omitted. Can be set at create and updated in place.

### Read-Only

- `id` (String) The ID of the network interface.
- `network_id` (String) The ID of the network associated with the subnet.
- `mac_address` (String) The MAC address of the network interface.
- `status` (String) The current status of the network interface.
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
