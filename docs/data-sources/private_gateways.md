---
page_title: "vnpaycloud_private_gateways Data Source - VNPayCloud"
subcategory: "Kubernetes"
description: |-
  Get the list of private gateways in VNPayCloud.
---

# vnpaycloud_private_gateways (Data Source)

Use this data source to get the list of all private gateways in the current project.

## Example Usage

```hcl
data "vnpaycloud_private_gateways" "all" {}

output "private_gateway_names" {
  value = [for g in data.vnpaycloud_private_gateways.all.private_gateways : g.name]
}
```

## Schema

### Read-Only

- `private_gateways` (List of Object) The list of private gateways. Each element has the attributes below.

### Nested Schema for `private_gateways`

- `id` (String) The ID of the private gateway.
- `name` (String) The name of the private gateway.
- `description` (String) A description of the private gateway.
- `load_balancer_id` (String) The ID of the internal load balancer provisioned for the private gateway.
- `subnet_id` (String) The ID of the subnet the private gateway is deployed into.
- `status` (String) The current status of the private gateway.
- `created_at` (String) The creation timestamp of the private gateway.
- `zone_id` (String) The ID of the zone the private gateway belongs to.
