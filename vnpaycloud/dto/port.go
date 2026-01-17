package dto

import (
	"terraform-provider-vnpaycloud/vnpaycloud/types"
)

type FixedIP struct {
	SubnetID  string `json:"subnet_id"`
	IPAddress string `json:"ip_address"`
}

type AddressPair struct {
	IPAddress  string `json:"ip_address"`
	MACAddress string `json:"mac_address"`
}

type DHCPOpt struct {
	OptName  string `json:"opt_name"`
	OptValue string `json:"opt_value"`
}

type Port struct {
	ID                  string        `json:"id"`
	Name                string        `json:"name"`
	NetworkID           string        `json:"network_id"`
	AdminStateUp        bool          `json:"admin_state_up"`
	MACAddress          string        `json:"mac_address"`
	Status              string        `json:"status"`
	DeviceID            string        `json:"device_id"`
	DeviceOwner         string        `json:"device_owner"`
	FixedIPs            []FixedIP     `json:"fixed_ips"`
	TenantID            string        `json:"tenant_id"` // or ProjectID based openstack version
	ProjectID           string        `json:"project_id"`
	SecurityGroups      []string      `json:"security_groups"`
	CreatedAt           string        `json:"created_at"`
	UpdatedAt           string        `json:"updated_at"`
	Description         string        `json:"description"`
	AllowedAddressPairs []AddressPair `json:"allowed_address_pairs"`
	Tags                []string      `json:"tags"`
}

type PortExtended struct {
	Port
	ExtraDHCPOptsExt
	PortSecurityExt
	PortsBindingExt
	PortDNSExt
	QoSPolicyExt
}

type GetPortResponse struct {
	Port PortExtended `json:"port"`
}

type CreatePortRequest struct {
	Port CreatePortOpts `json:"port" required:"true"`
}
type CreatePortOpts struct {
	NetworkID             string               `json:"network_id" required:"true"`
	Name                  string               `json:"name,omitempty"`
	Description           string               `json:"description,omitempty"`
	AdminStateUp          *bool                `json:"admin_state_up,omitempty"`
	MACAddress            string               `json:"mac_address,omitempty"`
	FixedIPs              any                  `json:"fixed_ips,omitempty"`
	DeviceID              string               `json:"device_id,omitempty"`
	DeviceOwner           string               `json:"device_owner,omitempty"`
	TenantID              string               `json:"tenant_id,omitempty"`
	ProjectID             string               `json:"project_id,omitempty"`
	SecurityGroups        *[]string            `json:"security_groups,omitempty"`
	AllowedAddressPairs   []AddressPair        `json:"allowed_address_pairs,omitempty"`
	PropagateUplinkStatus *bool                `json:"propagate_uplink_status,omitempty"`
	ValueSpecs            map[string]string    `json:"value_specs,omitempty"`
	VirtualIp             *bool                `json:"virtual_ip,omitempty"`
	ExtraDHCPOpts         []CreateExtraDHCPOpt `json:"extra_dhcp_opts,omitempty"`
	PortSecurityEnabled   *bool                `json:"port_security_enabled,omitempty"`
	HostID                string               `json:"binding:host_id,omitempty"`
	VNICType              string               `json:"binding:vnic_type,omitempty"`
	Profile               map[string]any       `json:"binding:profile,omitempty"`
	DNSName               string               `json:"dns_name,omitempty"`
	QoSPolicyID           string               `json:"qos_policy_id,omitempty"`
}

type CreatePortResponse struct {
	Port Port `json:"port"`
}

type FixedIPOpts struct {
	IPAddress       string
	IPAddressSubstr string
	SubnetID        string
}

