---
page_title: "vnpaycloud_volume_type Data Source - VNPayCloud"
subcategory: "Storage"
description: |-
  Get information about a volume type in VNPayCloud.
---

# vnpaycloud_volume_type (Data Source)

Use this data source to get information about an existing volume type, including its IOPS, encryption, and multi-attach capabilities.

## Example Usage

```hcl
data "vnpaycloud_volume_type" "ssd" {
  name = "c1-standard"
}

output "volume_type_iops" {
  value = data.vnpaycloud_volume_type.ssd.iops
}
```

```hcl
data "vnpaycloud_volume_type" "by_id" {
  id = "volume-type-abc123"
}
```

## Schema

### Optional (filter)

- `id` (String) The ID of the volume type.
- `name` (String) The name of the volume type.

### Read-Only

- `iops` (Number) The provisioned IOPS for this volume type.
- `is_encrypted` (Boolean) Whether volumes of this type are encrypted at rest.
- `is_multiattach` (Boolean) Whether volumes of this type support attachment to multiple instances simultaneously.
- `zone` (String) The availability zone where this volume type is available.
