---
page_title: "vnpaycloud_certificates Data Source - VNPayCloud"
subcategory: "Certificate"
description: |-
  Returns the metadata of every TLS certificate available in the project's zone.
---

# vnpaycloud_certificates (Data Source)

Returns metadata for every TLS certificate in the project's zone. Use it to discover certificate IDs to reference from a [`vnpaycloud_lb_listener`](../resources/lb_listener.md)'s `certificate_id`, `sni_certificate_ids`, or `certificate_authority_id` — instead of hardcoding opaque IDs.

~> **Metadata only.** This data source never returns secret material. Private keys, certificate PEM bodies, and secret references are deliberately excluded so they can never be written to Terraform state — only non-sensitive metadata is exposed.

## Example Usage

### List every certificate

```hcl
data "vnpaycloud_certificates" "all" {}

output "certificate_names" {
  value = [for c in data.vnpaycloud_certificates.all.certificates : c.name]
}
```

### Reference a certificate by name in a listener

```hcl
data "vnpaycloud_certificates" "all" {}

locals {
  server_cert = one([for c in data.vnpaycloud_certificates.all.certificates : c if c.name == "my-server-cert"])
}

resource "vnpaycloud_lb_listener" "https" {
  name             = "https"
  load_balancer_id = vnpaycloud_lb_loadbalancer.app.id
  protocol         = "HTTPS"
  protocol_port    = 443
  certificate_id   = local.server_cert.id
}
```

## Schema

### Read-Only

- `certificates` (List of Object) The certificates available in the zone (metadata only).
  - `id` (String) Backend certificate ID — pass to a listener's `certificate_id` / `sni_certificate_ids` / `certificate_authority_id`.
  - `name` (String) Certificate name.
  - `cert_type` (String) Certificate type: `CT_SIGNED`, `CT_SELF_SIGNED`, `CT_CA`.
  - `domain_name` (String) Domain the certificate was issued for.
  - `description` (String) Free-form description.
  - `expiration` (String) Expiry timestamp in ISO 8601 / RFC 3339 format.
  - `status` (String) Lifecycle status: `active`, `creating`, `disabled`, `deleting`, `deleted`, `error`, or `unknown`.
  - `zone_id` (String) Zone the certificate belongs to.
  - `load_balancer_ids` (List of String) IDs of load balancers currently using this certificate.
