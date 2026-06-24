---
page_title: "vnpaycloud_lb_listeners Data Source - VNPayCloud"
subcategory: "Load Balancer"
description: |-
  List all load balancer listeners in VNPayCloud.
---

# vnpaycloud_lb_listeners (Data Source)

Use this data source to list all load balancer listeners in the current project.

## Example Usage

```hcl
data "vnpaycloud_lb_listeners" "all" {}

output "all_listener_names" {
  value = data.vnpaycloud_lb_listeners.all.listeners[*].name
}

output "https_listener_ids" {
  value = [
    for l in data.vnpaycloud_lb_listeners.all.listeners :
    l.id if l.protocol == "HTTPS"
  ]
}
```

## Schema

### Read-Only

- `listeners` (List of Object) List of listeners. Each element contains:
  - `id` (String) The unique identifier of the listener.
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
