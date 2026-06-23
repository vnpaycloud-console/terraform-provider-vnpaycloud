---
page_title: "vnpaycloud_service_endpoint Data Source - VNPayCloud"
subcategory: "Networking"
description: |-
  Get information about a service endpoint in VNPayCloud.
---

# vnpaycloud_service_endpoint (Data Source)

Use this data source to get information about an existing [service endpoint](../resources/service_endpoint.md) by its ID.

## Example Usage

```hcl
data "vnpaycloud_service_endpoint" "example" {
  id = "se-abc12345"
}

output "endpoint_port" {
  value = data.vnpaycloud_service_endpoint.example.port
}
```

## Schema

### Required

- `id` (String) The ID of the service endpoint.

### Read-Only

- `name` (String) The name of the service endpoint.
- `description` (String) A human-readable description.
- `provider_id` (String) The service provider ID.
- `service_id` (String) The published service ID.
- `service_gateway_id` (String) The service gateway ID.
- `port` (Integer) The listener port.
- `allowed_cidrs` (List of String) Source CIDRs allowed to reach the endpoint.
- `listener_id` (String) The underlying listener ID.
- `pool_id` (String) The underlying pool ID.
- `health_monitor_id` (String) The underlying health monitor ID.
- `pool_member_ids` (List of String) The underlying pool member IDs.
- `operating_status` (String) The Octavia operating status.
- `provisioning_status` (String) The Octavia provisioning status.
- `status` (String) Lifecycle status.
- `created_at` (String) Creation timestamp.
