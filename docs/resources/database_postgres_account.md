---
page_title: "vnpaycloud_database_postgres_account Resource - VNPayCloud"
subcategory: "Database"
description: |-
  Manages a PostgreSQL account (database role) on a VNPayCloud DBaaS Postgres instance, including its privilege grants.
---

# vnpaycloud_database_postgres_account (Resource)

Manages a PostgreSQL account (database role) on a VNPayCloud DBaaS Postgres instance. Privileges are granted per database/schema through inline `grant` blocks; the provider reconciles them by issuing grant/revoke calls for the difference on update.

~> **`password` is write-only** — it is sent on create (and on change) but is never returned by the API. It is kept in Terraform state from your configuration. On `terraform import` you must set it in configuration; the next apply reconciles it.

## Example Usage

```hcl
resource "vnpaycloud_database_postgres_account" "app" {
  name                 = "app_user"
  postgres_instance_id = vnpaycloud_database_postgres_instance.pg.id
  password             = var.app_db_password

  grant {
    db_name   = "appdb"
    db_schema = "public"
    privilege = "readwrite"
  }

  grant {
    db_name   = "reportdb"
    db_schema = "public"
    privilege = "readonly"
  }
}
```

## Schema

### Required

- `name` (String, Forces new resource) — Role name. Lowercase letters, digits and underscores; must not start with a digit. 1–63 characters.
- `postgres_instance_id` (String, Forces new resource) — ID of the Postgres instance this account belongs to.
- `password` (String, Sensitive) — Account password (8–128 characters). Updating it issues a change-password call.

### Optional

- `grant` (Block Set) — Privileges granted to the account. Each block:
  - `db_name` (String, Required) — Target database.
  - `db_schema` (String, Required) — Target schema.
  - `privilege` (String, Required) — One of `readonly`, `readwrite`.

### Read-Only

- `id` (String) — Account ID.
- `status` (String) — Lifecycle status (`active`, `creating`, `error`, `deleting`, `deleted`).
- `created_at` (String) — Creation timestamp (RFC3339).

## Deletion

An account that **owns a database** cannot be deleted — the API rejects it with `cannot delete account '<name>' because it is the owner of database '<db>', please transfer ownership first`. Reassign the database's `owner` to another role (see `vnpaycloud_database_postgres_database`) before destroying the account. When the account owns a database that is itself managed by Terraform, add a `depends_on` so the database is destroyed first, or move ownership off the account in a prior apply.

## Import

```shell
# password is not returned by the API and must be set in configuration after import.
terraform import vnpaycloud_database_postgres_account.app <account-id>
```
