---
page_title: "vnpaycloud_floating_ip Data Source - VNPayCloud"
subcategory: "Networking"
description: |-
  Get information about a floating IP in VNPayCloud.
---

# vnpaycloud_floating_ip (Data Source)

Use this data source to get information about an existing floating IP address, including its current association with a port or instance.

## Example Usage

```hcl
data "vnpaycloud_floating_ip" "example" {
  address = "203.0.113.42"
}

output "floating_ip_status" {
  value = data.vnpaycloud_floating_ip.example.status
}
```

```hcl
data "vnpaycloud_floating_ip" "by_id" {
  id = "fip-ghi11223"
}
```

## Schema

### Optional (filter)

- `id` (String) The ID of the floating IP.
- `address` (String) The public floating IP address.

### Read-Only

- `status` (String) The current status of the floating IP (e.g., `ACTIVE`, `DOWN`).
- `port_id` (String) The ID of the port to which this floating IP is currently associated.
- `instance_id` (String) The ID of the instance associated with this floating IP, if any.
- `instance_name` (String) The name of the instance associated with this floating IP, if any.
- `created_at` (String) The timestamp when the floating IP was allocated, in ISO 8601 format.
