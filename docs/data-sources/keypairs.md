---
page_title: "vnpaycloud_keypairs Data Source - VNPayCloud"
subcategory: "Compute"
description: |-
  List all key pairs in VNPayCloud.
---

# vnpaycloud_keypairs (Data Source)

Use this data source to list all SSH key pairs in the current project.

## Example Usage

```hcl
data "vnpaycloud_keypairs" "all" {}

output "all_keypair_names" {
  value = data.vnpaycloud_keypairs.all.key_pairs[*].name
}

output "keypair_public_keys" {
  value = {
    for kp in data.vnpaycloud_keypairs.all.key_pairs :
    kp.name => kp.public_key
  }
}
```

## Schema

### Read-Only

- `key_pairs` (List of Object) List of key pairs. Each element contains:
  - `id` (String) The ID of the key pair.
  - `name` (String) The name of the key pair.
  - `public_key` (String) The OpenSSH-formatted public key of the key pair.
  - `created_at` (String) The timestamp when the key pair was created, in ISO 8601 format.
