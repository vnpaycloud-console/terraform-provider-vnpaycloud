---
page_title: "vnpaycloud_lb_l7rule Resource - VNPayCloud"
subcategory: "Load Balancer"
description: |-
  Manages an L7 rule attached to an L7 policy within VNPayCloud.
---

# vnpaycloud_lb_l7rule (Resource)

Manages an L7 rule attached to a [`vnpaycloud_lb_l7policy`](lb_l7policy.md). An L7 rule specifies a single match condition (path prefix, hostname, header value, etc.). A policy fires when **all** of its rules match (AND semantics). To express OR, use multiple policies.

## Example Usage

### Path prefix match

```hcl
resource "vnpaycloud_lb_l7rule" "api_path" {
  l7policy_id  = vnpaycloud_lb_l7policy.route_api.id
  rule_type    = "PATH"
  compare_type = "STARTS_WITH"
  value        = "/api/"
}
```

### Host header match

```hcl
resource "vnpaycloud_lb_l7rule" "api_subdomain" {
  l7policy_id  = vnpaycloud_lb_l7policy.route_api.id
  rule_type    = "HOST_NAME"
  compare_type = "EQUAL_TO"
  value        = "api.example.com"
}
```

### Cookie match (rule_type=COOKIE requires `key`)

```hcl
resource "vnpaycloud_lb_l7rule" "internal_cookie" {
  l7policy_id  = vnpaycloud_lb_l7policy.block_internal.id
  rule_type    = "COOKIE"
  compare_type = "EQUAL_TO"
  key          = "session_tier"
  value        = "internal"
}
```

### Inverted match (NOT condition)

```hcl
resource "vnpaycloud_lb_l7rule" "not_static" {
  l7policy_id  = vnpaycloud_lb_l7policy.route_api.id
  rule_type    = "PATH"
  compare_type = "STARTS_WITH"
  value        = "/static/"
  invert       = true
}
```

## Schema

### Required

- `l7policy_id` (String, ForceNew) The ID of the parent L7 policy.
- `rule_type` (String) The attribute of the HTTP request to inspect. One of:
  - `HOST_NAME` — the `Host` header.
  - `PATH` — the URL path.
  - `COOKIE` — value of a cookie. Requires `key`.
- `compare_type` (String) How to compare. One of `REGEX`, `STARTS_WITH`, `ENDS_WITH`, `CONTAINS`, `EQUAL_TO`.
- `value` (String) The string to match against. Length `1`–`255`.

### Optional

- `key` (String) The name of the cookie. **Required** when `rule_type` is `COOKIE`; must be **empty** for other types.
- `invert` (Boolean, Default `false`) Invert the match (NOT). When `true`, the rule matches when the value does **not** satisfy the comparison.

### Read-Only

- `id` (String) The L7 rule ID.
- `status` (String) Lifecycle status: `active`, `creating`, `pending_create`, `pending_update`, `pending_delete`, `deleting`, `disabled`, `error`, `unknown`.

## In-place updates

`rule_type`, `compare_type`, `value`, `key`, `invert` are updatable.

`l7policy_id` is `ForceNew`.

## Timeouts

- `create` - (Default `10 minutes`)
- `update` - (Default `10 minutes`)
- `delete` - (Default `10 minutes`)

~> **Rate limit:** see [Rate limits](../index.md#rate-limits) — applies to all create/update/delete on this resource type.

## Import

**L7 rules use a composite import ID** — unlike other LB resources which import by a single ID. The format is `<l7policy_id>/<rule_id>` (slash-separated, parent policy ID first).

```shell
terraform import vnpaycloud_lb_l7rule.example <l7policy-id>/<rule-id>
```
