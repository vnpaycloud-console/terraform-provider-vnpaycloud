---
page_title: "vnpaycloud_kubernetes_credentials_rotation Resource - VNPayCloud"
subcategory: "Kubernetes"
description: |-
  Rotates the OpenStack application credentials used by a Kubernetes cluster.
---

# vnpaycloud_kubernetes_credentials_rotation (Resource)

Rotates the two application credentials a `vnpaycloud_kubernetes_cluster` uses in your project: the data-plane credential (used by the platform to create and delete worker VMs) and the OCCM credential (mounted into the cluster for the cloud controller manager, the Cinder CSI driver and Velero). Both are replaced in one operation and the old credentials are revoked afterwards.

Rotate when a credential may have leaked, for example because someone with admin kubeconfig access left the team.

This resource models a one-shot action: applying it triggers a rotation and waits until the cluster is `active` again. To rotate again later, change `triggers` (or taint the resource). Destroying it only removes it from state.

~> **Note:** The platform reports the cluster as `active` even if the rotation had to be rolled back to the previous credentials. Check the cluster events in the console if a rotation was expected to change behaviour and did not.

## Example Usage

```hcl
resource "vnpaycloud_kubernetes_credentials_rotation" "main" {
  cluster_id = vnpaycloud_kubernetes_cluster.main.id

  triggers = {
    rotation = "2026-11-01"
  }
}
```

## Schema

### Required

- `cluster_id` (String, ForceNew) The ID of the cluster whose credentials are rotated. The cluster must be `active`.

### Optional

- `triggers` (Map of String, ForceNew) Arbitrary values; changing any of them re-runs the rotation.

### Read-Only

- `id` (String) Composite of `cluster_id` and `rotated_at`.
- `rotated_at` (String) UTC timestamp (RFC 3339) recorded when the rotation finished.

## Timeouts

- `create` - (Default `30 minutes`) Used while the platform rotates the credentials, restarts the cloud controller manager, CSI driver and Velero, and the cluster returns to `active`.
