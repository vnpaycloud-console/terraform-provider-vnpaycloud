---
page_title: "vnpaycloud_buckets Data Source - VNPayCloud"
subcategory: "Object Storage"
description: |-
  List all object storage buckets in VNPayCloud.
---

# vnpaycloud_buckets (Data Source)

Use this data source to list all object storage buckets in the current project.

## Example Usage

```hcl
data "vnpaycloud_buckets" "all" {}

output "all_bucket_names" {
  value = data.vnpaycloud_buckets.all.buckets[*].bucket_name
}

output "buckets_by_region" {
  value = {
    for b in data.vnpaycloud_buckets.all.buckets :
    b.bucket_name => b.region
  }
}

output "buckets_with_policy" {
  value = [
    for b in data.vnpaycloud_buckets.all.buckets :
    b.bucket_name if b.policy_name != ""
  ]
}
```

## Schema

### Read-Only

- `buckets` (List of Object) List of object storage buckets. Each element contains:
  - `id` (String) The unique identifier of the bucket.
  - `bucket_name` (String) The name of the bucket.
  - `region` (String) The region where the bucket is hosted.
  - `created_at` (String) The timestamp when the bucket was created, in ISO 8601 format.
  - `policy_name` (String) The name of the access policy applied to the bucket, if any.
