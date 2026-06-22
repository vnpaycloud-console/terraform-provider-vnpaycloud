---
page_title: "vnpaycloud_services Data Source - VNPayCloud"
subcategory: "Networking"
description: |-
  List published services available for service endpoints in VNPayCloud.
---

# vnpaycloud_services (Data Source)

Returns the catalogue of published services that can be exposed through a [service endpoint](../resources/service_endpoint.md), optionally filtered by provider and/or name. Each service belongs to a [provider](service_providers.md).

## Example Usage

```hcl
data "vnpaycloud_service_providers" "all" {}

locals {
  acme = one([for p in data.vnpaycloud_service_providers.all.providers : p if p.name == "acme-provider"])
}

data "vnpaycloud_services" "acme" {
  provider_id = local.acme.id
}

output "service_names" {
  value = [for s in data.vnpaycloud_services.acme.services : s.name]
}
```

## Schema

### Optional (filter)

- `provider_id` (String) Only return services belonging to this provider.
- `name` (String) Only return services matching this name.

### Read-Only

- `services` (List of Object) The catalogue of services.
  - `id` (String) Service ID — pass this to `vnpaycloud_service_endpoint.service_id`.
  - `name` (String) Service name.
  - `description` (String) Free-form description.
  - `provider_id` (String) The provider this service belongs to.
  - `service_domain` (String) The service domain.
  - `status` (String) Service status.
