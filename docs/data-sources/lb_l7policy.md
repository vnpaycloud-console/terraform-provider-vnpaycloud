---
page_title: "vnpaycloud_lb_l7policy Data Source - VNPayCloud"
subcategory: "Load Balancer"
description: |-
  Look up an L7 policy by ID.
---

# vnpaycloud_lb_l7policy (Data Source)

Read attributes of an existing L7 policy attached to a load balancer listener. Use this to reference a policy created outside Terraform (e.g., via console) or to chain L7 rules without keeping the policy in your state.

## Example Usage

```hcl
data "vnpaycloud_lb_l7policy" "by_id" {
  id = "sb_iaas_portal_l7_policy_lb_resourceXXXXXXXXXXXXXXXX"
}

output "policy_action" {
  value = data.vnpaycloud_lb_l7policy.by_id.action
}
```

## Schema

### Required

- `id` (String) The L7 policy ID.

### Read-Only

- `name` (String) The policy name.
- `description` (String) Human-readable description.
- `listener_id` (String) ID of the parent listener.
- `action` (String) `REJECT`, `REDIRECT_TO_URL`, `REDIRECT_TO_POOL`, or `REDIRECT_PREFIX`.
- `position` (Number) Evaluation order — lower wins first.
- `redirect_pool_id` (String) Target pool ID when `action = REDIRECT_TO_POOL` (empty otherwise).
- `redirect_url` (String) Target URL when `action = REDIRECT_TO_URL` or `REDIRECT_PREFIX` (empty otherwise).
- `status` (String) Lifecycle status: `active`, `creating`, `pending_create`, `pending_update`, `pending_delete`, `deleting`, `disabled`, `error`, `unknown`.
