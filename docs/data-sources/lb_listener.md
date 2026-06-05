---
page_title: "vnpaycloud_lb_listener Data Source - VNPayCloud"
subcategory: "Load Balancer"
description: |-
  Get information about a load balancer listener in VNPayCloud.
---

# vnpaycloud_lb_listener (Data Source)

Use this data source to get information about an existing load balancer listener. A listener defines the protocol and port on which the load balancer accepts incoming connections.

## Example Usage

```hcl
data "vnpaycloud_lb_listener" "example" {
  id = "lst-bcd66778"
}

output "listener_protocol" {
  value = data.vnpaycloud_lb_listener.example.protocol
}

output "listener_default_pool" {
  value = data.vnpaycloud_lb_listener.example.default_pool_id
}
```

## Schema

### Required (filter)

- `id` (String) The ID of the load balancer listener.

### Read-Only

- `name` (String) The name of the listener.
- `description` (String) A human-readable description for the listener.
- `load_balancer_id` (String) The ID of the load balancer this listener belongs to.
- `protocol` (String) The protocol the listener accepts (e.g., `HTTP`, `HTTPS`, `TCP`, `UDP`).
- `protocol_port` (Number) The port number on which the listener accepts connections (e.g., `80`, `443`).
- `default_pool_id` (String) The ID of the default backend pool to which traffic is forwarded.
- `insert_headers` (List of String) The list of header names inserted into the request before forwarding to the backend.
- `allowed_cidrs` (List of String) The list of CIDR blocks permitted to connect to this listener.
- `connection_limit` (Number) The maximum number of connections permitted for this listener.
- `timeout_client_data` (Number) Frontend client inactivity timeout in milliseconds.
- `timeout_member_connect` (Number) Backend member connection timeout in milliseconds.
- `timeout_member_data` (Number) Backend member inactivity timeout in milliseconds.
- `certificate_id` (String) Server certificate ID. Set for `HTTPS` listeners.
- `certificate_authority_id` (String) Client CA certificate ID for mutual TLS.
- `sni_certificate_ids` (List of String) SNI certificate IDs.
- `status` (String) Lifecycle status: `active`, `creating`, `pending_create`, `pending_update`, `pending_delete`, `deleting`, `disabled`, `error`, `unknown`.
- `created_at` (String) The timestamp when the listener was created, in ISO 8601 format.
