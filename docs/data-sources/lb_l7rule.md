---
page_title: "vnpaycloud_lb_l7rule Data Source - VNPayCloud"
subcategory: "Load Balancer"
description: |-
  Look up an L7 rule by ID under its parent L7 policy.
---

# vnpaycloud_lb_l7rule (Data Source)

Read attributes of an existing L7 rule. Rules are nested under a parent L7 policy, so both `l7policy_id` and `id` are required to identify a rule uniquely.

## Example Usage

```hcl
data "vnpaycloud_lb_l7rule" "by_id" {
  l7policy_id = "sb_iaas_portal_l7_policy_lb_resourceAAAAAA"
  id          = "sb_iaas_portal_l7_rule_lb_resourceBBBBBB"
}

output "rule_value" {
  value = data.vnpaycloud_lb_l7rule.by_id.value
}
```

## Schema

### Required

- `id` (String) The L7 rule ID.
- `l7policy_id` (String) The parent L7 policy ID. Required because rule reads are scoped to the policy.

### Read-Only

- `rule_type` (String) Attribute inspected: `HOST_NAME`, `PATH`, `FILE_TYPE`, `HEADER`, `COOKIE`, `SSL_CONN_HAS_CERT`, `SSL_VERIFY_RESULT`, `SSL_DN_FIELD`.
- `compare_type` (String) Comparison: `REGEX`, `STARTS_WITH`, `ENDS_WITH`, `CONTAINS`, `EQUAL_TO`.
- `value` (String) The string compared against.
- `key` (String) Header / cookie / DN field name (set when `rule_type` is `HEADER`, `COOKIE`, or `SSL_DN_FIELD`).
- `invert` (Boolean) `true` = NOT semantic; rule matches when value does not satisfy `compare_type`.
- `status` (String) Lifecycle status: `active`, `creating`, `pending_create`, `pending_update`, `pending_delete`, `deleting`, `disabled`, `error`, `unknown`.
