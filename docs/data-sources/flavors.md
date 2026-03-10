---
page_title: "vnpaycloud_flavors Data Source - VNPayCloud"
subcategory: "Compute"
description: |-
  List all flavors in VNPayCloud.
---

# vnpaycloud_flavors (Data Source)

Use this data source to list all available compute flavors in the current zone.

## Example Usage

```hcl
data "vnpaycloud_flavors" "all" {}

output "all_flavor_names" {
  value = data.vnpaycloud_flavors.all.flavors[*].name
}

output "high_cpu_flavors" {
  value = [
    for f in data.vnpaycloud_flavors.all.flavors :
    f.name if f.vcpus >= 8
  ]
}
```

## Schema

### Read-Only

- `flavors` (List of Object) List of flavors. Each element contains:
  - `id` (String) The unique identifier of the flavor.
  - `name` (String) The name of the flavor.
  - `vcpus` (Number) The number of virtual CPUs.
  - `ram_mb` (Number) The amount of RAM in megabytes (MB).
  - `disk_gb` (Number) The root disk size in gigabytes (GB).
  - `is_public` (Boolean) Whether the flavor is publicly available.
  - `zone` (String) The availability zone where this flavor is available.
