---
page_title: "vnpaycloud_private_gateway Data Source - VNPayCloud"
subcategory: "Kubernetes"
description: |-
  Get information about a private gateway in VNPayCloud.
---

# vnpaycloud_private_gateway (Data Source)

Use this data source to get information about an existing private gateway. Look it up by `id`, or by `name` within the current project. Prefer `id` for deterministic lookup.

## Example Usage

```hcl
data "vnpaycloud_private_gateway" "by_name" {
  name = "my-private-gateway"
}

output "gateway_load_balancer_id" {
  value = data.vnpaycloud_private_gateway.by_name.load_balancer_id
}
```

## Schema

### Optional

- `id` (String) The ID of the private gateway to look up. One of `id` or `name` must be set.
- `name` (String) The name of the private gateway to look up. One of `id` or `name` must be set. When both are set, `id` takes precedence.

### Read-Only

- `description` (String) A description of the private gateway.
- `load_balancer_id` (String) The ID of the internal load balancer provisioned for the private gateway.
- `subnet_id` (String) The ID of the subnet the private gateway is deployed into.
- `status` (String) The current status of the private gateway.
- `created_at` (String) The creation timestamp of the private gateway.
- `zone_id` (String) The ID of the zone the private gateway belongs to.
