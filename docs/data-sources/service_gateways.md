---
page_title: "vnpaycloud_service_gateways Data Source - VNPayCloud"
subcategory: "Networking"
description: |-
  List service gateways in VNPayCloud.
---

# vnpaycloud_service_gateways (Data Source)

Use this data source to list all [service gateways](../resources/service_gateway.md) in the project's zone.

## Example Usage

```hcl
data "vnpaycloud_service_gateways" "all" {}

output "gateway_names" {
  value = [for g in data.vnpaycloud_service_gateways.all.service_gateways : g.name]
}
```

## Schema

### Read-Only

- `service_gateways` (List of Object) The list of service gateways. Each object has the same read-only attributes as the [`vnpaycloud_service_gateway`](service_gateway.md) data source (`id`, `name`, `description`, `subnet_id`, `vpc_id`, `flavor_id`, `allowed_icmp`, `vip_address`, `load_balancer_id`, `port_id`, `operating_status`, `provisioning_status`, `status`, `created_at`).
