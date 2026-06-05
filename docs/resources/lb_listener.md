---
page_title: "vnpaycloud_lb_listener Resource - VNPayCloud"
subcategory: "Load Balancer"
description: |-
  Manages a load balancer listener within VNPayCloud.
---

# vnpaycloud_lb_listener (Resource)

Manages a listener for a load balancer within VNPayCloud. A listener defines the protocol and port on which the load balancer accepts incoming connections and forwards them to a backend pool.

## Example Usage

### HTTP listener

```hcl
resource "vnpaycloud_lb_loadbalancer" "app" {
  name      = "app-loadbalancer"
  subnet_id = "subnet-abc12345"
  flavor    = "t1-small"
}

resource "vnpaycloud_lb_listener" "http" {
  name             = "http-listener"
  load_balancer_id = vnpaycloud_lb_loadbalancer.app.id
  protocol         = "HTTP"
  protocol_port    = 80
}
```

### HTTP listener with headers, ACL, and timeouts

```hcl
resource "vnpaycloud_lb_listener" "http_advanced" {
  name             = "http-advanced-listener"
  description      = "HTTP listener with advanced settings"
  load_balancer_id = vnpaycloud_lb_loadbalancer.app.id
  protocol         = "HTTP"
  protocol_port    = 8080

  insert_headers         = ["X-Forwarded-For", "X-Forwarded-Port", "X-Forwarded-Proto"]
  allowed_cidrs          = ["10.0.0.0/8", "192.168.1.0/24"]
  connection_limit       = 10000
  timeout_client_data    = 50000
  timeout_member_connect = 5000
  timeout_member_data    = 50000
}
```

### HTTPS listener with certificate and SNI

Look up certificate IDs by name with the [`vnpaycloud_certificates`](../data-sources/certificates.md) data source instead of hardcoding opaque IDs:

```hcl
data "vnpaycloud_certificates" "all" {}

locals {
  server_cert_id = one([for c in data.vnpaycloud_certificates.all.certificates : c.id if c.name == "my-server-cert"])
  client_ca_id   = one([for c in data.vnpaycloud_certificates.all.certificates : c.id if c.name == "my-client-ca"])
}

resource "vnpaycloud_lb_listener" "tls" {
  name                     = "tls-listener"
  load_balancer_id         = vnpaycloud_lb_loadbalancer.app.id
  protocol                 = "HTTPS"
  protocol_port            = 443
  certificate_id           = local.server_cert_id
  certificate_authority_id = local.client_ca_id # optional, for mutual TLS
  sni_certificate_ids      = [local.server_cert_id]

  # X-SSL-* headers only allowed when SSL is terminated at the listener
  insert_headers = ["X-Forwarded-For", "X-SSL-Client-Verify", "X-SSL-Client-CN"]
}
```

### TCP listener

```hcl
resource "vnpaycloud_lb_listener" "tcp" {
  name             = "tcp-listener"
  load_balancer_id = vnpaycloud_lb_loadbalancer.app.id
  protocol         = "TCP"
  protocol_port    = 6443
}
```

### `insert_headers` — protocol contrast (valid vs invalid)

```hcl
# ✅ HTTP listener — only X-Forwarded-* allowed
resource "vnpaycloud_lb_listener" "ok_http" {
  name             = "ok-http"
  protocol         = "HTTP"
  protocol_port    = 80
  load_balancer_id = vnpaycloud_lb_loadbalancer.app.id
  insert_headers   = ["X-Forwarded-For", "X-Forwarded-Port", "X-Forwarded-Proto"]
}

# ✅ HTTPS listener — X-Forwarded-* AND X-SSL-* allowed (SSL terminated at the listener)
resource "vnpaycloud_lb_listener" "ok_tls" {
  name             = "ok-tls"
  protocol         = "HTTPS"
  protocol_port    = 443
  load_balancer_id = vnpaycloud_lb_loadbalancer.app.id
  certificate_id   = local.server_cert_id
  insert_headers   = ["X-Forwarded-For", "X-SSL-Client-Verify", "X-SSL-Client-CN"]
}

# ❌ INVALID — X-SSL-* on plain HTTP listener (no TLS termination, header has no value to inject)
# Server rejects at create with a clear error; nothing prevents this at plan time.
# resource "vnpaycloud_lb_listener" "bad" {
#   protocol         = "HTTP"
#   protocol_port    = 80
#   load_balancer_id = vnpaycloud_lb_loadbalancer.app.id
#   insert_headers   = ["X-SSL-Client-CN"]  # rejected server-side
# }

# ❌ INVALID — insert_headers on TCP/UDP/HTTPS-passthrough listener (no HTTP layer to insert into)
# resource "vnpaycloud_lb_listener" "bad_tcp" {
#   protocol         = "TCP"
#   protocol_port    = 6443
#   load_balancer_id = vnpaycloud_lb_loadbalancer.app.id
#   insert_headers   = ["X-Forwarded-For"]  # rejected server-side
# }
```

## Schema

### Required

