---
page_title: "vnpaycloud_kubernetes_api_whitelist Resource - VNPayCloud"
subcategory: "Kubernetes"
description: |-
  Restricts which source IPs may reach the public Kubernetes API of a cluster.
---

# vnpaycloud_kubernetes_api_whitelist (Resource)

Manages the API whitelist of a **public** `vnpaycloud_kubernetes_cluster`. By default the public kube-apiserver endpoint of a cluster accepts connections from any IP; this resource limits it to the listed IPv4 CIDRs.

The platform always adds the addresses it needs itself (the VPC SNAT IP of the cluster and the management IPs of VNPayCloud) so nodes and the console keep working. Those system addresses are not shown in `entry` and cannot be removed.

~> **Note:** Only clusters with `access_mode = "public"` support a whitelist. Private clusters, clusters whose VPC has no SNAT, or zones whose load balancer provider cannot filter by source IP report `supported = false`; applying a whitelist to such a cluster fails. At least one `entry` is required, so a configuration cannot accidentally open the API to every IP; destroying the resource removes the restriction and allows every IP again.

~> **Note:** Applying a whitelist moves the cluster to `updating` for a short while and the apply waits until it is `active` again. Do not run other day-2 actions on the same cluster in parallel.

## Example Usage

```hcl
resource "vnpaycloud_kubernetes_api_whitelist" "office" {
  cluster_id = vnpaycloud_kubernetes_cluster.main.id

  entry {
    cidr = "203.0.113.0/24"
    note = "HQ office"
  }

  entry {
    cidr = "198.51.100.7/32"
    note = "CI runner"
  }
}
```

## Schema

### Required

- `cluster_id` (String, ForceNew) The ID of the public Kubernetes cluster. Changing this creates a new resource.

### Optional

- `entry` (Block Set, Required, 1 to 50) Allowed sources. Each block supports:
  - `cidr` (String, Required) An IPv4 network in CIDR notation, e.g. `203.0.113.0/24` or `198.51.100.7/32`. Must be a network address (not `203.0.113.5/24`) and must not be `0.0.0.0/0`. Validated at plan time. Write single hosts as `/32`; the API stores them that way.
  - `note` (String) Free text up to 64 characters, stored alongside the CIDR.

### Read-Only

- `id` (String) Same as `cluster_id`.
- `endpoint` (String) The public IP of the Kubernetes API that the whitelist protects.
- `private_access` (Boolean) Whether the cluster is only reachable privately.
- `max_entries` (Number) Maximum number of entries the platform accepts (currently 50).
- `supported` (Boolean) Whether this cluster can enforce a whitelist.
- `unsupported_reason` (String) Why `supported` is false, empty otherwise.

## Timeouts

- `create` - (Default `15 minutes`) Used while the platform applies the list and the cluster returns to `active`.
- `update` - (Default `15 minutes`) Same as create.
- `delete` - (Default `15 minutes`) Used while the list is cleared.

## Import

Import using the cluster ID:

```shell
terraform import vnpaycloud_kubernetes_api_whitelist.office <cluster-id>
```
