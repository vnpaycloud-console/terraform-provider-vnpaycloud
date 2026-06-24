---
page_title: "vnpaycloud_lb_l7rules Data Source - VNPayCloud"
subcategory: "Load Balancer"
description: |-
  List all L7 rules of a given L7 policy in VNPayCloud.
---

# vnpaycloud_lb_l7rules (Data Source)

Use this data source to list all L7 rules belonging to a specific L7 policy. Rule listing is scoped to a single policy, so `l7policy_id` is required.

## Example Usage

```hcl
data "vnpaycloud_lb_l7rules" "by_policy" {
  l7policy_id = "l7p-abc12345"
}

output "all_rule_values" {
  value = data.vnpaycloud_lb_l7rules.by_policy.l7rules[*].value
}

output "host_name_rule_ids" {
  value = [
    for r in data.vnpaycloud_lb_l7rules.by_policy.l7rules :
    r.id if r.rule_type == "HOST_NAME"
  ]
}
```

## Schema

### Required

- `l7policy_id` (String) The parent L7 policy ID. Required because rule listing is scoped to the policy.

### Read-Only

- `l7rules` (List of Object) List of L7 rules in the policy. Each element contains:
  - `id` (String) The unique identifier of the L7 rule.
  - `l7policy_id` (String) The parent L7 policy ID.
  - `rule_type` (String) Attribute inspected: `HOST_NAME`, `PATH`, or `COOKIE`.
  - `compare_type` (String) Comparison: `REGEX`, `STARTS_WITH`, `ENDS_WITH`, `CONTAINS`, `EQUAL_TO`.
  - `value` (String) The string compared against.
  - `key` (String) Cookie name (set only when `rule_type` is `COOKIE`).
  - `invert` (Boolean) `true` = NOT semantic; the rule matches when the value does not satisfy `compare_type`.
  - `status` (String) Lifecycle status: `active`, `creating`, `pending_create`, `pending_update`, `pending_delete`, `deleting`, `disabled`, `error`, `unknown`.
