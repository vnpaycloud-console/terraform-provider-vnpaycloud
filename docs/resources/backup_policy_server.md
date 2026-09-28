---
page_title: "vnpaycloud_backup_policy_server Resource - VNPayCloud"
subcategory: "Backup"
description: |-
  Manages a VNPayCloud server backup policy — the schedule and retention rules applied to servers or volumes.
---

# vnpaycloud_backup_policy_server (Resource)

Manages a server backup policy. A policy defines when backups run and how many recovery points are kept, and writes those recovery points into one or more backup vaults.

A daily schedule is always required. Weekly and monthly schedules are optional; omit the block to disable that tier. Second-tier schedules are only sent when `second_tier_enabled` is `true`.

## Example Usage

```hcl
data "vnpaycloud_backup_vault" "servers" {
  name = "prod-server-vault"
}

resource "vnpaycloud_backup_policy_server" "daily" {
  name          = "prod-daily"
  description   = "Daily backups for production servers"
  start_hour    = 2
  resource_type = "vm"
  is_auto       = true

  backup_vault_ids = [data.vnpaycloud_backup_vault.servers.id]

  daily {
    retentions = 7
  }
}
```

### With weekly and monthly tiers

```hcl
resource "vnpaycloud_backup_policy_server" "layered" {
  name          = "prod-layered"
  start_hour    = 1
  resource_type = "vm"

  backup_vault_ids = [data.vnpaycloud_backup_vault.servers.id]

  daily {
    retentions = 14
  }

  weekly {
    retentions  = 4
    day_of_week = 1
  }

  monthly {
    retentions   = 6
    day_type     = "day_of_month"
    day_of_month = 1
  }
}
```

### With a second tier

```hcl
resource "vnpaycloud_backup_policy_server" "archival" {
  name          = "prod-archival"
  start_hour    = 3
  resource_type = "volume"

  backup_vault_ids = [data.vnpaycloud_backup_vault.servers.id]

  daily {
    retentions = 7
  }

  second_tier_enabled = true

  second_tier_weekly {
    retentions  = 4
    day_of_week = 0
  }

  second_tier_yearly {
    retentions   = 2
    month        = 12
    day_of_month = 31
  }
}
```

## Schema

### Required

- `name` (String, Forces new resource) — Policy name. Letters, digits, spaces, hyphens, underscores and dots only. 3–512 characters (validated at plan time).
- `start_hour` (Number) — Hour of day the backup run starts, 0–23 (validated at plan time). Can be updated in place.
- `resource_type` (String, Forces new resource) — What the policy protects. One of `vm`, `volume` (validated at plan time).
- `backup_vault_ids` (List of String, Forces new resource) — Vaults the recovery points are written to. At least one is required (validated at plan time): one ceph vault, optionally a second s3 vault for the second tier.
- `daily` (Block, max 1) — Daily schedule. Always required.
  - `retentions` (Number) — Number of daily recovery points to keep, 7–14 (validated at plan time).

### Optional

- `description` (String) — Free-form description, up to 512 characters (validated at plan time). Can be updated in place.
- `is_auto` (Boolean) — Who decides when the run starts. Defaults to `false`, which runs it at `start_hour`. Set it to `true` to hand the choice to the platform, which spreads runs across its off-peak window; `start_hour` is then ignored.
- `is_auto_apply_for_volume` (Boolean) — Whether the policy is applied automatically to newly attached volumes. Defaults to `false`.
- `weekly` (Block, max 1) — Weekly schedule. Omit to disable.
  - `retentions` (Number) — Weekly recovery points to keep, 1–14 (validated at plan time).
  - `day_of_week` (Number) — Day the weekly backup runs, 1–7 (validated at plan time); `1` = Sunday … `7` = Saturday.
- `monthly` (Block, max 1) — Monthly schedule. Omit to disable.
  - `retentions` (Number) — Monthly recovery points to keep, 1–14 (validated at plan time).
  - `day_type` (String) — How the monthly day is selected. One of `day_of_month`, `day_of_1st_week`, `day_of_2nd_week`, `day_of_3rd_week`, `day_of_4th_week`, `day_of_5th_week` (validated at plan time).
  - `day_of_month` (Number) — Day of month, 1–31 (validated at plan time). **Required when `day_type` is `day_of_month`**, and must be omitted otherwise.
  - `day_of_week` (Number) — Day of week, 1–7 (validated at plan time); `1` = Sunday … `7` = Saturday. **Required when `day_type` selects a week** (`day_of_1st_week` … `day_of_5th_week`), and must be omitted when `day_type` is `day_of_month`.
- `second_tier_enabled` (Boolean) — Whether second-tier schedules apply. Defaults to `false`. Second-tier blocks are only sent when this is `true`. The second tier writes recovery points to the s3 vault listed in `backup_vault_ids`, on its own schedule and retention. Its `day_of_week` fields count from `0` (Sunday) rather than `1` like the first-tier blocks above.
- `second_tier_weekly` (Block, max 1) — Second-tier weekly schedule.
  - `retentions` (Number) — Recovery points to keep, 1–14 (validated at plan time).
  - `day_of_week` (Number) — Day of week, 0–6 (validated at plan time); `0` = Sunday … `6` = Saturday.
- `second_tier_monthly` (Block, max 1) — Second-tier monthly schedule.
  - `retentions` (Number) — Recovery points to keep, 1–14 (validated at plan time).
  - `day_type` (String) — Same values as `monthly.day_type` (validated at plan time).
  - `day_of_month` (Number) — Day of month, 1–31 (validated at plan time).
  - `day_of_week` (Number) — Day of week, 0–6 (validated at plan time); `0` = Sunday … `6` = Saturday.
- `second_tier_yearly` (Block, max 1) — Second-tier yearly schedule.
  - `retentions` (Number) — Recovery points to keep, 1–14 (validated at plan time).
  - `month` (Number) — Month the backup runs, 1–12 (validated at plan time).
  - `day_of_month` (Number) — Day of month, 1–31 (validated at plan time).

### Read-Only

- `id` (String) — Policy ID.
- `purpose` (String) — Policy purpose. Set by the platform; policies created through Terraform are always `on_demand`.
- `protected_servers` (Number) — Servers currently covered by the policy.
- `protected_volumes` (Number) — Volumes currently covered by the policy.
- `protected_volume_sizes` (Number) — Total size of covered volumes, in GB.
- `compliance_state` (String) — Whether the policy is meeting its schedule.
- `compliance_msg` (String) — Detail behind `compliance_state`.
- `status` (String) — Lifecycle status (`active`, `creating`, `error`, `deleting`, `deleted`).
- `created_at` (String) — Creation timestamp (RFC3339).

## Timeouts

- `create` — default 10 minutes.
- `delete` — default 10 minutes.
## Import

```shell
terraform import vnpaycloud_backup_policy_server.daily <policy-id>
```
