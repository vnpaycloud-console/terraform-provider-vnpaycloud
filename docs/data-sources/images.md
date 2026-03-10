---
page_title: "vnpaycloud_images Data Source - VNPayCloud"
subcategory: "Compute"
description: |-
  List all images in VNPayCloud.
---

# vnpaycloud_images (Data Source)

Use this data source to list all available compute images in the current zone.

## Example Usage

```hcl
data "vnpaycloud_images" "all" {}

output "all_image_names" {
  value = data.vnpaycloud_images.all.images[*].name
}

output "ubuntu_images" {
  value = [
    for img in data.vnpaycloud_images.all.images :
    img.name if img.os_type == "ubuntu"
  ]
}
```

## Schema

### Read-Only

- `images` (List of Object) List of images. Each element contains:
  - `id` (String) The unique identifier of the image.
  - `name` (String) The name of the image.
  - `os_type` (String) The operating system type (e.g., `ubuntu`, `centos`, `windows`).
  - `os_version` (String) The operating system version (e.g., `Ubuntu 22.04 LTS`).
  - `min_disk_gb` (Number) The minimum disk size in gigabytes (GB) required to use this image.
  - `status` (String) The current status of the image (e.g., `active`).
