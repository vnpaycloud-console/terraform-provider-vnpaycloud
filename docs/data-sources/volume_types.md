---
page_title: "vnpaycloud_volume_types Data Source - VNPayCloud"
subcategory: "Storage"
description: |-
  List all volume types in VNPayCloud.
---

# vnpaycloud_volume_types (Data Source)

Use this data source to list all available volume types in the current zone.

## Example Usage

```hcl
data "vnpaycloud_volume_types" "all" {}

output "all_volume_type_names" {
  value = data.vnpaycloud_volume_types.all.volume_types[*].name
}

output "encrypted_volume_types" {
  value = [
    for vt in data.vnpaycloud_volume_types.all.volume_types :
    vt.name if vt.is_encrypted
  ]
}
```

## Schema

### Read-Only

- `volume_types` (List of Object) List of volume types. Each element contains:
  - `id` (String) The unique identifier of the volume type.
  - `name` (String) The name of the volume type.
  - `iops` (Number) The provisioned IOPS for this volume type.
  - `is_encrypted` (Boolean) Whether volumes of this type are encrypted at rest.
  - `is_multiattach` (Boolean) Whether volumes of this type support multi-attach.
  - `zone` (String) The availability zone where this volume type is available.
