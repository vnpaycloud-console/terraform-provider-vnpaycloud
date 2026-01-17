package dto

import "time"

type FloatingIP struct {
	ID                string    `json:"id"`
	Description       string    `json:"description"`
	FloatingNetworkID string    `json:"floating_network_id"`
	FloatingIP        string    `json:"floating_ip_address"`
	PortID            string    `json:"port_id"`
	FixedIP           string    `json:"fixed_ip_address"`
	TenantID          string    `json:"tenant_id"`
	UpdatedAt         time.Time `json:"-"`
	CreatedAt         time.Time `json:"-"`
	ProjectID         string    `json:"project_id"`
	Status            string    `json:"status"`
	RouterID          string    `json:"router_id"`
	Tags              []string  `json:"tags"`
	DNSName           string    `json:"dns_name"`
	DNSDomain         string    `json:"dns_domain"`
}

type GetFloatingIPResponse struct {
	FloatingIP FloatingIP `json:"floatingip"`
}

type CreateFloatingIPOpts struct {
	Description       string            `json:"description,omitempty"`
	FloatingNetworkID string            `json:"floating_network_id" required:"true"`
	FloatingIP        string            `json:"floating_ip_address,omitempty"`
	PortID            string            `json:"port_id,omitempty"`
	FixedIP           string            `json:"fixed_ip_address,omitempty"`
	SubnetID          string            `json:"subnet_id,omitempty"`
	TenantID          string            `json:"tenant_id,omitempty"`
	ProjectID         string            `json:"project_id,omitempty"`
	ValueSpecs        map[string]string `json:"value_specs,omitempty"`
	DNSName           string            `json:"dns_name,omitempty"`
	DNSDomain         string            `json:"dns_domain,omitempty"`
}

type CreateFloatingIPRequest struct {
	FloatingIP CreateFloatingIPOpts `json:"floatingip"`
}
type CreateFloatingIPResponse struct {
	FloatingIP FloatingIP `json:"floatingip"`
}

type ListFloatingIPOpts struct {
	ID                string `q:"id"`
	Description       string `q:"description"`
	FloatingNetworkID string `q:"floating_network_id"`
	PortID            string `q:"port_id"`
	FixedIP           string `q:"fixed_ip_address"`
	FloatingIP        string `q:"floating_ip_address"`
	TenantID          string `q:"tenant_id"`
	ProjectID         string `q:"project_id"`
	Limit             int    `q:"limit"`
	Marker            string `q:"marker"`
	SortKey           string `q:"sort_key"`
	SortDir           string `q:"sort_dir"`
	RouterID          string `q:"router_id"`
	Status            string `q:"status"`
	Tags              string `q:"tags"`
	TagsAny           string `q:"tags-any"`
	NotTags           string `q:"not-tags"`
	NotTagsAny        string `q:"not-tags-any"`
}

type ListFloatingIPResponse struct {
	FloatingIPs []FloatingIP `json:"floatingips"`
}

type UpdateFloatingIPOpts struct {
	Description *string `json:"description,omitempty"`
	PortID      *string `json:"port_id,omitempty"`
	FixedIP     string  `json:"fixed_ip_address,omitempty"`
}

type UpdateFloatingIPRequest struct {
	FloatingIP UpdateFloatingIPOpts `json:"floatingip"`
}

type UpdateFloatingIPResponse struct {
	FloatingIP FloatingIP `json:"floatingip"`
}
