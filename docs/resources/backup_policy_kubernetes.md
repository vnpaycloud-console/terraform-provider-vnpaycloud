---
page_title: "vnpaycloud_backup_policy_kubernetes Resource - VNPayCloud"
subcategory: "Backup"
description: |-
  Manages a VNPayCloud Kubernetes (PVC) backup policy — the schedule and retention rules applied to persistent volumes in a workload cluster.
---

# vnpaycloud_backup_policy_kubernetes (Resource)

Manages a Kubernetes backup policy. A policy defines when persistent-volume backups run and how many recovery points are kept, and writes those recovery points into a workload-cluster backup vault.

At least one of `daily`, `weekly` or `monthly` must be present — a policy with no schedule is rejected. Omit a block to disable that tier.

The policy is written to exactly one backup vault, which must have `purpose = "workload_cluster"` and `type = "s3"`.

## Example Usage

```hcl
resource "vnpaycloud_backup_vault" "workload" {
  name    = "prod-k8s-vault"
  purpose = "workload_cluster"
  type    = "s3"
}

resource "vnpaycloud_backup_policy_kubernetes" "daily" {
  name        = "prod-k8s-daily"
  description = "Daily PVC backups for the production cluster"
  is_auto     = true

  auto_apply_new_volume = true

  backup_vault_ids = [vnpaycloud_backup_vault.workload.id]

  daily {
    retentions = 7
  }
}
```

### With weekly and monthly tiers

```hcl
resource "vnpaycloud_backup_policy_kubernetes" "layered" {
  name       = "prod-k8s-layered"
  start_hour = 1

  backup_vault_ids = [vnpaycloud_backup_vault.workload.id]

  daily {
    retentions = 14
  }

  weekly {
    retentions  = 4
    day_of_week = 1
  }

  monthly {
    retentions   = 6
    day_type     = "fixed_day"
    day_of_month = 1
  }
}
```

### Monthly on the third Monday

```hcl
resource "vnpaycloud_backup_policy_kubernetes" "monthly" {
  name       = "prod-k8s-monthly"
  start_hour = 3

  backup_vault_ids = [vnpaycloud_backup_vault.workload.id]

  monthly {
    retentions  = 3
    day_type    = "third_week"
    day_of_week = 1
  }
}
```

## Schema

### Required

- `name` (String, Forces new resource) — Policy name. Letters, digits, hyphens, underscores and dots only (no spaces). 3–50 characters (validated at plan time).
- `backup_vault_ids` (List of String, Forces new resource) — Exactly one workload-cluster S3 vault the recovery points are written to.

### Optional

At least one schedule block (`daily`, `weekly` or `monthly`) is required.

- `description` (String) — Free-form description, up to 512 characters (validated at plan time). Can be updated in place.
- `auto_apply_new_volume` (Boolean) — Whether the policy is applied automatically to new persistent volumes in the cluster. Defaults to `false`. Can be updated in place.
- `is_auto` (Boolean) — Who decides when the run starts. Defaults to `false`, which runs it at `start_hour`. Set it to `true` to hand the choice to the platform, which spreads runs across its off-peak window; `start_hour` is then forced to `0`. Can be updated in place.
- `start_hour` (Number, Computed) — Hour of day the backup run starts, 0–23 (validated at plan time). Computed when omitted. Can be updated in place. Only takes effect while `is_auto` is `false`; with `is_auto = true` the platform forces it to `0`.
- `run_priority` (Number) — Ordering priority when multiple policies run. Computed when omitted. Can be updated in place.
- `daily` (Block, max 1) — Daily schedule. Omit to disable.
  - `retentions` (Number) — Number of daily recovery points to keep, 7–14 (validated at plan time).
- `weekly` (Block, max 1) — Weekly schedule. Omit to disable.
  - `retentions` (Number) — Weekly recovery points to keep, 1–14 (validated at plan time).
  - `day_of_week` (Number) — Day the weekly backup runs, 0–6 (validated at plan time); `0` = Sunday … `6` = Saturday.
- `monthly` (Block, max 1) — Monthly schedule. Omit to disable.
  - `retentions` (Number) — Monthly recovery points to keep, 1–14 (validated at plan time).
  - `day_type` (String) — How the monthly day is selected. One of `fixed_day`, `first_week`, `second_week`, `third_week`, `fourth_week`, `fifth_week` (validated at plan time).
  - `day_of_month` (Number) — Day of month, 1–31 (validated at plan time). **Required when `day_type` is `fixed_day`**, and must be omitted otherwise.
  - `day_of_week` (Number) — Day of week, 0–6 (validated at plan time; `0` = Sunday). Used when `day_type` selects a week (`first_week` … `fifth_week`); must be omitted when `day_type` is `fixed_day`.

### Read-Only

- `id` (String) — Policy ID.
- `purpose` (String) — Policy purpose. Set by the platform; policies created through Terraform are always `on_demand`.
- `status` (String) — Lifecycle status (`active`, `creating`, `error`, `deleting`, `deleted`).
- `created_at` (String) — Creation timestamp (RFC3339).

## Timeouts

- `create` — default 10 minutes.
- `delete` — default 10 minutes.
## Import

```shell
terraform import vnpaycloud_backup_policy_kubernetes.daily <policy-id>
```
