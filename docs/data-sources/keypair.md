---
page_title: "vnpaycloud_keypair Data Source - VNPayCloud"
subcategory: "Compute"
description: |-
  Get information about a key pair in VNPayCloud.
---

# vnpaycloud_keypair (Data Source)

Use this data source to get information about an existing SSH key pair, including its public key and fingerprint. This is useful when you need to reference a key pair that was created outside of Terraform.

## Example Usage

```hcl
data "vnpaycloud_keypair" "example" {
  name = "my-ssh-key"
}

output "public_key" {
  value = data.vnpaycloud_keypair.example.public_key
}

output "fingerprint" {
  value = data.vnpaycloud_keypair.example.fingerprint
}
```

## Schema

### Required (filter)

- `name` (String) The name of the key pair to look up.

### Read-Only

- `public_key` (String) The OpenSSH-formatted public key of the key pair.
- `fingerprint` (String) The MD5 fingerprint of the public key.
- `created_at` (String) The timestamp when the key pair was created, in ISO 8601 format.
