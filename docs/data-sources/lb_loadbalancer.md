---
page_title: "vnpaycloud_lb_loadbalancer Data Source - VNPayCloud"
subcategory: "Load Balancer"
description: |-
  Get information about a load balancer in VNPayCloud.
---

# vnpaycloud_lb_loadbalancer (Data Source)

Use this data source to get information about an existing load balancer, including its virtual IP address and lifecycle status.

## Example Usage

```hcl
data "vnpaycloud_lb_loadbalancer" "example" {
  id = "lb-yza44556"
}

output "lb_vip_address" {
  value = data.vnpaycloud_lb_loadbalancer.example.vip_address
}
```

## Schema

### Required (filter)

- `id` (String) The ID of the load balancer.

### Read-Only

- `name` (String) The name of the load balancer.
- `description` (String) A human-readable description of the load balancer.
- `vip_address` (String) The virtual IP address (VIP) of the load balancer.
- `vip_subnet_id` (String) The ID of the subnet in which the virtual IP is allocated.
- `status` (String) Lifecycle status: `active`, `creating`, `pending_create`, `pending_update`, `pending_delete`, `deleting`, `disabled`, `error`, `unknown`.
- `created_at` (String) The timestamp when the load balancer was created, in ISO 8601 format.
- `floating_ip_id` (String) The ID of the floating IP attached to the load balancer. Returned as an empty string (`""`) when the LB is internal-only (no FIP attached).
