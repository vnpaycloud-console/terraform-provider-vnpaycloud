---
page_title: "vnpaycloud_keypair Resource - VNPayCloud"
subcategory: "Compute"
description: |-
  Manages an SSH key pair within VNPayCloud.
---

# vnpaycloud_keypair (Resource)

Manages an SSH key pair within VNPayCloud. Key pairs are used to authenticate SSH access to compute instances. You can either import an existing public key or let VNPayCloud generate a new key pair for you.

~> **Note:** When VNPayCloud generates a key pair (i.e., `public_key` is omitted), the `private_key` attribute is only available immediately after creation. It is not stored remotely and cannot be retrieved later. Ensure you save the private key from the Terraform state or output immediately.

## Example Usage

### Importing an existing public key

```hcl
resource "vnpaycloud_keypair" "existing" {
  name       = "my-existing-key"
  public_key = file("~/.ssh/id_rsa.pub")
}
```

### Generating a new key pair

```hcl
resource "vnpaycloud_keypair" "generated" {
  name = "my-generated-key"
}

output "private_key_pem" {
  value     = vnpaycloud_keypair.generated.private_key
  sensitive = true
}
```

## Schema

### Required

- `name` (String, ForceNew) The name of the key pair. Must be unique within the project. Changing this creates a new key pair.

### Optional

- `public_key` (String, ForceNew, Computed) The OpenSSH-formatted public key to import. If omitted, VNPayCloud will generate a new key pair and the private key will be returned in `private_key`. Changing this creates a new key pair.

### Read-Only

- `id` (String) The ID of the key pair.
- `private_key` (String, Sensitive) The private key in PEM format. Only populated when VNPayCloud generates the key pair (i.e., `public_key` was not provided). This value is only available at creation time.
- `fingerprint` (String) The MD5 fingerprint of the public key.
- `created_at` (String) The creation timestamp of the key pair in ISO 8601 format.

## Timeouts

- `create` - (Default `5 minutes`) Used for creating the key pair.
- `delete` - (Default `5 minutes`) Used for deleting the key pair.

## Import

Key pairs can be imported using the `name`:

```shell
terraform import vnpaycloud_keypair.example <keypair-name>
```

~> **Note:** Importing a key pair does not import the private key. The `private_key` attribute will be empty after import.
