---
page_title: "vnpaycloud_backup_policy_kubernetes_list Data Source - VNPayCloud"
subcategory: "Backup"
description: |-
  Lists VNPayCloud Kubernetes (PVC) backup policies, optionally filtered by name.
---

# vnpaycloud_backup_policy_kubernetes_list (Data Source)

Lists Kubernetes backup policies in the project's zone. The `name` filter is optional; with it unset, every policy in the zone is returned.

## Example Usage

```hcl
data "vnpaycloud_backup_policy_kubernetes_list" "all" {}

data "vnpaycloud_backup_policy_kubernetes_list" "prod" {
  name = "prod-k8s-daily"
}

output "policy_names" {
  value = [for p in data.vnpaycloud_backup_policy_kubernetes_list.all.backup_policy_kubernetes : p.name]
}
```

## Schema

### Optional

- `name` (String) — Filter by policy name.

### Read-Only

- `backup_policy_kubernetes` (List of Object) — Matching policies. Each element contains:
  - `id` (String) — Policy ID.
  - `name` (String) — Policy name.
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
