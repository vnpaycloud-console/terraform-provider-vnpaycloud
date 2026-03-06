---
page_title: "vnpaycloud_instance Resource - VNPayCloud"
subcategory: "Compute"
description: |-
  Manages a compute instance within VNPayCloud.
---

# vnpaycloud_instance (Resource)

Manages a compute instance (virtual machine) within VNPayCloud. Instances can be booted from an image or a snapshot and support flexible flavor configurations including custom vCPU and RAM sizing.

## Example Usage

```hcl
resource "vnpaycloud_keypair" "deployer" {
  name = "deployer-key"
}

resource "vnpaycloud_instance" "web" {
  name               = "web-server-01"
  image              = "ubuntu-22.04"
  flavor             = "s.4c8r"
  root_disk_gb       = 40
  root_disk_type     = "SSD"
  key_pair           = vnpaycloud_keypair.deployer.name
  security_groups    = ["default", "web-sg"]
  network_interface_ids = ["nic-abc123"]

  user_data = <<-EOF
    #!/bin/bash
    apt-get update -y
    apt-get install -y nginx
  EOF
}
```

## Schema

### Required

- `name` (String) The name of the instance.
- `root_disk_gb` (Number, ForceNew) The size of the root disk in gigabytes. Changing this creates a new instance.
- `root_disk_type` (String, ForceNew) The type of the root disk (e.g., `SSD`, `HDD`). Changing this creates a new instance.

### Optional

- `image` (String, ForceNew) The image name or ID to boot the instance from. Conflicts with `snapshot_id`. Changing this creates a new instance.
- `snapshot_id` (String, ForceNew) The ID of a volume snapshot to boot the instance from. Conflicts with `image`. Changing this creates a new instance.
- `flavor` (String) The flavor name defining the vCPU and RAM resources for the instance (e.g., `s.4c8r`). Mutually exclusive with `is_custom_flavor`.
- `is_custom_flavor` (Boolean) Set to `true` to use custom vCPU and RAM values instead of a named flavor. When enabled, `custom_vcpus` and `custom_ram_mb` must be provided.
- `custom_vcpus` (Number) Number of vCPUs for the instance when using a custom flavor. Required when `is_custom_flavor` is `true`.
- `custom_ram_mb` (Number) Amount of RAM in megabytes for the instance when using a custom flavor. Required when `is_custom_flavor` is `true`.
- `key_pair` (String, ForceNew, Computed) The name of the SSH key pair to inject into the instance. Changing this creates a new instance. If not specified and the image supports it, a key pair may be computed.
- `security_groups` (List of String) A list of security group names to associate with the instance.
- `network_interface_ids` (List of String) A list of network interface IDs to attach to the instance.
- `server_group_id` (String, ForceNew) The ID of the server group to place the instance in. Changing this creates a new instance.
- `user_data` (String, ForceNew, Sensitive) User data script to pass to the instance at boot time. Changing this creates a new instance.
- `is_user_data_base64` (Boolean, ForceNew) Set to `true` if the `user_data` value is already Base64-encoded. Changing this creates a new instance.

### Read-Only

- `id` (String) The ID of the instance.
- `image_name` (String) The name of the image used to boot the instance.
- `image_id` (String) The ID of the image used to boot the instance.
- `flavor_name` (String) The resolved flavor name of the instance.
- `volume_ids` (List of String) List of volume IDs attached to the instance.
- `status` (String) The current status of the instance (e.g., `ACTIVE`, `SHUTOFF`, `ERROR`).
- `power_state` (String) The current power state of the instance (e.g., `running`, `stopped`).
- `zone_id` (String) The availability zone ID where the instance is deployed.
- `created_at` (String) The creation timestamp of the instance in ISO 8601 format.

## Timeouts

- `create` - (Default `30 minutes`) Used for creating the instance.
- `update` - (Default `30 minutes`) Used for updating the instance (e.g., resizing, changing security groups).
- `delete` - (Default `10 minutes`) Used for deleting the instance.

## Import

Instances can be imported using the `id`:

```shell
terraform import vnpaycloud_instance.example <instance-id>
```