// ListPortOpts allows the filtering and sorting of paginated collections through
// the API. Filtering is achieved by passing in struct field values that map to
// the port attributes you want to see returned. SortKey allows you to sort
// by a particular port attribute. SortDir sets the direction, and is either
// `asc' or `desc'. Marker and Limit are used for pagination.
type ListPortOpts struct {
	Status         string   `q:"status"`
	Name           string   `q:"name"`
	Description    string   `q:"description"`
	AdminStateUp   *bool    `q:"admin_state_up"`
	NetworkID      string   `q:"network_id"`
	TenantID       string   `q:"tenant_id"`
	ProjectID      string   `q:"project_id"`
	DeviceOwner    string   `q:"device_owner"`
	MACAddress     string   `q:"mac_address"`
	ID             string   `q:"id"`
	DeviceID       string   `q:"device_id"`
	Limit          int      `q:"limit"`
	Marker         string   `q:"marker"`
	SortKey        string   `q:"sort_key"`
	SortDir        string   `q:"sort_dir"`
	Tags           string   `q:"tags"`
	TagsAny        string   `q:"tags-any"`
	NotTags        string   `q:"not-tags"`
	NotTagsAny     string   `q:"not-tags-any"`
	SecurityGroups []string `q:"security_groups"`
	FixedIPs       []FixedIPOpts
}

type ListPortsResponse struct {
	Ports []Port `json:"ports"`
}

type CreateExtraDHCPOpt struct {
	OptName   string          `json:"opt_name" required:"true"`
	OptValue  string          `json:"opt_value" required:"true"`
	IPVersion types.IPVersion `json:"ip_version,omitempty"`
}

type UpdateExtraDHCPOpt struct {
	OptName   string          `json:"opt_name" required:"true"`
	OptValue  *string         `json:"opt_value"`
	IPVersion types.IPVersion `json:"ip_version,omitempty"`
}

type ExtraDHCPOptsExt struct {
	ExtraDHCPOpts []ExtraDHCPOpt `json:"extra_dhcp_opts"`
}

type ExtraDHCPOpt struct {
	OptName   string `json:"opt_name"`
	OptValue  string `json:"opt_value"`
	IPVersion int    `json:"ip_version"`
}
type PortSecurityExt struct {
	PortSecurityEnabled bool `json:"port_security_enabled"`
}

type PortsBindingExt struct {
	HostID     string         `json:"binding:host_id"`
	VIFDetails map[string]any `json:"binding:vif_details"`
	VIFType    string         `json:"binding:vif_type"`
	VNICType   string         `json:"binding:vnic_type"`
	Profile    map[string]any `json:"binding:profile"`
}

type PortDNSExt struct {
	DNSName       string              `json:"dns_name"`
	DNSAssignment []map[string]string `json:"dns_assignment"`
}

type QoSPolicyExt struct {
	QoSPolicyID string `json:"qos_policy_id"`
}

type IP struct {
	SubnetID  string `json:"subnet_id"`
	IPAddress string `json:"ip_address,omitempty"`
}

type UpdatePortOpts struct {
	Name                  *string            `json:"name,omitempty"`
	Description           *string            `json:"description,omitempty"`
	AdminStateUp          *bool              `json:"admin_state_up,omitempty"`
	FixedIPs              any                `json:"fixed_ips,omitempty"`
	DeviceID              *string            `json:"device_id,omitempty"`
	DeviceOwner           *string            `json:"device_owner,omitempty"`
	SecurityGroups        *[]string          `json:"security_groups,omitempty"`
	AllowedAddressPairs   *[]AddressPair     `json:"allowed_address_pairs,omitempty"`
	PropagateUplinkStatus *bool              `json:"propagate_uplink_status,omitempty"`
	ValueSpecs            *map[string]string `json:"value_specs,omitempty"`
	RevisionNumber        *int               `json:"-" h:"If-Match"`
	VirtualIp             *bool              `json:"virtual_ip,omitempty"`
}

type UpdatePortRequest struct {
	UpdatePortOpts
	PortSecurityEnabled *bool                `json:"port_security_enabled,omitempty"`
	ExtraDHCPOpts       []UpdateExtraDHCPOpt `json:"extra_dhcp_opts,omitempty"`
	DNSName             *string              `json:"dns_name,omitempty"`
	QoSPolicyID         *string              `json:"qos_policy_id,omitempty"`
	HostID              *string              `json:"binding:host_id,omitempty"`
	VNICType            string               `json:"binding:vnic_type,omitempty"`
	Profile             map[string]any       `json:"binding:profile,omitempty"`
}
