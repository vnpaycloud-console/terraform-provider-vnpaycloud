---
page_title: "vnpaycloud_database_postgres_versions Data Source - VNPayCloud"
subcategory: "Database"
description: |-
  Lists the PostgreSQL engine versions supported by DBaaS.
---

# vnpaycloud_database_postgres_versions (Data Source)

Lists the PostgreSQL engine versions supported for `vnpaycloud_database_postgres_instance`.

## Example Usage

```hcl
data "vnpaycloud_database_postgres_versions" "all" {}

output "latest_pg_version" {
  value = data.vnpaycloud_database_postgres_versions.all.versions[length(data.vnpaycloud_database_postgres_versions.all.versions) - 1]
}
```

## Schema

### Read-Only

- `versions` (List of String) — Supported PostgreSQL versions (e.g. `15.13`, `16.9`, `17.5`).
