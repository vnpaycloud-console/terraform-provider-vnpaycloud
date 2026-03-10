---
page_title: "vnpaycloud_bucket Resource - VNPayCloud"
subcategory: "Object Storage"
description: |-
  Manages an object storage bucket within VNPayCloud.
---

# vnpaycloud_bucket (Resource)

Manages an object storage bucket within VNPayCloud. Buckets are the fundamental containers for storing objects (files) in VNPayCloud Object Storage, which is compatible with the S3 API. Buckets are region-specific and can optionally have Object Lock enabled for WORM (Write Once, Read Many) compliance.

~> **Note:** This resource does not support in-place updates. All attributes are ForceNew — any change will destroy the existing bucket and create a new one. Destroying a bucket that contains objects will fail unless all objects are deleted first.

## Example Usage

### Standard bucket

```hcl
resource "vnpaycloud_bucket" "assets" {
  bucket_name = "my-application-assets"
  region      = "HCM"
}
```

### Bucket with a specific storage policy and Object Lock

```hcl
resource "vnpaycloud_bucket" "compliance" {
  bucket_name        = "compliance-archive-2024"
  region             = "HAN"
  storage_policy_id  = "policy-cold-storage"
  enable_object_lock = true
}
```

### Using with AWS provider for S3-compatible operations

```hcl
resource "vnpaycloud_bucket" "data" {
  bucket_name = "app-data-bucket"
  region      = "HCM"
}

# After creating, use the S3-compatible endpoint with your preferred S3 provider
output "bucket_name" {
  value = vnpaycloud_bucket.data.bucket_name
}
```

## Schema

### Required

- `bucket_name` (String, ForceNew) The globally unique name for the bucket. Must comply with DNS naming conventions: 3-63 characters, lowercase letters, numbers, and hyphens only, must start and end with a letter or number. Changing this creates a new bucket.
- `region` (String, ForceNew) The region where the bucket will be created (e.g., `HCM`, `HAN`). Changing this creates a new bucket.

### Optional

- `storage_policy_id` (String, ForceNew) The ID of the storage policy to apply to the bucket, which determines the storage tier and replication behavior. If not specified, the region default policy is used. Changing this creates a new bucket.
- `enable_object_lock` (Boolean, ForceNew) Whether to enable S3 Object Lock on the bucket. When enabled, objects can be stored using WORM (Write Once, Read Many) model to prevent deletion or modification for a defined period. Object Lock cannot be disabled after the bucket is created. Defaults to `false`. Changing this creates a new bucket.

### Read-Only

- `id` (String) The ID of the bucket.
- `created_at` (String) The creation timestamp of the bucket in ISO 8601 format.
- `policy_name` (String) The name of the storage policy applied to the bucket.

## Timeouts

This resource uses default Terraform timeouts. No explicit timeout configuration is supported.

## Import

Buckets can be imported using the `id`:

```shell
terraform import vnpaycloud_bucket.example <bucket-id>
```
