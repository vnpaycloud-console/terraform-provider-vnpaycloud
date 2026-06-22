---
page_title: "vnpaycloud_database_redis_versions Data Source - VNPayCloud"
subcategory: "Database"
description: |-
  Lists the Redis engine versions supported by DBaaS.
---

# vnpaycloud_database_redis_versions (Data Source)

Lists the Redis engine versions supported for `vnpaycloud_database_redis_instance` and `vnpaycloud_database_redis_sentinel_instance` (includes Valkey builds).

## Example Usage

```hcl
data "vnpaycloud_database_redis_versions" "all" {}

output "redis_versions" {
  value = data.vnpaycloud_database_redis_versions.all.versions
}
```

## Schema

### Read-Only

- `versions` (List of String) — Supported Redis/Valkey versions (e.g. `7.4.1`, `valkey-8.1.1`).
