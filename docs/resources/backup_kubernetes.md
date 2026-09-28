---
page_title: "vnpaycloud_backup_kubernetes Resource - VNPayCloud"
subcategory: "Backup"
description: |-
  Manages a VNPayCloud Kubernetes cluster backup — puts a workload cluster under a Kubernetes backup policy.
---

# vnpaycloud_backup_kubernetes (Resource)

Manages a Kubernetes cluster backup: it puts one workload cluster under a Kubernetes backup policy. The policy decides when backups run and how long recovery points are kept; this resource decides *which cluster* is protected. The whole cluster is protected.

A cluster can be protected by only one backup at a time. Changing the protected cluster or the policy replaces the resource; the description can be changed in place.

## Example Usage

```hcl
resource "vnpaycloud_backup_vault" "workload" {
  name    = "prod-workload-vault"
  purpose = "workload_cluster"
  type    = "s3"
}

resource "vnpaycloud_backup_policy_kubernetes" "daily" {
  name       = "prod-k8s-daily"
  is_auto    = true

  backup_vault_ids = [vnpaycloud_backup_vault.workload.id]

  daily {
    retentions = 7
  }
}

resource "vnpaycloud_backup_kubernetes" "prod" {
  protected_cluster_id = vnpaycloud_kubernetes_cluster.prod.id
  pvc_backup_policy_id = vnpaycloud_backup_policy_kubernetes.daily.id
  description          = "Full backup of the production cluster"
}
```

## Schema

### Required

- `protected_cluster_id` (String, Forces new resource) — ID of the workload cluster to protect. 3–100 characters (validated at plan time).
- `pvc_backup_policy_id` (String, Forces new resource) — ID of the Kubernetes backup policy that governs the schedule and retention. 3–100 characters (validated at plan time).

### Optional

- `description` (String) — Free-text description. Up to 512 characters. Can be updated in place.

### Read-Only

- `id` (String) — Cluster backup ID.
- `name` (String) — Backup name (derived from the protected cluster).
- `cluster_backup_type` (String) — Backup type. Always `on_demand` when created through Terraform.
- `zone_id` (String) — Zone the backup lives in.
- `project_id` (String) — Project the backup belongs to.
- `customer_username` (String) — Owner of the backup.
- `status` (String) — Lifecycle status (`active`, `creating`, `error`, `deleting`, `deleted`).
- `created_at` (String) — Creation timestamp (RFC3339).

## Timeouts

- `create` — default 15 minutes.
- `delete` — default 15 minutes.
## Import

```shell
terraform import vnpaycloud_backup_kubernetes.prod <cluster-backup-id>
```
