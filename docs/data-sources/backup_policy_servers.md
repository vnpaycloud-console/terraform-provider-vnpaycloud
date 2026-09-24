---
page_title: "vnpaycloud_backup_policy_servers Data Source - VNPayCloud"
subcategory: "Backup"
description: |-
  Lists VNPayCloud server backup policies, optionally filtered by name.
---

# vnpaycloud_backup_policy_servers (Data Source)

Lists server backup policies in the project's zone. The `name` filter is optional; with it unset, every policy in the zone is returned.

## Example Usage

```hcl
data "vnpaycloud_backup_policy_servers" "all" {}

data "vnpaycloud_backup_policy_servers" "prod" {
  name = "prod-daily"
}

output "policy_names" {
  value = [for p in data.vnpaycloud_backup_policy_servers.all.backup_policy_servers : p.name]
}
```

## Schema

### Optional

- `name` (String) — Filter by policy name.

### Read-Only

- `backup_policy_servers` (List of Object) — Matching policies. Each element contains:
  - `id` (String) — Policy ID.
  - `name` (String) — Policy name.
  - `description` (String) — Policy description.
  - `start_hour` (Number) — Hour of day the backup run starts.
  - `resource_type` (String) — What the policy protects (`vm`, `volume`).
  - `purpose` (String) — Policy purpose. Policies created through Terraform are always `on_demand`.
  - `is_auto` (Boolean) — Whether the platform picks the run hour instead of using `start_hour`.
  - `is_auto_apply_for_volume` (Boolean) — Whether the policy is applied automatically to newly attached volumes.
  - `backup_vault_ids` (List of String) — Vaults the recovery points are written to.
  - `daily` (List of Object) — Daily schedule. Each element contains:
    - `retentions` (Number) — Daily recovery points kept.
  - `weekly` (List of Object) — Weekly schedule. Empty when disabled. Each element contains:
    - `retentions` (Number) — Weekly recovery points kept.
    - `day_of_week` (Number) — Day the weekly backup runs, 1–7.
  - `monthly` (List of Object) — Monthly schedule. Empty when disabled. Each element contains:
    - `retentions` (Number) — Monthly recovery points kept.
    - `day_type` (String) — How the monthly day is selected: `day_of_month`, or `day_of_1st_week` … `day_of_5th_week`.
    - `day_of_month` (Number) — Day of month, 1–31. Set when `day_type` is `day_of_month`.
    - `day_of_week` (Number) — Day of week, 1–7. Set when `day_type` selects a week.
  - `second_tier_enabled` (Boolean) — Whether second-tier schedules apply.
  - `second_tier_weekly` (List of Object) — Second-tier weekly schedule. Each element contains:
    - `retentions` (Number) — Recovery points kept.
    - `day_of_week` (Number) — Day of week, 0–6.
  - `second_tier_monthly` (List of Object) — Second-tier monthly schedule. Each element contains:
    - `retentions` (Number) — Recovery points kept.
    - `day_type` (String) — Same values as `monthly.day_type`.
    - `day_of_month` (Number) — Day of month, 1–31.
    - `day_of_week` (Number) — Day of week, 0–6.
  - `second_tier_yearly` (List of Object) — Second-tier yearly schedule. Each element contains:
    - `retentions` (Number) — Recovery points kept.
    - `month` (Number) — Month the backup runs, 1–12.
    - `day_of_month` (Number) — Day of month, 1–31.
  - `protected_servers` (Number) — Servers currently covered by the policy.
  - `protected_volumes` (Number) — Volumes currently covered by the policy.
  - `protected_volume_sizes` (Number) — Total size of covered volumes, in GB.
  - `compliance_state` (String) — Whether the policy is meeting its schedule.
  - `compliance_msg` (String) — Detail behind `compliance_state`.
  - `status` (String) — Lifecycle status.
  - `created_at` (String) — Creation timestamp (RFC3339).
