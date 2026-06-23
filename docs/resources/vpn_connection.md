---
page_title: "vnpaycloud_vpn_connection Resource - VNPayCloud"
subcategory: "Networking"
description: |-
  Manages a VPN connection within VNPayCloud.
---

# vnpaycloud_vpn_connection (Resource)

Manages a VPN connection within VNPayCloud. A VPN connection links a VPN gateway, a customer gateway, and a VPN public IP.

~> **Note:** VPN connection topology and tunnel configuration fields are immutable after creation. Changing those fields forces Terraform to replace the VPN connection. Updating `name` or `description` is not currently supported in Terraform and is rejected during planning to avoid recreating the VPN tunnel.

## Mode-Specific Requirements

The valid configuration depends on the VPN type and, for route-based VPN, on the referenced customer gateway's `routing_mode`.

| VPN connection mode | Required fields/blocks | Must not be set |
|---|---|---|
| Policy-based | `vpn_type = "POLICY_BASED"`, `ipsec_auth_config` | `route_base_config`, `connection_bgp_config` |
| Route-based static | `vpn_type = "ROUTE_BASED"`, customer gateway `routing_mode = "STATIC"`, `ipsec_auth_config`, `route_base_config` | `connection_bgp_config` |
| Route-based BGP | `vpn_type = "ROUTE_BASED"`, customer gateway `routing_mode = "DYNAMIC"`, `ipsec_auth_config`, `route_base_config`, `connection_bgp_config` | none |

Additional backend validation:

- The VPN gateway, customer gateway, and VPN public IP must be in the same zone and must be `active`.
- `vpn_type` must match both the VPN gateway and the customer gateway.
- The VPN gateway must have at least one VPC attached before creating a VPN connection.
- A VPN public IP can be used by only one blocking VPN connection.
- A customer gateway can be used by only one blocking VPN connection.
- A policy-based VPN gateway can have only one blocking VPN connection.
- Route-based static and route-based BGP connections cannot be mixed on the same VPN gateway while blocking connections exist.
- Multiple route-based connections on the same VPN gateway cannot use customer gateways with the same `public_ip`.
- `ike_profile_config` and `ipsec_profile_config` are optional. If omitted, backend defaults are used.

## Example Usage

### Route-Based Static VPN Connection

```hcl
resource "vnpaycloud_vpn_connection" "static" {
  name        = "tf-vpn-conn-static"
  description = "Route-based static VPN connection"

  vpn_gateway_id      = vnpaycloud_vpn_gateway.route_based.id
  customer_gateway_id = vnpaycloud_customer_gateway.static.id
  vpn_public_ip_id    = vnpaycloud_vpn_public_ip.main.id
  vpn_type            = "ROUTE_BASED"

  ipsec_auth_config {
    pre_shared_key = var.vpn_pre_shared_key
  }

  ike_profile_config {
    ike_version      = "IKE_V2"
    ike_lifetime     = 28800
    ike_close_action = "START"
    ike_dh           = "GROUP_14"
    ike_encryption   = "AES128_GCM96"
    ike_hash         = "SHA256"
    ike_prf          = "SHA1"
    ike_dpd_action   = "RESTART"
    ike_dpd_interval = 30
    ike_dpd_timeout  = 120
    ikev2_reauth     = false
  }

  ipsec_profile_config {
    ipsec_lifetime         = 3600
    ipsec_pfs              = "GROUP_14"
    ipsec_encryption       = "AES256"
    ipsec_hash             = "SHA256"
    ipsec_disable_rekey    = false
    ipsec_lifetime_bytes   = 0
    ipsec_lifetime_packets = 0
  }

  route_base_config {
    vti_mss = 1360
  }
}
```

### Route-Based BGP VPN Connection

