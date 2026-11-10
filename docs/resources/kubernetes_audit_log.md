---
page_title: "vnpaycloud_kubernetes_audit_log Resource - VNPayCloud"
subcategory: "Kubernetes"
description: |-
  Enables kube-apiserver audit logging on a Kubernetes cluster.
---

# vnpaycloud_kubernetes_audit_log (Resource)

Enables audit logging of the kube-apiserver on a `vnpaycloud_kubernetes_cluster`, using the audit policy managed by the platform. Enabling patches the control plane and replaces its nodes one at a time; the apply waits until the cluster is `active` again with audit logging on.

The control plane must meet the platform baseline first (`audit_log_ready = true` on the `vnpaycloud_kubernetes_control_plane_baseline` data source). For clusters created before the baseline existed, apply a `vnpaycloud_kubernetes_control_plane_rolling_update` and make this resource depend on it.

~> **Note:** The platform has no "disable" action. Destroying this resource only removes it from state; audit logging stays on. If the platform reports audit logging off at refresh time, the resource is dropped from state and the next apply enables it again.

## Example Usage

```hcl
resource "vnpaycloud_kubernetes_control_plane_rolling_update" "main" {
  cluster_id = vnpaycloud_kubernetes_cluster.main.id
}

resource "vnpaycloud_kubernetes_audit_log" "main" {
  cluster_id = vnpaycloud_kubernetes_cluster.main.id

  depends_on = [vnpaycloud_kubernetes_control_plane_rolling_update.main]
}
```

## Schema

### Required

- `cluster_id` (String, ForceNew) The ID of the cluster. The cluster must be `active`, not currently rolling its control plane, and meet the baseline.

### Read-Only

- `id` (String) Same as `cluster_id`.
- `enabled` (Boolean) Always `true` while the resource exists.

## Timeouts

- `create` - (Default `90 minutes`) Used while the control-plane nodes are replaced with audit logging enabled.

## Import

Import using the cluster ID (the cluster must already have audit logging enabled):

```shell
terraform import vnpaycloud_kubernetes_audit_log.main <cluster-id>
```
