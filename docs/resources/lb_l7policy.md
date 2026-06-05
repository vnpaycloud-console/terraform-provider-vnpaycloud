---
page_title: "vnpaycloud_lb_l7policy Resource - VNPayCloud"
subcategory: "Load Balancer"
description: |-
  Manages an L7 policy attached to a load balancer listener within VNPayCloud.
---

# vnpaycloud_lb_l7policy (Resource)

Manages an L7 policy attached to a load balancer listener. An L7 policy inspects incoming HTTP requests (matched by associated [`vnpaycloud_lb_l7rule`](lb_l7rule.md) resources) and routes them to a different backend pool, redirects to a URL/prefix, or rejects the request.

L7 policies are evaluated in `position` order; the first matching policy wins.

~> **Listener protocol requirement:** L7 policies can only attach to listeners with `protocol = HTTP` or `protocol = HTTPS` (enforced server-side). L7 inspection requires plaintext HTTP traffic, which is only available when SSL is terminated at the listener.

## Example Usage

### Redirect /api/* to a different pool

```hcl
resource "vnpaycloud_lb_pool" "api_pool" {
  name             = "api-backend"
  load_balancer_id = vnpaycloud_lb_loadbalancer.app.id
  listener_id      = vnpaycloud_lb_listener.http.id
  protocol         = "HTTP"
  lb_algorithm     = "ROUND_ROBIN"
}

resource "vnpaycloud_lb_l7policy" "route_api" {
  name             = "route-api-prefix"
  listener_id      = vnpaycloud_lb_listener.http.id
  action           = "REDIRECT_TO_POOL"
  position         = 1
  redirect_pool_id = vnpaycloud_lb_pool.api_pool.id
}

resource "vnpaycloud_lb_l7rule" "api_path" {
  l7policy_id  = vnpaycloud_lb_l7policy.route_api.id
  rule_type    = "PATH"
  compare_type = "STARTS_WITH"
  value        = "/api/"
}
```

### Redirect HTTP to HTTPS

```hcl
resource "vnpaycloud_lb_l7policy" "force_https" {
  name         = "force-https"
  listener_id  = vnpaycloud_lb_listener.http.id
  action       = "REDIRECT_TO_URL"
  position     = 1
  redirect_url = "https://example.com/"
}
```

### Reject requests to internal paths

```hcl
resource "vnpaycloud_lb_l7policy" "block_internal" {
  name        = "block-internal"
  listener_id = vnpaycloud_lb_listener.http.id
  action      = "REJECT"
  position    = 2
}

resource "vnpaycloud_lb_l7rule" "internal_path" {
  l7policy_id  = vnpaycloud_lb_l7policy.block_internal.id
  rule_type    = "PATH"
  compare_type = "STARTS_WITH"
  value        = "/internal/"
}
```

## Schema

### Required

- `listener_id` (String, ForceNew) The ID of the listener this policy attaches to. Listener protocol must be `HTTP` or `HTTPS`.
- `action` (String) The action to perform when the policy matches. One of:
  - `REJECT` — drop the request (returns 403). `redirect_url` and `redirect_pool_id` must be empty.
  - `REDIRECT_TO_URL` — return a 302 with `Location: <redirect_url>` **exactly as given**; the original request path and query are discarded. Use when every match should land on the same destination URL.
  - `REDIRECT_TO_POOL` — route the request to the specified pool. Requires `redirect_pool_id`. The pool's protocol must be compatible with the listener's protocol.

### Optional

- `name` (String) The policy name. The schema marks it optional, but the server **requires** a name of length `3`–`250` (no leading/trailing whitespace) and rejects an empty value — always set one. Unlike health monitors, an L7 policy name is not auto-generated.
- `description` (String) A human-readable description. Length `0`–`255`.
- `position` (Number, Optional, Computed) Evaluation order (lower = higher priority); must be `>= 1`. If omitted, the server assigns the position. To control ordering across multiple policies on a listener, set an explicit value `>= 1`.
- `redirect_pool_id` (String) Required when `action = REDIRECT_TO_POOL`. Forbidden otherwise.
- `redirect_url` (String) Required when `action = REDIRECT_TO_URL`. Must start with `http://`, `https://`, or `/`. Forbidden for other actions.

### Read-Only

- `id` (String) The L7 policy ID.
- `status` (String) Lifecycle status: `active`, `creating`, `pending_create`, `pending_update`, `pending_delete`, `deleting`, `disabled`, `error`, `unknown`.

## In-place updates

`name`, `description`, `action`, `position`, `redirect_pool_id`, `redirect_url` are updatable.

`listener_id` is `ForceNew`.

## Timeouts

- `create` - (Default `10 minutes`)
- `update` - (Default `10 minutes`)
- `delete` - (Default `10 minutes`)

~> **Rate limit:** see [Rate limits](../index.md#rate-limits) — applies to all create/update/delete on this resource type.

## Import

```shell
terraform import vnpaycloud_lb_l7policy.example <l7policy-id>
```
