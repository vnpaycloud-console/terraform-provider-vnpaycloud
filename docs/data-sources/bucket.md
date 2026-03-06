---
page_title: "vnpaycloud_bucket Data Source - VNPayCloud"
subcategory: "Object Storage"
description: |-
  Get information about an object storage bucket in VNPayCloud.
---

# vnpaycloud_bucket (Data Source)

Use this data source to get information about an existing object storage bucket, including its region, storage usage, and applied bucket policy.

## Example Usage

```hcl
data "vnpaycloud_bucket" "example" {
  bucket_name = "my-app-assets"
}

output "bucket_region" {
  value = data.vnpaycloud_bucket.example.region
}

output "bucket_size_bytes" {
  value = data.vnpaycloud_bucket.example.size_bytes
}

output "object_count" {
  value = data.vnpaycloud_bucket.example.object_count
}
```

## Schema

### Required (filter)

- `bucket_name` (String) The globally unique name of the object storage bucket.

### Read-Only

- `region` (String) The region where the bucket is located (e.g., `HN`, `HCM`).
- `created_at` (String) The timestamp when the bucket was created, in ISO 8601 format.
- `policy_name` (String) The name of the access policy applied to this bucket (e.g., `private`, `public-read`, `public-read-write`). Empty string if no policy is applied.
- `size_bytes` (Number) The total size of all objects stored in this bucket, in bytes.
- `object_count` (Number) The total number of objects stored in this bucket.