- `name` (String) The name of the listener. Length `3`–`250`, no leading/trailing whitespace.
- `load_balancer_id` (String, ForceNew) The ID of the load balancer to attach to.
- `protocol` (String, ForceNew) The protocol the listener accepts. One of:
  - `HTTP` — plaintext HTTP.
  - `HTTPS` — TLS terminated at the listener; requires `certificate_id` (server-side enforced).
  - `TCP` — raw TCP passthrough.
  - `UDP` — raw UDP.

  **Listener ↔ Pool protocol compatibility** — a pool attached to this listener (via `default_pool_id` or the pool's `listener_id`) must use a compatible `protocol`. Incompatible combinations are rejected by the backend:

  | Listener `protocol` \ Pool `protocol` | `HTTP` | `HTTPS` | `TCP` | `UDP` | `PROXY` |
  |---|:---:|:---:|:---:|:---:|:---:|
  | `HTTP` | ✓ |  |  |  | ✓ |
  | `HTTPS` | ✓ |  |  |  | ✓ |
  | `TCP` | ✓ | ✓ | ✓ |  | ✓ |
  | `UDP` |  |  |  | ✓ |  |
- `protocol_port` (Number, ForceNew) Port on which the listener accepts traffic. Range `1`–`65535`. Must be unique per `(load_balancer_id, protocol)` — the provider rejects duplicates server-side.

### Optional

- `description` (String) A human-readable description. Length `0`–`255`.
- `default_pool_id` (String, Optional, Computed) The ID of the default pool to route traffic to.

  **TL;DR:** at most one default per listener; once set, removing the value will not detach it — destroy and recreate the listener instead.

  Attach via `listener_id` on the pool or `default_pool_id` here — both converge and neither recreates. Via the pool's `listener_id` the first `plan` after create is clean (no drift); setting `default_pool_id` here instead shows one benign in-place sync that settles after a second `apply` (no recreate).

  A listener accepts **at most one** default pool. Creating a second pool with `listener_id` pointing at a listener that already has a default is rejected server-side with a clear error; swap by updating this field on the listener instead. The platform does not support detaching a default once attached — clearing this field from config will not clear the attachment server-side (drift is suppressed). To remove entirely, destroy and recreate the listener.
- `insert_headers` (List of String, Computed) HTTP headers to insert into the request before forwarding. Server-default applies when omitted (no drift on import). Protocol rules enforced server-side:
  - `HTTP` listener: only `X-Forwarded-*` headers (`X-Forwarded-For`, `X-Forwarded-Port`, `X-Forwarded-Proto`).
  - `HTTPS` listener: `X-Forwarded-*` plus `X-SSL-Client-Verify`, `X-SSL-Client-Has-Cert`, `X-SSL-Client-DN`, `X-SSL-Client-CN`, `X-SSL-Issuer`, `X-SSL-Client-SHA1`, `X-SSL-Client-Not-Before`, `X-SSL-Client-Not-After`.
  - Other protocols (`TCP`, `UDP`): no headers allowed.
- `allowed_cidrs` (List of String, Computed) CIDR blocks permitted to connect. Each element must be a valid CIDR (validated at plan time). Server-default applies when omitted (no drift on import).
- `connection_limit` (Number, Computed) Maximum concurrent connections. Stored as `0` if omitted at create — set an explicit value (e.g. `10000`), or `-1` for unlimited.
- `timeout_client_data` (Number, Computed) Frontend client inactivity timeout (ms). Stored as `0` if omitted at create — set an explicit value.
- `timeout_member_connect` (Number, Computed) Backend member connection timeout (ms). Stored as `0` if omitted at create — set an explicit value.
- `timeout_member_data` (Number, Computed) Backend member inactivity timeout (ms). Stored as `0` if omitted at create — set an explicit value.
- `certificate_id` (String) Server certificate ID. **Required when `protocol = HTTPS`** (enforced server-side, not at plan time). Server validates the certificate exists and is type `CT_SIGNED` or `CT_SELF_SIGNED`. Forbidden for other protocols.
- `certificate_authority_id` (String) Client CA certificate ID for mutual TLS. Only valid for `HTTPS`. Server validates type is `CT_CA` or `CT_INTERMEDIATE_CA`.
- `sni_certificate_ids` (List of String) SNI certificate IDs. Only valid for `HTTPS`. Server validates each exists and is a valid server certificate type.

### Read-Only

- `id` (String) The listener ID.
- `status` (String) Lifecycle status: `active`, `creating`, `pending_create`, `pending_update`, `pending_delete`, `deleting`, `disabled`, `error`, `unknown`.
- `created_at` (String) Creation timestamp.

## In-place updates

The following attributes can be updated without recreation:
`name`, `description`, `default_pool_id`, `insert_headers`, `allowed_cidrs`, `connection_limit`, `timeout_*`, `certificate_id`, `certificate_authority_id`, `sni_certificate_ids`.

`load_balancer_id`, `protocol`, `protocol_port` are `ForceNew`.

## Timeouts

- `create` - (Default `10 minutes`)
- `update` - (Default `10 minutes`)
- `delete` - (Default `10 minutes`)

~> **Rate limit:** see [Rate limits](../index.md#rate-limits) — applies to all create/update/delete on this resource type.

## Import

```shell
terraform import vnpaycloud_lb_listener.example <listener-id>
```

After import you only need to declare the **required** fields (`name`, `load_balancer_id`, `protocol`, `protocol_port`) — `insert_headers`, `allowed_cidrs`, and other server-default fields are preserved from the live resource and will not show drift in `terraform plan`. Declare them explicitly only when you intend to manage them.
