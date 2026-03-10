---
page_title: "vnpaycloud_image Data Source - VNPayCloud"
subcategory: "Compute"
description: |-
  Get information about an image in VNPayCloud.
---

# vnpaycloud_image (Data Source)

Use this data source to get information about an existing compute image, including its OS type, version, and status.

## Example Usage

```hcl
data "vnpaycloud_image" "ubuntu" {
  name = "Ubuntu 22.04 LTS"
}

output "image_id" {
  value = data.vnpaycloud_image.ubuntu.id
}
```

```hcl
data "vnpaycloud_image" "by_id" {
  id = "image-abc123"
}
```

## Schema

### Optional (filter)

- `id` (String) The ID of the image.
- `name` (String) The name of the image.

### Read-Only

- `os_type` (String) The operating system type (e.g., `ubuntu`, `centos`, `windows`).
- `os_version` (String) The operating system version (e.g., `Ubuntu 22.04 LTS`).
- `min_disk_gb` (Number) The minimum disk size in gigabytes (GB) required to use this image.
- `status` (String) The current status of the image (e.g., `active`).
