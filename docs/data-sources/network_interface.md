---
page_title: "vnpaycloud_network_interface Data Source - VNPayCloud"
subcategory: "Networking"
description: |-
  Get information about a network interface in VNPayCloud.
---

# vnpaycloud_network_interface (Data Source)

Use this data source to get information about an existing network interface (port), including its IP address, MAC address, and security group associations.

## Example Usage

```hcl
data "vnpaycloud_network_interface" "example" {
  name = "my-network-interface"
}

output "interface_ip" {
  value = data.vnpaycloud_network_interface.example.ip_address
}
```

```hcl
data "vnpaycloud_network_interface" "by_id" {
  id = "nic-jkl33445"
}
```

## Schema

### Optional (filter)

- `id` (String) The ID of the network interface.
- `name` (String) The name of the network interface.

### Read-Only

- `network_id` (String) The ID of the network this interface is attached to.
- `subnet_id` (String) The ID of the subnet this interface belongs to.
- `ip_address` (String) The primary fixed IP address assigned to this interface.
- `mac_address` (String) The MAC address of the network interface.
- `security_groups` (List of String) A list of security group IDs associated with this interface.
- `port_security_enabled` (Boolean) Whether port security (anti-spoofing) is enabled on this interface.
- `allowed_address_pairs` (List of Object) A list of allowed address pairs for this interface. Each object contains:
  - `ip_address` (String) The allowed IP address or CIDR.
  - `mac_address` (String) The allowed MAC address (optional).
- `network_type` (String) The type of network this interface is connected to (e.g., `vxlan`, `flat`).
- `status` (String) The current status of the network interface (e.g., `ACTIVE`, `DOWN`, `BUILD`).
- `created_at` (String) The timestamp when the network interface was created, in ISO 8601 format.
