---
page_title: "vnpaycloud_flavor Data Source - VNPayCloud"
subcategory: "Compute"
description: |-
  Get information about a flavor in VNPayCloud.
---

# vnpaycloud_flavor (Data Source)

Use this data source to get information about an existing compute flavor, including its vCPU count, RAM, and disk size.

## Example Usage

```hcl
data "vnpaycloud_flavor" "example" {
  name = "a-pro-small.2x4"
}

output "flavor_vcpus" {
  value = data.vnpaycloud_flavor.example.vcpus
}
```

```hcl
data "vnpaycloud_flavor" "by_id" {
  id = "flavor-abc123"
}
```

## Schema

### Optional (filter)

- `id` (String) The ID of the flavor.
- `name` (String) The name of the flavor.

### Read-Only

- `vcpus` (Number) The number of virtual CPUs.
- `ram_mb` (Number) The amount of RAM in megabytes (MB).
- `disk_gb` (Number) The root disk size in gigabytes (GB).
- `is_public` (Boolean) Whether the flavor is publicly available.
- `zone` (String) The availability zone where this flavor is available.
