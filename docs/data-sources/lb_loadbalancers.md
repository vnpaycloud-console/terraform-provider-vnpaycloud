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
    lb.vip_address if lb.status == "active"
  ]
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
  - `status` (String) Lifecycle status: `active`, `creating`, `pending_create`, `pending_update`, `pending_delete`, `deleting`, `disabled`, `error`, `unknown`.
  - `created_at` (String) The timestamp when the load balancer was created, in ISO 8601 format.
  - `floating_ip_id` (String) The ID of the floating IP attached to the load balancer. Returned as an empty string (`""`) when the LB is internal-only (no FIP attached).
