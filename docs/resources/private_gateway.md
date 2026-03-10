---
page_title: "vnpaycloud_private_gateway Resource - VNPayCloud"
subcategory: "Networking"
description: |-
  Manages a private gateway resource within VNPayCloud.
---

# vnpaycloud_private_gateway (Resource)

Manages a private gateway resource within VNPayCloud. A private gateway provides dedicated private connectivity between your VPC and on-premises networks or other VNPayCloud services.

## Example Usage

```hcl
resource "vnpaycloud_private_gateway" "example" {
  name        = "my-private-gateway"
  description = "Private gateway for on-premises connectivity"
}
```

## Schema

### Required

- `name` (String) The name of the private gateway.

### Optional

- `description` (String) A description of the private gateway.

### Read-Only

- `id` (String) The ID of the private gateway.
- `load_balancer_id` (String) The ID of the internal load balancer provisioned for the private gateway.
- `subnet_id` (String) The ID of the subnet the private gateway is deployed into.
- `flavor_id` (String) The ID of the compute flavor used for the private gateway.
- `status` (String) The current status of the private gateway.
- `created_at` (String) The creation timestamp of the private gateway.

## Timeouts

- `create` - (Default `10 minutes`) Used for creating the private gateway.
- `delete` - (Default `10 minutes`) Used for deleting the private gateway.

## Import

Private gateways can be imported using the `id`:

```shell
terraform import vnpaycloud_private_gateway.example <private-gateway-id>
```
