---
page_title: "vnpaycloud_service_endpoint Resource - VNPayCloud"
subcategory: "Networking"
description: |-
  Manages a service endpoint within VNPayCloud.
---

# vnpaycloud_service_endpoint (Resource)

Manages a service endpoint within VNPayCloud. A service endpoint exposes an admin-published **service** (identified by `provider_id` + `service_id`) on a listener port of an existing [service gateway](service_gateway.md). It provisions asynchronously, wiring up the listener, pool, health monitor, and pool members behind the scenes.

~> **Asynchronous resource.** Create and delete return immediately; Terraform then waits for the endpoint to converge to `active` (or `deleted`). The default timeout is 10 minutes.

~> **All endpoints on a gateway share one provider.** The first endpoint created on a service gateway fixes its provider; subsequent endpoints must use the same `provider_id`, a distinct `port`, and the gateway must be `active`.

~> **The provider must serve your zone.** `provider_id` must reference a provider that has a network configured in the gateway's availability zone. If it does not, creation fails with `This datacenter is invalid in this provider`. Providers are admin-managed; the [`vnpaycloud_service_providers`](../data-sources/service_providers.md) data source lists every provider in the project and cannot pre-filter by zone, so confirm with your administrator that the provider is published in your zone.

~> **Provider gateway quota.** The first service endpoint on a gateway consumes one service-gateway slot from the provider's network in that datacenter. If the provider has no slots left, creation fails with `No available service gateway quota in this provider for this datacenter`. This is an admin-managed capacity limit — ask your administrator to raise the provider's quota in your zone.

## Example Usage

```hcl
# Discover the provider and service IDs
data "vnpaycloud_service_providers" "all" {}

locals {
  provider = one([for p in data.vnpaycloud_service_providers.all.providers : p if p.name == "acme-provider"])
}

data "vnpaycloud_services" "acme" {
  provider_id = local.provider.id
}

locals {
  service = one([for s in data.vnpaycloud_services.acme.services : s if s.name == "acme-db"])
}

resource "vnpaycloud_service_endpoint" "example" {
  name               = "my-service-endpoint"
  description        = "Endpoint for acme-db"
  provider_id        = local.provider.id
  service_id         = local.service.id
  service_gateway_id = vnpaycloud_service_gateway.example.id
  port               = 5432
  allowed_cidrs      = ["10.0.0.0/16"]
}
```

## Schema

### Required

- `name` (String) The name of the service endpoint. Length `3`–`250`. Allowed characters: ASCII letters, digits, spaces, and `-` `_` `.` (must match `^[a-zA-Z0-9-_. ]*$`). Unique per zone.
- `provider_id` (String, ForceNew) The ID of the service provider. Look it up with the [`vnpaycloud_service_providers`](../data-sources/service_providers.md) data source. Changing it recreates the resource.
- `service_id` (String, ForceNew) The ID of the published service. Look it up with the [`vnpaycloud_services`](../data-sources/services.md) data source. Changing it recreates the resource.
- `service_gateway_id` (String, ForceNew) The ID of the [service gateway](service_gateway.md) this endpoint is created on. Changing it recreates the resource.
- `port` (Integer, ForceNew) The listener port (`1`–`65535`). Must be unique among endpoints on the same gateway. Changing it recreates the resource.
- `allowed_cidrs` (List of String) Source CIDRs allowed to reach the endpoint. Must contain **at least one** CIDR; an empty list is rejected at plan time. To allow all sources, set `["0.0.0.0/0"]`. IPv4 and IPv6 CIDRs are accepted. Updatable in place.

### Optional

- `description` (String) A human-readable description.

### Read-Only

- `id` (String) The service endpoint ID.
- `listener_id` (String) The ID of the underlying listener.
- `pool_id` (String) The ID of the underlying pool.
- `health_monitor_id` (String) The ID of the underlying health monitor.
- `pool_member_ids` (List of String) The IDs of the underlying pool members.
- `operating_status` (String) The operating status.
- `provisioning_status` (String) The provisioning status.
- `status` (String) Lifecycle status: `active`, `creating`, `deleting`, `error`, `deleted`, `unknown`.
- `created_at` (String) Creation timestamp (RFC 3339).

## In-place updates

`name`, `description`, and `allowed_cidrs` are updatable in place. `provider_id`, `service_id`, `service_gateway_id`, and `port` are `ForceNew`.

## Timeouts

- `create` - (Default `10 minutes`)
- `update` - (Default `10 minutes`)
- `delete` - (Default `10 minutes`)

## Import

```shell
terraform import vnpaycloud_service_endpoint.example <service-endpoint-id>
```
