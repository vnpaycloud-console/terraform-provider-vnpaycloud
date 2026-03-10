---
page_title: "vnpaycloud_lb_loadbalancers Data Source - VNPayCloud"
subcategory: "Load Balancer"
description: |-
  List all load balancers in VNPayCloud.
---

# vnpaycloud_lb_loadbalancers (Data Source)

Use this data source to list all load balancers in the current project.

## Example Usage

```hcl
data "vnpaycloud_lb_loadbalancers" "all" {}

output "all_lb_names" {
  value = data.vnpaycloud_lb_loadbalancers.all.load_balancers[*].name
}

output "active_lb_vip_addresses" {
  value = [
    for lb in data.vnpaycloud_lb_loadbalancers.all.load_balancers :
    lb.vip_address if lb.status == "ACTIVE"
  ]
}

output "lb_listener_map" {
  value = {
    for lb in data.vnpaycloud_lb_loadbalancers.all.load_balancers :
    lb.name => lb.listener_ids
  }
}
```

## Schema

### Read-Only

- `load_balancers` (List of Object) List of load balancers. Each element contains:
  - `id` (String) The unique identifier of the load balancer.
  - `name` (String) The name of the load balancer.
  - `description` (String) A human-readable description of the load balancer.
  - `vip_address` (String) The virtual IP address of the load balancer.
  - `vip_subnet_id` (String) The ID of the subnet where the virtual IP resides.
  - `status` (String) The current status of the load balancer (e.g., `ACTIVE`, `PENDING_CREATE`, `ERROR`).
  - `listener_ids` (List of String) A list of listener IDs attached to this load balancer.
  - `created_at` (String) The timestamp when the load balancer was created, in ISO 8601 format.
