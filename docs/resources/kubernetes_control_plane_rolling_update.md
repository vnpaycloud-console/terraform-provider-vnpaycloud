---
page_title: "vnpaycloud_kubernetes_control_plane_rolling_update Resource - VNPayCloud"
subcategory: "Kubernetes"
description: |-
  Brings the control plane of a Kubernetes cluster up to the platform baseline.
---

# vnpaycloud_kubernetes_control_plane_rolling_update (Resource)

Brings the control-plane nodes of a `vnpaycloud_kubernetes_cluster` up to the platform's current **control plane baseline** (for example a larger system disk and the managed audit policy). Features such as `vnpaycloud_kubernetes_audit_log` require the baseline.

Applying this resource compares the cluster with the baseline. If nothing is outdated it does nothing and records `performed = false`. Otherwise it triggers the rolling update, which replaces the control-plane nodes one at a time, and waits until the cluster is `active` and meets the baseline. Clusters created after the baseline was introduced already meet it.

This resource models a one-shot action. The baseline attributes are refreshed on every plan, so `rolling_update_needed` tells you when the platform published a newer baseline; change `triggers` (or taint the resource) to run the update again. Destroying it only removes it from state.

~> **Note:** A rolling update of the control plane takes a while (typically 20–40 minutes) and the cluster stays `updating` throughout. Do not run other day-2 actions on the same cluster in parallel.

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

- `cluster_id` (String, ForceNew) The ID of the cluster. The cluster must be `active` and the baseline must be available in its zone (`supported = true`).

### Optional

- `triggers` (Map of String, ForceNew) Arbitrary values; changing any of them re-runs the comparison and, if needed, the rolling update.

### Read-Only

- `id` (String) Same as `cluster_id`.
- `performed` (Boolean) Whether this apply actually replaced the control-plane nodes.
- `supported` (Boolean) Whether a baseline exists for this cluster.
- `unsupported_reason` (String) Why `supported` is false, empty otherwise.
- `rolling_update_needed` (Boolean) Whether the control plane currently lags behind the baseline.
- `outdated_items` (List of String) Which parts lag behind: `CONTROL_PLANE_DISK`, `AUDIT_POLICY`.
- `control_plane_rolling` (Boolean) Whether control-plane nodes are being replaced right now.
- `audit_log_enabled` (Boolean) Whether kube-apiserver audit logging is enabled.
- `audit_log_ready` (Boolean) Whether the control plane meets everything audit logging needs.

## Timeouts

- `create` - (Default `90 minutes`) Used while the control-plane nodes are replaced.