```hcl
resource "vnpaycloud_vpn_connection" "bgp" {
  name        = "tf-vpn-conn-bgp"
  description = "Route-based BGP VPN connection"

  vpn_gateway_id      = vnpaycloud_vpn_gateway.route_based.id
  customer_gateway_id = vnpaycloud_customer_gateway.bgp.id
  vpn_public_ip_id    = vnpaycloud_vpn_public_ip.main.id
  vpn_type            = "ROUTE_BASED"

  ipsec_auth_config {
    pre_shared_key = var.vpn_pre_shared_key
  }

  ike_profile_config {
    ike_version      = "IKE_V2"
    ike_lifetime     = 28800
    ike_close_action = "START"
    ike_dh           = "GROUP_14"
    ike_encryption   = "AES128_GCM96"
    ike_hash         = "SHA256"
    ike_prf          = "SHA1"
    ike_dpd_action   = "RESTART"
    ike_dpd_interval = 30
    ike_dpd_timeout  = 120
    ikev2_reauth     = false
  }

  ipsec_profile_config {
    ipsec_lifetime         = 3600
    ipsec_pfs              = "GROUP_14"
    ipsec_encryption       = "AES256"
    ipsec_hash             = "SHA256"
    ipsec_disable_rekey    = false
    ipsec_lifetime_bytes   = 0
    ipsec_lifetime_packets = 0
  }

  route_base_config {
    vti_mss = 1360
  }

  connection_bgp_config {
    bgp_keepalive = 60
    bgp_holdtime  = 180
  }
}
```

### Policy-Based VPN Connection

```hcl
resource "vnpaycloud_vpn_connection" "policy" {
  name        = "tf-vpn-conn-policy"
  description = "Policy-based VPN connection"

  vpn_gateway_id      = vnpaycloud_vpn_gateway.policy_based.id
  customer_gateway_id = vnpaycloud_customer_gateway.policy.id
  vpn_public_ip_id    = vnpaycloud_vpn_public_ip.main.id
  vpn_type            = "POLICY_BASED"

  ipsec_auth_config {
    pre_shared_key = var.vpn_pre_shared_key
  }

  ike_profile_config {
    ike_version      = "IKE_V2"
    ike_lifetime     = 28800
    ike_close_action = "START"
    ike_dh           = "GROUP_14"
    ike_encryption   = "AES128_GCM96"
    ike_hash         = "SHA256"
    ike_prf          = "SHA1"
    ike_dpd_action   = "RESTART"
    ike_dpd_interval = 30
    ike_dpd_timeout  = 120
    ikev2_reauth     = false
  }

  ipsec_profile_config {
    ipsec_lifetime         = 3600
    ipsec_pfs              = "GROUP_14"
    ipsec_encryption       = "AES256"
    ipsec_hash             = "SHA256"
    ipsec_disable_rekey    = false
    ipsec_lifetime_bytes   = 0
    ipsec_lifetime_packets = 0
  }
}
```

## Schema

### Required

- `name` (String) The name of the VPN connection. Length must be between `3` and `255`, and it may contain only letters, digits, `.`, `_`, `-` and spaces. Updating this field is not currently supported in Terraform and is rejected during planning.
- `vpn_gateway_id` (String, ForceNew) The ID of the VPN gateway used by this VPN connection.
- `customer_gateway_id` (String, ForceNew) The ID of the customer gateway used by this VPN connection.
- `vpn_type` (String, ForceNew) The VPN type. Valid values are `POLICY_BASED` and `ROUTE_BASED`.
- `vpn_public_ip_id` (String, ForceNew) The ID of the VPN public IP associated with this VPN connection.
- `ipsec_auth_config` (Block List, Sensitive, ForceNew, Max: 1) The IPSec authentication configuration for this VPN connection. Required for all VPN connection modes.

### Optional

- `description` (String) A human-readable description of the VPN connection. Updating this field is not currently supported in Terraform and is rejected during planning.
- `ike_profile_config` (Block List, ForceNew, Max: 1) The IKE profile configuration.
- `ipsec_profile_config` (Block List, ForceNew, Max: 1) The IPSec profile configuration.
- `route_base_config` (Block List, ForceNew, Max: 1) The route-based VPN configuration. Required for route-based static and route-based BGP connections. Not allowed for policy-based connections.
- `connection_bgp_config` (Block List, ForceNew, Max: 1) The BGP timer configuration. Required only for route-based BGP connections. Not allowed for policy-based or route-based static connections.

### Nested Schema for `ipsec_auth_config`

#### Required

- `pre_shared_key` (String, Sensitive) The pre-shared key used for IPSec authentication. Length must be between `8` and `255`, and it may contain only letters, digits, `-`, `_` and `.` (spaces and other punctuation are rejected by the backend).

### Nested Schema for `ike_profile_config`

#### Optional

