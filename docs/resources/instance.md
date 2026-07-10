---
page_title: "vnpaycloud_instance Resource - VNPayCloud"
subcategory: "Compute"
description: |-
  Manages a compute instance within VNPayCloud.
---

# vnpaycloud_instance (Resource)

Manages a compute instance (virtual machine) within VNPayCloud. Instances can be booted from an image and can be resized between named flavors.

## Example Usage

```hcl
resource "vnpaycloud_keypair" "deployer" {
  name = "deployer-key"
}

resource "vnpaycloud_instance" "web" {
  name               = "web-server-01"
  image              = "Ubuntu 22.04 LTS"
  flavor             = "a-pro-small.2x2"
  root_disk_gb       = 40
  root_disk_type     = "c1-standard"
  key_pair           = vnpaycloud_keypair.deployer.name
  network_interface_ids = ["nic-abc123"]

  user_data = <<-EOF
    #!/bin/bash
    apt-get update -y
    apt-get install -y nginx
  EOF
}
```

~> **Note:** `network_interface_ids` defines the interfaces attached **at launch** and is immutable — changing it forces the instance to be recreated. To attach or detach a network interface on an existing instance without recreating it, use the separate [`vnpaycloud_network_interface_attachment`](network_interface_attachment.md) resource. The provider does not refresh `network_interface_ids` from the live instance, so interfaces added later via `vnpaycloud_network_interface_attachment` do not cause drift here.

!> **Warning:** Booting an instance from `snapshot_id` is not supported yet. Use `image` when creating `vnpaycloud_instance`. If `snapshot_id` is set, the API rejects the create request with `booting an instance from a snapshot is not supported yet; use image`.

## Schema

### Required

- `name` (String) The name of the instance. Allowed characters are letters, numbers, hyphen (`-`), underscore (`_`), dot (`.`), and space.
- `root_disk_gb` (Number, ForceNew) The size of the root disk in gigabytes. Minimum value is `20`. Changing this creates a new instance.
- `root_disk_type` (String, ForceNew) The root disk volume type name (e.g., `c1-standard`). Use the `vnpaycloud_volume_types` data source to list available values. Changing this creates a new instance.

### Optional

- `image` (String, ForceNew) The image name to boot the instance from. Use this field when creating an instance. Changing this creates a new instance.
- `snapshot_id` (String, ForceNew) Reserved for future support. Booting from a snapshot is not supported yet; setting this field causes create to fail. Use `image` instead.
- `flavor` (String) The flavor name defining the vCPU and RAM resources for the instance (e.g., `a-pro-small.2x2`). Mutually exclusive with `is_custom_flavor`.
- `is_custom_flavor` (Boolean) Reserved for future use. Custom flavor create/resize is not currently supported by the Terraform provider; use `flavor` with a named flavor.
- `custom_vcpus` (Number) Reserved for future use.
- `custom_ram_mb` (Number) Reserved for future use.
- `key_pair` (String, ForceNew, Computed) The name of the SSH key pair to inject into the instance. Changing this creates a new instance. If not specified and the image supports it, a key pair may be computed.
- `network_interface_ids` (List of String, ForceNew) The network interface IDs to attach to the instance **at launch**. At least one interface is required. Changing this list creates a new instance — to attach or detach interfaces on a running instance, use [`vnpaycloud_network_interface_attachment`](network_interface_attachment.md) instead.
- `server_group_id` (String, ForceNew) The ID of the server group to place the instance in. Changing this creates a new instance.
- `user_data` (String, ForceNew, Sensitive) User data script to pass to the instance at boot time. Changing this creates a new instance.
- `is_user_data_base64` (Boolean, ForceNew) Set to `true` if the `user_data` value is already Base64-encoded. Changing this creates a new instance.

### Read-Only

- `id` (String) The ID of the instance.
- `image_name` (String) The name of the image used to boot the instance.
- `image_id` (String) The ID of the image used to boot the instance.
- `flavor_name` (String) The resolved flavor name of the instance.
- `security_groups` (List of String) The security groups currently effective on the instance. Read-only — security groups are managed per network interface (IP), not on the instance.
- `volume_ids` (List of String) List of volume IDs attached to the instance.
- `status` (String) The current status of the instance (e.g., `active`, `shutoff`, `error`).
- `power_state` (String) The current power state of the instance (e.g., `running`, `shutdown`).
- `zone_id` (String) The availability zone ID where the instance is deployed.
- `created_at` (String) The creation timestamp of the instance in ISO 8601 format.

## Timeouts

- `create` - (Default `30 minutes`) Used for creating the instance.
- `update` - (Default `30 minutes`) Used for updating the instance (e.g., renaming, resizing the flavor).
- `delete` - (Default `10 minutes`) Used for deleting the instance.

## Import

Instances can be imported using the `id`:

```shell
terraform import vnpaycloud_instance.example <instance-id>
```

After import, the provider reads back `image`, `flavor`, `root_disk_gb`, and `root_disk_type` from the live instance, so these match your configuration without forcing a replacement.

~> **Note:** `user_data` is a write-only field and is not read back from the instance. If you declare `user_data` in the configuration of an imported instance, the next plan will show a change on this `ForceNew` attribute and recreate the instance. Leave `user_data` unset on imported instances.

!> **Warning:** Do **not** declare `network_interface_ids` in the configuration of an imported instance. On import the provider reads back **all** interfaces currently on the instance — both the launch interfaces and any added later via [`vnpaycloud_network_interface_attachment`](network_interface_attachment.md) — and it cannot tell them apart. If your configuration lists only a subset (for example just the launch interface), the next plan will see a mismatch on this `ForceNew` attribute and **destroy and recreate the instance**. Leave `network_interface_ids` unset on imported instances and manage post-launch interfaces with `vnpaycloud_network_interface_attachment`.
