---
page_title: "vnpaycloud_floating_ips Data Source - VNPayCloud"
subcategory: "Networking"
description: |-
  List all floating IPs in VNPayCloud.
---

# vnpaycloud_floating_ips (Data Source)

Use this data source to list all floating IPs in the current project.

## Example Usage

```hcl
data "vnpaycloud_floating_ips" "all" {}

output "all_floating_addresses" {
  value = data.vnpaycloud_floating_ips.all.floating_ips[*].address
}

output "unassigned_floating_ip_ids" {
  value = [
    for fip in data.vnpaycloud_floating_ips.all.floating_ips :
    fip.id if fip.instance_id == ""
  ]
}
```

## Schema

### Read-Only

- `floating_ips` (List of Object) List of floating IPs. Each element contains:
  - `id` (String) The unique identifier of the floating IP.
  - `address` (String) The public floating IP address.
  - `status` (String) The current status of the floating IP (e.g., `ACTIVE`, `DOWN`, `ERROR`).
  - `port_id` (String) The ID of the port this floating IP is associated with, if any.
  - `instance_id` (String) The ID of the instance this floating IP is attached to, if any.
  - `instance_name` (String) The name of the instance this floating IP is attached to, if any.
  - `created_at` (String) The timestamp when the floating IP was created, in ISO 8601 format.
