---
page_title: "vnpaycloud_service_providers Data Source - VNPayCloud"
subcategory: "Networking"
description: |-
  List service providers available for service endpoints in VNPayCloud.
---

# vnpaycloud_service_providers (Data Source)

Returns the catalogue of service providers whose [services](services.md) can be exposed through a [service endpoint](../resources/service_endpoint.md). Providers and their services are published by the platform administrator.

## Example Usage

```hcl
data "vnpaycloud_service_providers" "all" {}

output "provider_names" {
  value = [for p in data.vnpaycloud_service_providers.all.providers : p.name]
}

locals {
  acme = one([for p in data.vnpaycloud_service_providers.all.providers : p if p.name == "acme-provider"])
}
```

## Schema

### Read-Only

- `providers` (List of Object) The catalogue of service providers.
  - `id` (String) Provider ID — pass this to `vnpaycloud_service_endpoint.provider_id`.
  - `name` (String) Provider name.
  - `status` (String) Provider status.
