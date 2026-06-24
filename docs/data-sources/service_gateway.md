---
page_title: "vnpaycloud_service_gateway Data Source - VNPayCloud"
subcategory: "Networking"
description: |-
  Get information about a service gateway in VNPayCloud.
---

# vnpaycloud_service_gateway (Data Source)

Use this data source to get information about an existing [service gateway](../resources/service_gateway.md) by its ID.

## Example Usage

```hcl
data "vnpaycloud_service_gateway" "example" {
  id = "sgw-abc12345"
}

output "gateway_vip" {
  value = data.vnpaycloud_service_gateway.example.vip_address
}
```

## Schema

### Required

- `id` (String) The ID of the service gateway.

### Read-Only

- `name` (String) The name of the service gateway.
- `description` (String) A human-readable description.
- `subnet_id` (String) The subnet ID where the VIP is allocated.
- `vpc_id` (String) The VPC ID the gateway belongs to.
- `flavor_id` (String) The service-gateway flavor ID.
- `allowed_icmp` (Boolean) Whether ICMP to the VIP is allowed.
- `vip_address` (String) The virtual IP address.
- `load_balancer_id` (String) The underlying managed load balancer ID.
- `port_id` (String) The VIP port ID.
- `operating_status` (String) The operating status.
- `provisioning_status` (String) The provisioning status.
- `status` (String) Lifecycle status.
- `created_at` (String) Creation timestamp.
