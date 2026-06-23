---
page_title: "vnpaycloud_service_endpoints Data Source - VNPayCloud"
subcategory: "Networking"
description: |-
  List service endpoints in VNPayCloud.
---

# vnpaycloud_service_endpoints (Data Source)

Use this data source to list [service endpoints](../resources/service_endpoint.md) in the project's zone, optionally filtered by service gateway.

## Example Usage

```hcl
data "vnpaycloud_service_endpoints" "on_gateway" {
  service_gateway_id = vnpaycloud_service_gateway.example.id
}

output "endpoint_ports" {
  value = [for e in data.vnpaycloud_service_endpoints.on_gateway.service_endpoints : e.port]
}
```

## Schema

### Optional (filter)

- `service_gateway_id` (String) Only return endpoints on this service gateway.

### Read-Only

- `service_endpoints` (List of Object) The list of service endpoints. Each object has the same read-only attributes as the [`vnpaycloud_service_endpoint`](service_endpoint.md) data source (`id`, `name`, `description`, `provider_id`, `service_id`, `service_gateway_id`, `port`, `allowed_cidrs`, `listener_id`, `pool_id`, `health_monitor_id`, `pool_member_ids`, `operating_status`, `provisioning_status`, `status`, `created_at`).
