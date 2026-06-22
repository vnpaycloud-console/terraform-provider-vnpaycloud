---
page_title: "vnpaycloud_database_redis_sentinel_account Resource - VNPayCloud"
subcategory: "Database"
description: |-
  Manages an account (ACL user) on a VNPayCloud DBaaS Redis Sentinel instance.
---

# vnpaycloud_database_redis_sentinel_account (Resource)

Manages an account (Redis ACL user) on a VNPayCloud DBaaS Redis Sentinel instance.

~> **`password` is write-only** — it is sent on create (and on change) but is never returned by the API. It is kept in Terraform state from your configuration. On `terraform import` you must set it in configuration; the next apply reconciles it.

## Example Usage

```hcl
resource "vnpaycloud_database_redis_sentinel_account" "app" {
  name                       = "appuser"
  redis_sentinel_instance_id = vnpaycloud_database_redis_sentinel_instance.redis.id
  password                   = var.redis_sentinel_account_password
  privilege_template         = "readwrite"
}
```

## Schema

### Required

- `name` (String, Forces new resource) — Account name. Lowercase letters, digits and dots; must start and end with an alphanumeric character. 1–63 characters.
- `redis_sentinel_instance_id` (String, Forces new resource) — ID of the Redis Sentinel instance this account belongs to.
- `password` (String, Sensitive) — Account password (8–128 characters). Updating it issues a change-password call, which re-applies `privilege_template` in the same request.
- `privilege_template` (String) — One of `readonly`, `readwrite`. Changing it alone issues a grant-privilege call; if `password` also changed in the same apply, the privilege is applied by the change-password call instead.

### Read-Only

- `id` (String) — Account ID.
- `status` (String) — Lifecycle status (`active`, `creating`, `error`, `deleting`, `deleted`).
- `created_at` (String) — Creation timestamp (RFC3339).

## Import

```shell
# password is not returned by the API and must be set in configuration after import.
terraform import vnpaycloud_database_redis_sentinel_account.app <account-id>
```
