---
page_title: "vnpaycloud_lb_l7policies Data Source - VNPayCloud"
subcategory: "Load Balancer"
description: |-
  List all L7 policies in VNPayCloud.
---

# vnpaycloud_lb_l7policies (Data Source)

Use this data source to list all L7 policies in the current project.

## Example Usage

```hcl
data "vnpaycloud_lb_l7policies" "all" {}

output "all_l7policy_names" {
  value = data.vnpaycloud_lb_l7policies.all.l7policies[*].name
}

output "redirect_to_url_policy_ids" {
  value = [
    for p in data.vnpaycloud_lb_l7policies.all.l7policies :
    p.id if p.action == "REDIRECT_TO_URL"
  ]
}
```

## Schema

### Read-Only

- `l7policies` (List of Object) List of L7 policies. Each element contains:
  - `id` (String) The unique identifier of the L7 policy.
  - `name` (String) The policy name.
  - `description` (String) Human-readable description.
  - `listener_id` (String) ID of the parent listener.
  - `action` (String) The action taken when the policy matches: `REJECT`, `REDIRECT_TO_URL`, or `REDIRECT_TO_POOL`.
  - `position` (Number) Evaluation order — lower position is evaluated first.
  - `redirect_pool_id` (String) Target pool ID when `action = REDIRECT_TO_POOL` (empty otherwise).
  - `redirect_url` (String) Target URL when `action = REDIRECT_TO_URL` (empty otherwise).
  - `status` (String) Lifecycle status: `active`, `creating`, `pending_create`, `pending_update`, `pending_delete`, `deleting`, `disabled`, `error`, `unknown`.
