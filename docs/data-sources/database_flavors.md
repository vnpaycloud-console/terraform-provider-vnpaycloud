---
page_title: "vnpaycloud_database_flavors Data Source - VNPayCloud"
subcategory: "Database"
description: |-
  Lists the DBaaS compute flavors available for database instances.
---

# vnpaycloud_database_flavors (Data Source)

Lists the DBaaS compute flavors (CPU/memory presets) available for PostgreSQL and Redis instances. Use a flavor's `id` for the `flavor_database_id` argument of a database instance.

## Example Usage

```hcl
data "vnpaycloud_database_flavors" "all" {}

resource "vnpaycloud_database_postgres_instance" "main" {
  flavor_database_id = data.vnpaycloud_database_flavors.all.flavors[0].id
  # ...
}
```

## Schema

### Read-Only

- `flavors` (List of Object) — Available flavors, each with:
  - `id` (String)
  - `name` (String)
  - `class` (String)
  - `ratio` (String)
  - `cpu_req` / `mem_req` (Number) — Requested CPU (millicores) / memory (MiB).
  - `cpu_limit` / `mem_limit` (Number) — CPU/memory limits.
