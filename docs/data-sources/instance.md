---
page_title: "vnpaycloud_instance Data Source - VNPayCloud"
subcategory: "Compute"
description: |-
  Get information about an instance in VNPayCloud.
---

# vnpaycloud_instance (Data Source)

Use this data source to get information about an existing compute instance (virtual machine), including its configuration, attached volumes, and network interfaces.

## Example Usage

```hcl
data "vnpaycloud_instance" "example" {
  name = "my-web-server"
}

output "instance_status" {
  value = data.vnpaycloud_instance.example.status
}
```

```hcl
data "vnpaycloud_instance" "by_id" {
  id = "srv-pqr77889"
}
```

## Schema

### Optional (filter)

- `id` (String) The ID of the instance.
- `name` (String) The name of the instance.

### Read-Only

- `image_name` (String) The name of the OS image used to boot the instance.
- `image_id` (String) The ID of the OS image used to boot the instance.
- `flavor_name` (String) The name of the compute flavor (e.g., `2c-4g`, `4c-8g`).
- `root_disk_gb` (Number) The size of the root disk in gigabytes (GB).
- `root_disk_type` (String) The storage type of the root disk (e.g., `SSD`, `HDD`).
- `volume_ids` (List of String) A list of additional block volume IDs currently attached to the instance.
- `status` (String) The current status of the instance (e.g., `ACTIVE`, `SHUTOFF`, `ERROR`, `BUILD`).
- `power_state` (String) The power state of the instance (e.g., `running`, `shutdown`, `paused`).
- `network_interface_ids` (List of String) A list of network interface IDs attached to the instance.
- `key_pair` (String) The name of the SSH key pair associated with the instance.
- `security_groups` (List of String) A list of security group names or IDs associated with the instance.
- `server_group_id` (String) The ID of the server group (affinity/anti-affinity group) this instance belongs to, if any.
- `zone_id` (String) The availability zone ID where the instance is deployed.
- `created_at` (String) The timestamp when the instance was created, in ISO 8601 format.
