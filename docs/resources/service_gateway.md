---
page_title: "vnpaycloud_service_gateway Resource - VNPayCloud"
subcategory: "Networking"
description: |-
  Manages a service gateway within VNPayCloud.
---

# vnpaycloud_service_gateway (Resource)

Manages a service gateway within VNPayCloud. A service gateway is a managed load balancer that fronts one or more [service endpoints](service_endpoint.md), letting workloads in your VPC reach published provider services privately. It is created in a subnet with a service-endpoint flavor and provisions asynchronously.

~> **Asynchronous resource.** Create and delete return immediately; Terraform then waits for the gateway to converge to `active` (or `deleted`). The default timeout is 10 minutes.

## Example Usage

```hcl
# Look up an available service-gateway flavor by name
data "vnpaycloud_service_gateway_flavors" "all" {}

locals {
  sg_flavor = one([for f in data.vnpaycloud_service_gateway_flavors.all.flavors : f if f.name == "sgw.small"])
}

resource "vnpaycloud_service_gateway" "example" {
  name         = "my-service-gateway"
  description  = "Service gateway for provider services"
  vpc_id       = vnpaycloud_vpc.main.id
  subnet_id    = vnpaycloud_subnet.main.id
  flavor_id    = local.sg_flavor.id
  allowed_icmp = true
}
```

## Schema

### Required

- `name` (String) The name of the service gateway. Length `3`–`250`. Allowed characters: ASCII letters, digits, spaces, and `-` `_` `.` (must match `^[a-zA-Z0-9-_. ]*$`). Unique per zone.
- `subnet_id` (String, ForceNew) The ID of the subnet where the gateway's VIP is allocated. Must be in the same zone as the provider `zone_id` and must not be a Kubernetes subnet. Changing it recreates the resource.
- `flavor_id` (String) The ID of the service-gateway flavor (purpose `service_endpoint`). Look it up with the [`vnpaycloud_service_gateway_flavors`](../data-sources/service_gateway_flavors.md) data source. Updatable in place: changing it resizes the gateway via a dedicated action (the gateway briefly re-provisions).

### Optional

- `description` (String) A human-readable description.
- `vpc_id` (String, ForceNew) The ID of the VPC the gateway belongs to. **Required when `subnet_id` belongs to a VPC** — in that case it must be set to that VPC's ID. Omit it only when `subnet_id` is a standalone subnet that does not belong to any VPC. Changing it recreates the resource.
- `allowed_icmp` (Boolean, Computed) Whether ICMP (ping) to the gateway VIP is allowed. Applied in-place via a dedicated action. Defaults to the server value when omitted.

### Read-Only

- `id` (String) The service gateway ID.
- `load_balancer_id` (String) The ID of the underlying managed load balancer.
- `vip_address` (String) The virtual IP address assigned to the gateway.
- `port_id` (String) The ID of the VIP port.
- `operating_status` (String) The operating status of the underlying load balancer.
- `provisioning_status` (String) The provisioning status of the underlying load balancer.
- `status` (String) Lifecycle status: `active`, `creating`, `deleting`, `error`, `deleted`, `unknown`.
- `created_at` (String) Creation timestamp (RFC 3339).

## In-place updates

`name`, `description`, `flavor_id` (resize action), and `allowed_icmp` (ICMP action) are updatable in place. `subnet_id` and `vpc_id` are `ForceNew`.

## Timeouts

- `create` - (Default `10 minutes`)
- `update` - (Default `10 minutes`)
- `delete` - (Default `10 minutes`)

## Import

```shell
terraform import vnpaycloud_service_gateway.example <service-gateway-id>
```
