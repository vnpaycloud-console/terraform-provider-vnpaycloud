---
page_title: "VNPayCloud Provider"
subcategory: ""
description: |-
  The VNPayCloud provider is used to interact with VNPay Cloud resources.
---

# VNPayCloud Provider

The VNPayCloud provider is used to manage [VNPay Cloud](https://console.vnpaycloud.vn) resources. The provider needs to be configured with the proper credentials before it can be used.

Use the navigation to the left to read about the available resources and data sources.

## Example Usage

```hcl
terraform {
  required_providers {
    vnpaycloud = {
      source = "terraform-provider-vnpaycloud/vnpaycloud"
    }
  }
}

provider "vnpaycloud" {
  base_url = "https://console.vnpaycloud.vn"
  token    = var.vnpaycloud_token
  zone_id  = "HCMSDN01"
}

# Create a VPC
resource "vnpaycloud_vpc" "example" {
  name = "my-vpc"
  cidr = "10.0.0.0/16"
}
```

## Authentication

The VNPayCloud provider uses a Personal Access Token (PAT) for authentication. You can generate a PAT from the [VNPay Cloud Console](https://console.vnpaycloud.vn).

The token can be provided via:
- The `token` argument in the provider block
- The `VNPAYCLOUD_TOKEN` environment variable

```hcl
provider "vnpaycloud" {
  token = "vtx_pat_XXXXXXXXXXXXXXXXXXXXXXXXXXXX"
}
```

Or using environment variables:

```bash
export VNPAYCLOUD_BASE_URL="https://console.vnpaycloud.vn"
export VNPAYCLOUD_TOKEN="vtx_pat_XXXXXXXXXXXXXXXXXXXXXXXXXXXX"
export VNPAYCLOUD_ZONE_ID="HCMSDN01"
```

## Schema

### Required

- `base_url` (String) The base URL of the VNPay Cloud API. Can also be set with the `VNPAYCLOUD_BASE_URL` environment variable.
- `token` (String, Sensitive) Personal Access Token for authentication. Can also be set with the `VNPAYCLOUD_TOKEN` environment variable.
- `zone_id` (String) The availability zone ID. Can also be set with the `VNPAYCLOUD_ZONE_ID` environment variable.
