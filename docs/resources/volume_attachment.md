---
page_title: "vnpaycloud_volume_attachment Resource - VNPayCloud"
subcategory: "Storage"
description: |-
  Manages a volume attachment to a compute instance within VNPayCloud.
---

# vnpaycloud_volume_attachment (Resource)

Manages the attachment of a block storage volume to a compute instance within VNPayCloud. This resource creates a persistent association between a volume and a server.

~> **Note:** All attributes are ForceNew — any change to `volume_id` or `server_id` will destroy the existing attachment and create a new one.

## Example Usage

```hcl
resource "vnpaycloud_volume" "data" {
  name        = "app-data-volume"
  size        = 50
  volume_type = "SSD"
}

resource "vnpaycloud_instance" "app" {
  name           = "app-server"
  image          = "ubuntu-22.04"
  flavor         = "s.4c8r"
  root_disk_gb   = 20
  root_disk_type = "SSD"
}

resource "vnpaycloud_volume_attachment" "data_attach" {
  volume_id = vnpaycloud_volume.data.id
  server_id = vnpaycloud_instance.app.id
}
```

## Schema

### Required

- `volume_id` (String, ForceNew) The ID of the volume to attach. Changing this creates a new attachment.
- `server_id` (String, ForceNew) The ID of the compute instance to attach the volume to. Changing this creates a new attachment.

### Read-Only

- `id` (String) The ID of the volume attachment.
- `device` (String) The device path on the instance where the volume is exposed (e.g., `/dev/vdb`).
- `status` (String) The current status of the attachment (e.g., `attached`, `detaching`).
- `attached_at` (String) The timestamp when the volume was attached, in ISO 8601 format.

## Timeouts

- `create` - (Default `10 minutes`) Used for attaching the volume to the instance.
- `delete` - (Default `10 minutes`) Used for detaching the volume from the instance.

## Import

Volume attachments can be imported using the attachment `id`:

```shell
terraform import vnpaycloud_volume_attachment.example <attachment-id>
```
