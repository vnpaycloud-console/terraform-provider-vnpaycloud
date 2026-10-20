---
page_title: "vnpaycloud_internet_gateway_acl Resource - VNPayCloud"
subcategory: "Networking"
description: |-
  Manages the ACL (CIDR blacklist) of an internet gateway within VNPayCloud.
---

# vnpaycloud_internet_gateway_acl (Resource)

Manages the ACL of an internet gateway within VNPayCloud. The ACL is a blacklist of public IPv4 CIDR blocks: traffic to and from any address inside those blocks is dropped at the gateway.

An internet gateway has exactly one ACL, so at most one `vnpaycloud_internet_gateway_acl` resource may target a given `internet_gateway_id`; a second one pointing at the same gateway is rejected. The `match` list is applied as a whole — every apply replaces the entire blacklist with the list in the configuration.

~> **Note:** The internet gateway must be active before its ACL can be set. When both are created in the same configuration, reference the gateway through `vnpaycloud_internet_gateway.<name>.id` so Terraform orders them correctly.

~> **Note:** Creating this resource on an internet gateway that already has a blacklist is rejected, so an existing blacklist is never silently replaced. Import it instead, or clear it before applying.

~> **Note:** Destroying this resource empties the blacklist on the gateway; the gateway itself is left untouched.

## Example Usage

### Blacklist a Few Public Ranges

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

resource "vnpaycloud_internet_gateway" "main" {
  name   = "my-igw"
  vpc_id = vnpaycloud_vpc.main.id

  depends_on = [vnpaycloud_subnet.app]
}

resource "vnpaycloud_internet_gateway_acl" "main" {
  internet_gateway_id = vnpaycloud_internet_gateway.main.id

  match = [
    "203.0.0.0/24",
    "8.8.8.8/32",
  ]
}
```

### Blacklist Built From a Variable

```hcl
variable "blocked_cidrs" {
  type    = list(string)
  default = ["45.33.32.0/24", "198.41.0.4/32"]
}

resource "vnpaycloud_internet_gateway_acl" "from_var" {
  internet_gateway_id = vnpaycloud_internet_gateway.main.id
  match               = var.blocked_cidrs
}
```

## Schema

### Required

- `internet_gateway_id` (String, ForceNew) The ID of the internet gateway whose ACL is managed. Changing this creates a new ACL on the new gateway and clears the one on the old gateway.
- `match` (Set of String) The blacklisted CIDR blocks. At least one entry is required. Each entry must be a public IPv4 CIDR block written in full CIDR notation — a bare address such as `8.8.8.8` is rejected, write `8.8.8.8/32` instead. Duplicate entries are collapsed into one. The number of entries is limited by the internet gateway CIDR blacklist quota of the project.

The following ranges are reserved and are rejected at plan time: `0.0.0.0/8`, `10.0.0.0/8`, `127.0.0.0/8`, `169.254.0.0/16`, `172.16.0.0/12`, `192.0.0.0/24`, `192.0.2.0/24`, `192.88.99.0/24`, `192.168.0.0/16`, `198.18.0.0/15`, `198.51.100.0/24`, `203.0.113.0/24`, `224.0.0.0/4`, `240.0.0.0/4`, `255.255.255.255/32`.

### Read-Only

- `id` (String) The ID of the internet gateway the ACL belongs to.
- `zone_id` (String) The availability zone ID where the internet gateway is deployed.

## Timeouts

- `create` - (Default `10 minutes`) Used for applying the ACL.
- `update` - (Default `10 minutes`) Used for replacing the ACL.
- `delete` - (Default `10 minutes`) Used for clearing the ACL.

## Import

An internet gateway ACL can be imported using the ID of its internet gateway:

```shell
terraform import vnpaycloud_internet_gateway_acl.main <internet-gateway-id>
```
