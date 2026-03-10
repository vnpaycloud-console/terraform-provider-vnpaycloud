---
page_title: "vnpaycloud_lb_loadbalancer Data Source - VNPayCloud"
subcategory: "Load Balancer"
description: |-
  Get information about a load balancer in VNPayCloud.
---

# vnpaycloud_lb_loadbalancer (Data Source)

Use this data source to get information about an existing load balancer, including its virtual IP address and associated listeners.

## Example Usage

```hcl
data "vnpaycloud_lb_loadbalancer" "example" {
  id = "lb-yza44556"
}

output "lb_vip_address" {
  value = data.vnpaycloud_lb_loadbalancer.example.vip_address
}

output "lb_listener_ids" {
  value = data.vnpaycloud_lb_loadbalancer.example.listener_ids
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
- `status` (String) The current provisioning status of the load balancer (e.g., `ACTIVE`, `PENDING_CREATE`, `ERROR`).
- `listener_ids` (List of String) A list of listener IDs attached to this load balancer.
- `created_at` (String) The timestamp when the load balancer was created, in ISO 8601 format.