- `ike_version` (String) The IKE protocol version. Valid values are `IKE_V1` and `IKE_V2`. Defaults to `IKE_V2`.
- `ike_lifetime` (Number) The IKE SA lifetime in seconds. Range is `0` to `86400`. Defaults to `28800`.
- `ike_close_action` (String) The action to take when the IKE SA is closed. Valid values are `NONE`, `TRAP`, and `START`. Defaults to `START`.
- `ike_dh` (String) The IKE Diffie-Hellman group. Valid values are `GROUP_1`, `GROUP_2`, `GROUP_5`, and `GROUP_14` through `GROUP_32`. Defaults to `GROUP_14`.
- `ike_encryption` (String) The IKE encryption algorithm. Valid values are `AES128`, `AES192`, `AES256`, `AES128_GCM96`, `AES128_GCM128`, `AES256_GCM96`, and `AES256_GCM128`. Defaults to `AES128_GCM96`.
- `ike_hash` (String) The IKE integrity hash algorithm. Defaults to `SHA256`.
- `ike_prf` (String) The IKE pseudo-random function algorithm. Defaults to `SHA1`.
- `ike_dpd_action` (String) The IKE dead peer detection action. Valid values are `TRAP`, `CLEAR`, and `RESTART`. Defaults to `CLEAR`.
- `ike_dpd_interval` (Number) The IKE dead peer detection interval in seconds. Range is `2` to `86400`. Defaults to `30`.
- `ike_dpd_timeout` (Number) The IKE dead peer detection timeout in seconds. Range is `2` to `86400`. Defaults to `120`.
- `ikev2_reauth` (Boolean) Whether IKEv2 reauthentication is enabled. Defaults to `true`.

### Nested Schema for `ipsec_profile_config`

#### Optional

- `ipsec_lifetime` (Number) The IPSec SA lifetime in seconds. Range is `30` to `86400`. Defaults to `3600`.
- `ipsec_pfs` (String) The IPSec perfect forward secrecy group. Valid values are `GROUP_1`, `GROUP_2`, `GROUP_5`, and `GROUP_14` through `GROUP_32`. Defaults to `GROUP_14`.
- `ipsec_encryption` (String) The IPSec encryption algorithm. Valid values are `AES128`, `AES192`, `AES256`, `AES128_GCM96`, `AES128_GCM128`, `AES256_GCM96`, and `AES256_GCM128`. Defaults to `AES256`.
- `ipsec_hash` (String) The IPSec integrity hash algorithm. Defaults to `SHA256`.
- `ipsec_disable_rekey` (Boolean) Whether IPSec rekey is disabled. Defaults to `false`.
- `ipsec_lifetime_bytes` (Number) The IPSec SA lifetime in bytes. Must be `0` (disabled) or at least `1024`. Defaults to `0`.
- `ipsec_lifetime_packets` (Number) The IPSec SA lifetime in packets. Must be `0` (disabled) or at least `1024`. Defaults to `0`.

### Nested Schema for `route_base_config`

#### Optional

- `vti_mss` (Number) The TCP MSS value configured on the VTI interface. Defaults to `1350`.

### Nested Schema for `connection_bgp_config`

#### Optional

- `bgp_keepalive` (Number) The BGP keepalive interval in seconds. Range is `4` to `65535`. Defaults to `60`.
- `bgp_holdtime` (Number) The BGP hold time in seconds. Range is `4` to `65535`. Defaults to `180`.

### Read-Only

- `id` (String) The ID of the VPN connection.
- `status` (String) The current status of the VPN connection.
- `created_at` (String) The creation timestamp of the VPN connection.

## Timeouts

- `create` - (Default `10 minutes`) Used for creating the VPN connection.
- `delete` - (Default `10 minutes`) Used for deleting the VPN connection.

## Import

VPN connections can be imported using the `id`:

```shell
terraform import vnpaycloud_vpn_connection.example <vpn-connection-id>
```

~> **Note:** The `ipsec_auth_config.pre_shared_key` is write-only and is never returned by the API, so it cannot be recovered on import. Because the pre-shared key is required and immutable, the first plan after an import shows the resource being replaced (the in-config key is treated as a new value). To adopt an existing VPN connection without replacing it, set the same pre-shared key in configuration and add `lifecycle { ignore_changes = [ipsec_auth_config] }`.
