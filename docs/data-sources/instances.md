---
page_title: "vnpaycloud_instances Data Source - VNPayCloud"
subcategory: "Compute"
description: |-
  List all instances in VNPayCloud.
---

# vnpaycloud_instances (Data Source)

Use this data source to list all compute instances in the current project.

## Example Usage

```hcl
data "vnpaycloud_instances" "all" {}

output "all_instance_names" {
  value = data.vnpaycloud_instances.all.instances[*].name
}

output "running_instance_ids" {
  value = [
    for inst in data.vnpaycloud_instances.all.instances :
    inst.id if inst.power_state == "running"
  ]
}

output "instance_flavor_map" {
  value = {
    for inst in data.vnpaycloud_instances.all.instances :
    inst.name => inst.flavor_name
  }
}
```

## Schema

### Read-Only

- `instances` (List of Object) List of instances. Each element contains:
  - `id` (String) The unique identifier of the instance.
  - `name` (String) The name of the instance.
  - `image_name` (String) The name of the image used to create this instance.
  - `flavor_name` (String) The name of the flavor (instance type) of this instance.
  - `status` (String) The current status of the instance (e.g., `active`, `shutoff`, `error`).
  - `power_state` (String) The power state of the instance (e.g., `running`, `shutdown`).
  - `key_pair` (String) The name of the SSH key pair associated with the instance.
  - `server_group_id` (String) The ID of the server group this instance belongs to, if any.
  - `zone_id` (String) The availability zone ID where the instance is deployed.
  - `created_at` (String) The timestamp when the instance was created, in ISO 8601 format.
