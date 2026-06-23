---
page_title: "vnpaycloud_vpn_gateway_vpc_attachment Resource - VNPayCloud"
subcategory: "Networking"
description: |-
  Attaches a VPC to a VNPayCloud VPN gateway.
---

# vnpaycloud_vpn_gateway_vpc_attachment (Resource)

Attaches a VPC to a VPN gateway within VNPayCloud. This resource manages the attachment lifecycle - creating it attaches the VPC to the VPN gateway, destroying it detaches it.

~> **Note:** Both `vpn_gateway_id` and `vpc_id` are immutable; changing either will force creation of a new attachment.

Do not manage the same VPN gateway/VPC pair with more than one `vnpaycloud_vpn_gateway_vpc_attachment` resource instance.

## Example Usage

The `vpc_id` can reference either an existing VPC ID or a VPC created in the same Terraform configuration. The example below creates a new VPC and attaches it to a VPN gateway.

```hcl
resource "vnpaycloud_vpc" "main" {
  name = "tf-vpn-vpc"
  cidr = "10.0.0.0/16"
}

resource "vnpaycloud_vpn_gateway" "main" {
  name     = "tf-vpngw"
  vpn_type = "ROUTE_BASED"
}

resource "vnpaycloud_vpn_gateway_vpc_attachment" "main" {
  vpn_gateway_id = vnpaycloud_vpn_gateway.main.id
  vpc_id         = vnpaycloud_vpc.main.id
}
```

## Schema

### Required

- `vpn_gateway_id` (String, ForceNew) The ID of the VPN gateway to attach. Changing this creates a new attachment.
- `vpc_id` (String, ForceNew) The ID of the VPC to attach to the VPN gateway. Changing this creates a new attachment.

## Timeouts

- `create` - (Default `10 minutes`) Used for attaching the VPC.
- `delete` - (Default `10 minutes`) Used for detaching the VPC.

## Import

VPN gateway VPC attachments can be imported with a composite ID:

```shell
terraform import vnpaycloud_vpn_gateway_vpc_attachment.example <vpn-gateway-id>/<vpc-id>
```
