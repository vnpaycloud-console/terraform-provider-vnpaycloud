---
page_title: "vnpaycloud_backup_policy_kubernetes Data Source - VNPayCloud"
subcategory: "Backup"
description: |-
  Looks up a single VNPayCloud Kubernetes (PVC) backup policy by ID or name.
---

# vnpaycloud_backup_policy_kubernetes (Data Source)

Looks up a single Kubernetes backup policy. Specify at least one of `id` or `name`. When both are set, the policy is fetched by `id` and the lookup fails if its name does not match.

## Example Usage

```hcl
data "vnpaycloud_backup_policy_kubernetes" "by_name" {
  name = "prod-k8s-daily"
}

data "vnpaycloud_backup_policy_kubernetes" "by_id" {
  id = "bpk-0a1b2c3d"
}
```

## Schema

### Optional

- `id` (String) — Policy ID. At least one of `id` or `name` must be set.
- `name` (String) — Policy name. Must match exactly one policy. At least one of `id` or `name` must be set.

### Read-Only

- `description` (String) — Policy description.
- `auto_apply_new_volume` (Boolean) — Whether the policy is applied automatically to new persistent volumes in the cluster.
- `is_auto` (Boolean) — Whether the platform picks the run hour instead of using `start_hour`.
- `start_hour` (Number) — Hour of day the backup run starts.
- `run_priority` (Number) — Ordering priority when multiple policies run.
- `purpose` (String) — Policy purpose. Policies created through Terraform are always `on_demand`.
- `backup_vault_ids` (List of String) — Vault the recovery points are written to.
- `daily` (List of Object) — Daily schedule. Empty when disabled. Each element contains:
  - `retentions` (Number) — Daily recovery points kept.
- `weekly` (List of Object) — Weekly schedule. Empty when disabled. Each element contains:
  - `retentions` (Number) — Weekly recovery points kept.
  - `day_of_week` (Number) — Day the weekly backup runs.
- `monthly` (List of Object) — Monthly schedule. Empty when disabled. Each element contains:
  - `retentions` (Number) — Monthly recovery points kept.
  - `day_type` (String) — How the monthly day is selected.
  - `day_of_month` (Number) — Day of month. Set when `day_type` is `day_of_month`.
  - `day_of_week` (Number) — Day of week. Set when `day_type` selects a week.
- `status` (String) — Lifecycle status.
- `created_at` (String) — Creation timestamp (RFC3339).
