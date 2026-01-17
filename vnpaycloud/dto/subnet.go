package dto

import "terraform-provider-vnpaycloud/vnpaycloud/types"

type Subnet struct {
	ID                string           `json:"id"`
	NetworkID         string           `json:"network_id"`
	Name              string           `json:"name"`
	Description       string           `json:"description"`
	IPVersion         int              `json:"ip_version"`
	CIDR              string           `json:"cidr"`
	GatewayIP         string           `json:"gateway_ip"`
	DNSNameservers    []string         `json:"dns_nameservers"`
	DNSPublishFixedIP bool             `json:"dns_publish_fixed_ip"`
	ServiceTypes      []string         `json:"service_types"`
	AllocationPools   []AllocationPool `json:"allocation_pools"`
	HostRoutes        []HostRoute      `json:"host_routes"`
	EnableDHCP        bool             `json:"enable_dhcp"`
	TenantID          string           `json:"tenant_id"`
	ProjectID         string           `json:"project_id"`
	IPv6AddressMode   string           `json:"ipv6_address_mode"`
	IPv6RAMode        string           `json:"ipv6_ra_mode"`
	SubnetPoolID      string           `json:"subnetpool_id"`
	Tags              []string         `json:"tags"`
	RevisionNumber    int              `json:"revision_number"`
	VPCID             string           `json:"vpc_id"`
}

type GetSubnetResponse struct {
	Subnet Subnet `json:"subnet"`
}

type SubnetCreateOpts struct {
	NetworkID         string            `json:"network_id" required:"true"`
	CIDR              string            `json:"cidr,omitempty"`
	Name              string            `json:"name,omitempty"`
	Description       string            `json:"description,omitempty"`
	TenantID          string            `json:"tenant_id,omitempty"`
	ProjectID         string            `json:"project_id,omitempty"`
	AllocationPools   []AllocationPool  `json:"allocation_pools,omitempty"`
	GatewayIP         *string           `json:"gateway_ip,omitempty"`
	IPVersion         types.IPVersion   `json:"ip_version,omitempty"`
	EnableDHCP        *bool             `json:"enable_dhcp,omitempty"`
	DNSNameservers    []string          `json:"dns_nameservers,omitempty"`
	DNSPublishFixedIP *bool             `json:"dns_publish_fixed_ip,omitempty"`
	ServiceTypes      []string          `json:"service_types,omitempty"`
	HostRoutes        []HostRoute       `json:"host_routes,omitempty"`
	IPv6AddressMode   string            `json:"ipv6_address_mode,omitempty"`
	IPv6RAMode        string            `json:"ipv6_ra_mode,omitempty"`
	SubnetPoolID      string            `json:"subnetpool_id,omitempty"`
	Prefixlen         int               `json:"prefixlen,omitempty"`
	VPCID             string            `json:"vpc_id,omitempty"`
	ValueSpecs        map[string]string `json:"value_specs,omitempty"`
}

type CreateSubnetRequest struct {
	Subnet SubnetCreateOpts `json:"subnet"`
}

type CreateSubnetResponse struct {
	Subnet Subnet `json:"subnet"`
}

type HostRoute struct {
	DestinationCIDR string `json:"destination"`
	NextHop         string `json:"nexthop"`
}

type AllocationPool struct {
	Start string `json:"start"`
	End   string `json:"end"`
}

type ListSubnetParams struct {
	Name              string `q:"name"`
	Description       string `q:"description"`
	DNSPublishFixedIP *bool  `q:"dns_publish_fixed_ip"`
	EnableDHCP        *bool  `q:"enable_dhcp"`
	NetworkID         string `q:"network_id"`
	TenantID          string `q:"tenant_id"`
	ProjectID         string `q:"project_id"`
	IPVersion         int    `q:"ip_version"`
	GatewayIP         string `q:"gateway_ip"`
	CIDR              string `q:"cidr"`
	IPv6AddressMode   string `q:"ipv6_address_mode"`
	IPv6RAMode        string `q:"ipv6_ra_mode"`
	ID                string `q:"id"`
	SubnetPoolID      string `q:"subnetpool_id"`
	Limit             int    `q:"limit"`
	Marker            string `q:"marker"`
	SortKey           string `q:"sort_key"`
	SortDir           string `q:"sort_dir"`
	Tags              string `q:"tags"`
	TagsAny           string `q:"tags-any"`
	NotTags           string `q:"not-tags"`
	NotTagsAny        string `q:"not-tags-any"`
	VPCID             string `q:"vpc_id"`
}

type ListSubnetResponse struct {
	Subnets []Subnet `json:"subnets"`
}
