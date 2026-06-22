---
page_title: "vnpaycloud_service_gateway_flavors Data Source - VNPayCloud"
subcategory: "Networking"
description: |-
  Returns the catalogue of service-gateway flavors available in the project's zone.
---

# vnpaycloud_service_gateway_flavors (Data Source)

Returns the catalogue of flavors available for creating a [`vnpaycloud_service_gateway`](../resources/service_gateway.md). These are load-balancer flavors with the `service_endpoint` purpose.

## Example Usage

```hcl
data "vnpaycloud_service_gateway_flavors" "all" {}

output "sg_flavor_names" {
  value = [for f in data.vnpaycloud_service_gateway_flavors.all.flavors : f.name]
}

locals {
  small = one([for f in data.vnpaycloud_service_gateway_flavors.all.flavors : f if f.name == "sgw.small"])
}

resource "vnpaycloud_service_gateway" "app" {
  name      = "app-sg"
  subnet_id = var.subnet_id
  flavor_id = local.small.id
}
```

## Schema

### Read-Only

- `flavors` (List of Object) The catalogue of service-gateway flavors.
  - `id` (String) Backend flavor ID — pass this to `vnpaycloud_service_gateway.flavor_id`.
  - `name` (String) Flavor name.
  - `description` (String) Free-form description.
