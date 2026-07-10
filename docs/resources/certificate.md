---
page_title: "vnpaycloud_certificate Resource - VNPayCloud"
subcategory: "Certificate"
description: |-
  Manages a TLS certificate in VNPayCloud.
---

# vnpaycloud_certificate (Resource)

Manages a TLS certificate that can be referenced by other resources (for example an `vnpaycloud_lb_listener` for HTTPS, or a database instance for TLS). A certificate can be created in one of three ways, selected with the `type` argument:

- `self_signed` — the backend generates a private key and a self-signed certificate from the subject fields you provide.
- `import` — you upload an existing private key + certificate (and optionally an intermediate CA chain).
- `import_ca` — you upload a CA certificate.

~> **Note:** Certificates are immutable. Any change to a configured argument forces the certificate to be destroyed and recreated.

~> **Secrets:** For `self_signed`, the generated private key is stored by the backend and is never returned to Terraform. For `import`, `private_key` / `certificate_body` / `intermediate_ca` are write-only sensitive inputs kept in state.

## Example Usage

### Self-signed certificate

```hcl
resource "vnpaycloud_certificate" "self" {
  type              = "self_signed"
  name              = "my-self-signed"
  description       = "self-signed cert for testing"
  country           = "VN"
  province          = "Ha Noi"
  city              = "Ha Noi"
  organization      = "VNPay"
  organization_unit = "Cloud"
  domain_name       = "example.vnpaycloud.vn"
  email             = "admin@vnpaycloud.vn"
  key_type          = "4096 RSA"
  expiration        = 3650
  digest_algorithm  = "SHA-512"
}
```

### Import an existing certificate

```hcl
resource "vnpaycloud_certificate" "imported" {
  type             = "import"
  name             = "my-imported-cert"
  private_key      = file("server.key")
  certificate_body = file("server.crt")
  intermediate_ca  = file("chain.crt") # optional
}
```

### Import a CA certificate

```hcl
resource "vnpaycloud_certificate" "ca" {
  type             = "import_ca"
  name             = "my-ca"
  certificate_body = file("ca.crt")
}
```

## Schema

### Required

- `type` (String, ForceNew) How the certificate is created. One of `self_signed`, `import`, `import_ca`.
- `name` (String, ForceNew) The certificate name. 3-255 characters: letters, digits, space, `-`, `_`, `.`.

### Optional

- `description` (String, ForceNew) A human-readable description.

#### For `type = self_signed`

The following are **required** when `type = self_signed`:

- `domain_name` (String, ForceNew) Common Name (CN) / domain for the certificate.
- `country` (String, ForceNew) Subject country (e.g. `VN`).
- `province` (String, ForceNew) Subject state/province.
- `organization` (String, ForceNew) Subject organization.
- `organization_unit` (String, ForceNew) Subject organizational unit.
- `email` (String, ForceNew) Subject email address.

Optional subject fields:

- `city` (String, ForceNew) Subject city/locality.
- `key_type` (String, ForceNew) Private key type. One of `2048 RSA`, `4096 RSA`. Defaults to `4096 RSA`.
- `expiration` (Number, ForceNew) Lifetime in days. Defaults to `3650`.
- `digest_algorithm` (String, ForceNew) Signing digest. One of `SHA-256`, `SHA-512`. Defaults to `SHA-512`.

#### For `type = import` / `type = import_ca`

- `private_key` (String, ForceNew, Sensitive) PEM private key. Required for `import`.
- `certificate_body` (String, ForceNew, Sensitive) PEM certificate body. Required for `import` and `import_ca`.
- `intermediate_ca` (String, ForceNew, Sensitive) PEM intermediate CA chain. Optional, `import` only.

### Read-Only

- `id` (String) The ID of the certificate.
- `cert_type` (String) The backend certificate type (e.g. `CT_SELF_SIGNED`, `CT_SIGNED`, `CT_CA`).
- `domain_name` (String) The domain name of the certificate.
- `expires_at` (String) The expiration timestamp in ISO 8601 format.
- `status` (String) The current status of the certificate.
- `zone_id` (String) The zone where the certificate resides.
- `load_balancer_ids` (List of String) IDs of load balancers currently using this certificate.

## Import

Certificates can be imported using the `id`:

```shell
terraform import vnpaycloud_certificate.example <certificate-id>
```

After import, keep the Terraform configuration to fields that can be read back from the API, such as `type`, `name`, `description`, and `domain_name` when present. Do not include write-only upload inputs (`private_key`, `certificate_body`, `intermediate_ca`) or self-signed creation-only subject fields (`country`, `province`, `city`, `organization`, `organization_unit`, `email`, `key_type`, `expiration`, `digest_algorithm`) unless you intend Terraform to replace the certificate.
