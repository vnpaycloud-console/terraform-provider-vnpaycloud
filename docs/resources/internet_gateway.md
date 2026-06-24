---
page_title: "vnpaycloud_internet_gateway Resource - VNPayCloud"
subcategory: "Networking"
description: |-
  Manages an internet gateway resource within VNPayCloud.
---

# vnpaycloud_internet_gateway (Resource)

Manages an internet gateway resource within VNPayCloud. An internet gateway enables communication between instances in a VPC and the internet. It can be attached to or detached from a VPC after creation by updating the `vpc_id` field.

~> **Note:** The `name` and `description` fields are immutable. Changing them will force creation of a new internet gateway. The `vpc_id` field is updatable and triggers an attach or detach operation without recreating the resource.

## Example Usage

### Internet Gateway Attached to a VPC

```hcl
resource "vnpaycloud_vpc" "main" {
  name = "my-vpc"
  cidr = "10.0.0.0/16"
}

resource "vnpaycloud_subnet" "app" {
  name   = "app-subnet"
  vpc_id = vnpaycloud_vpc.main.id
  cidr   = "10.0.1.0/24"
}

resource "vnpaycloud_internet_gateway" "example" {
  name        = "my-igw"
  description = "Internet gateway for main VPC"
  vpc_id      = vnpaycloud_vpc.main.id

  depends_on = [vnpaycloud_subnet.app]
}
```

~> **Note:** Attaching an internet gateway requires the VPC to already have at least one network/subnet. If you create the VPC, subnet, and internet gateway in the same configuration, add an explicit dependency from the internet gateway to the subnet.

~> **Note:** A VPC can have only one attached internet gateway. Attaching a second gateway to the same VPC is rejected by the API.

~> **Note:** An internet gateway cannot be detached/destroyed while a floating IP is still associated with its VPC (e.g. a VPC-SNAT floating IP), which fails with `Cannot detach internet gateway from VPC attached floating IP`. Remove or disassociate such floating IPs first — add a `depends_on` from the internet gateway to those `vnpaycloud_floating_ip` resources so Terraform destroys them in the right order.

### Standalone Internet Gateway (Unattached)

```hcl
resource "vnpaycloud_internet_gateway" "standalone" {
  name = "spare-igw"
}
```

## Schema

### Required

- `name` (String, ForceNew) The name of the internet gateway. Length 3–250; may only contain ASCII letters, digits, spaces, and the characters `- _ .`. Changing this creates a new internet gateway.

### Optional

- `description` (String, ForceNew) A description of the internet gateway. Changing this creates a new internet gateway.
- `vpc_id` (String) The ID of the VPC to attach this internet gateway to. This field can be updated to attach or detach the gateway from a VPC without recreating the resource.

### Read-Only

- `id` (String) The ID of the internet gateway.
- `status` (String) The current status of the internet gateway.
- `zone_id` (String) The availability zone ID where the internet gateway is deployed.
- `created_at` (String) The creation timestamp of the internet gateway.

## Timeouts

- `create` - (Default `10 minutes`) Used for creating the internet gateway.
- `delete` - (Default `10 minutes`) Used for deleting the internet gateway.

## Import

Internet gateways can be imported using the `id`:

```shell
terraform import vnpaycloud_internet_gateway.example <internet-gateway-id>
```
