---
page_title: "vnpaycloud_kubernetes_control_plane_baseline Data Source - VNPayCloud"
subcategory: "Kubernetes"
description: |-
  Compare the control plane of a Kubernetes cluster with the platform baseline.
---

# vnpaycloud_kubernetes_control_plane_baseline (Data Source)

Reports how the control plane of a cluster compares with the platform's control plane baseline and whether kube-apiserver audit logging is enabled. Use it to decide whether a `vnpaycloud_kubernetes_control_plane_rolling_update` is needed before `vnpaycloud_kubernetes_audit_log`.

## Example Usage

```hcl
data "vnpaycloud_kubernetes_control_plane_baseline" "main" {
  cluster_id = vnpaycloud_kubernetes_cluster.main.id
}

output "needs_rolling_update" {
  value = data.vnpaycloud_kubernetes_control_plane_baseline.main.rolling_update_needed
}
```

## Schema

### Required

- `cluster_id` (String) The ID of the Kubernetes cluster.

### Read-Only

- `supported` (Boolean) Whether a baseline exists for this cluster and zone.
- `unsupported_reason` (String) Why `supported` is false, empty otherwise.
- `rolling_update_needed` (Boolean) Whether the control plane lags behind the baseline.
- `outdated_items` (List of String) Which parts lag behind: `CONTROL_PLANE_DISK`, `AUDIT_POLICY`.
- `control_plane_rolling` (Boolean) Whether control-plane nodes are being replaced right now.
- `audit_log_enabled` (Boolean) Whether kube-apiserver audit logging is enabled.
- `audit_log_ready` (Boolean) Whether the control plane meets everything audit logging needs.
