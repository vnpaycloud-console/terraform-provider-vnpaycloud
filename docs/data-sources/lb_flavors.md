---
page_title: "vnpaycloud_lb_flavors Data Source - VNPayCloud"
subcategory: "Load Balancer"
description: |-
  Returns the catalogue of load balancer flavors available in the project's zone.
---

# vnpaycloud_lb_flavors (Data Source)

Returns the full catalogue of load balancer flavors available for use when creating a [`vnpaycloud_lb_loadbalancer`](../resources/lb_loadbalancer.md). Each entry is the `name` and `id` the backend recognises, plus the zone the flavor belongs to.

Use it to:

- Discover valid `flavor` values without leaving Terraform.
- Pick a flavor dynamically (e.g. by name) instead of hard-coding it.

## Example Usage

### Print every available flavor

```hcl
data "vnpaycloud_lb_flavors" "all" {}

output "lb_flavor_names" {
  value = [for f in data.vnpaycloud_lb_flavors.all.flavors : f.name]
}
```

### Pick a flavor by name

```hcl
data "vnpaycloud_lb_flavors" "all" {}

locals {
  small_flavor = one([for f in data.vnpaycloud_lb_flavors.all.flavors : f if f.name == "t1-small"])
}

resource "vnpaycloud_lb_loadbalancer" "app" {
  name      = "app-lb"
  subnet_id = var.subnet_id
  flavor    = local.small_flavor.name
}
```

## Schema

### Read-Only

- `flavors` (List of Object) The catalogue of LB flavors.
  - `id` (String) Backend flavor ID.
  - `name` (String) Flavor name — pass this value to `vnpaycloud_lb_loadbalancer.flavor`.
  - `description` (String) Free-form description.
  - `zone_id` (String) Zone the flavor belongs to.
