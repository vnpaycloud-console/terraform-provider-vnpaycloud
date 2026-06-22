---
page_title: "vnpaycloud_database_postgres_database Resource - VNPayCloud"
subcategory: "Database"
description: |-
  Manages a PostgreSQL database on a VNPayCloud DBaaS Postgres instance.
---

# vnpaycloud_database_postgres_database (Resource)

Manages a PostgreSQL database (a logical database inside a DBaaS Postgres instance), owned by a Postgres account. Changing the owner in place issues a change-ownership call.

## Example Usage

```hcl
resource "vnpaycloud_database_postgres_account" "owner" {
  name                 = "app_owner"
  postgres_instance_id = vnpaycloud_database_postgres_instance.pg.id
  password             = var.owner_password
}

resource "vnpaycloud_database_postgres_database" "app" {
  name                 = "appdb"
  postgres_instance_id = vnpaycloud_database_postgres_instance.pg.id
  owner                = vnpaycloud_database_postgres_account.owner.name

  # Terminate active connections and force-drop on destroy.
  force_delete = true
}
```

## Schema

### Required

- `name` (String, Forces new resource) — Database name. Lowercase letters, digits and underscores; must not start with a digit. 1–63 characters.
- `postgres_instance_id` (String, Forces new resource) — ID of the Postgres instance.
- `owner` (String) — Database owner role. Updating it issues a change-ownership call.

### Optional

- `force_delete` (Boolean) — When `true`, terminate active connections and force-drop the database on destroy. Defaults to `false`.

### Read-Only

- `id` (String) — Database ID.
- `status` (String) — Lifecycle status (`active`, `creating`, `error`, `deleting`, `deleted`).
- `created_at` (String) — Creation timestamp (RFC3339).

## Import

```shell
terraform import vnpaycloud_database_postgres_database.app <database-id>
```
