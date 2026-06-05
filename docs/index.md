---
page_title: "VNPayCloud Provider"
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
      source = "vnpaycloud-console/vnpaycloud"
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

## Rate limits

VNPay Cloud applies per-user, per-method rate limits on **every** resource type. Concrete values vary by service and method, but the shape of the policy is the same everywhere:

- Write ops (`create` / `update` / `delete`) are the strictest — typically a few requests per minute per user per resource type, counted independently.
- Read / list ops are more generous but still capped, and reads against a resource that recently exceeded its write quota may also return `Too Many Requests` until the bucket clears.
- Status-transition ops (failover / enable / disable / change-provisioning-status, where applicable) have their own buckets, usually a bit more generous than writes.
- Failed attempts count against the bucket — retrying a `Too Many Requests` response immediately keeps it saturated.
- The provider **retries** both `503 Service Unavailable` (short backoff: 1s, 2s, 4s — max 3 attempts) and `Too Many Requests` / `429` (long backoff: 30s, 60s, 90s, 120s plus jitter — max 4 attempts, spaced out so the rate-limit bucket can clear). If those retries are exhausted the error surfaces — back off before re-running.

**Recommended workflow:** for changesets that touch many resources, run `terraform apply -parallelism=1` to serialize requests; if you hit a `Too Many Requests` error, back off at least one minute before retrying.
