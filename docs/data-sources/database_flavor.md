---
page_title: "vnpaycloud_database_flavor Data Source - VNPayCloud"
subcategory: "Database"
description: |-
  Looks up a single DBaaS compute flavor by ID.
---

# vnpaycloud_database_flavor (Data Source)

Looks up a single DBaaS compute flavor by its ID. To list all flavors, use `vnpaycloud_database_flavors`.

## Example Usage

```hcl
data "vnpaycloud_database_flavor" "standard" {
  id = "fl-xxxxxxxx"
}
```

## Schema

### Required

- `id` (String) — Flavor ID.

### Read-Only

- `name` (String)
- `class` (String)
- `ratio` (String)
- `cpu_req` / `mem_req` (Number)
- `cpu_limit` / `mem_limit` (Number)
