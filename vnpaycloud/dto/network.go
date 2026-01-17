package dto

import "time"

// Network represents, well, a network.
type Network struct {
	// UUID for the network
	ID string `json:"id"`

	// Human-readable name for the network. Might not be unique.
	Name string `json:"name"`

	// Description for the network
	Description string `json:"description"`

	// The administrative state of network. If false (down), the network does not
	// forward packets.
	AdminStateUp bool `json:"admin_state_up"`

	// Indicates whether network is currently operational. Possible values include
	// `ACTIVE', `DOWN', `BUILD', or `ERROR'. Plug-ins might define additional
	// values.
	Status string `json:"status"`

	// Subnets associated with this network.
	Subnets []string `json:"subnets"`

	// TenantID is the project owner of the network.
	TenantID string `json:"tenant_id"`

	// UpdatedAt and CreatedAt contain ISO-8601 timestamps of when the state of the
	// network last changed, and when it was created.
	UpdatedAt time.Time `json:"-"`
	CreatedAt time.Time `json:"-"`

	// ProjectID is the project owner of the network.
	ProjectID string `json:"project_id"`

	// Specifies whether the network resource can be accessed by any tenant.
	Shared bool `json:"shared"`

	// Availability zone hints groups network nodes that run services like DHCP, L3, FW, and others.
	// Used to make network resources highly available.
	AvailabilityZoneHints []string `json:"availability_zone_hints"`

	// Tags optionally set via extensions/attributestags
	Tags []string `json:"tags"`

	// RevisionNumber optionally set via extensions/standard-attr-revisions
	RevisionNumber int `json:"revision_number"`

	// Maximum Transmission Unit in bytes.
	MTU int `json:"mtu"`

	External bool `json:"router:external"`

	QoSPolicyID string `json:"qos_policy_id"`

	DNSDomain string `json:"dns_domain"`

	VLANTransparent bool `json:"vlan_transparent"`

	PortSecurityEnabled bool `json:"port_security_enabled"`

	Segments []Segments `json:"segments"`
}

type Segments struct {
	NetworkType     string `json:"provider:network_type"`
	PhysicalNetwork string `json:"provider:physical_network"`
	SegmentationID  int    `json:"provider:segmentation_id"`
}

type GetNetworkResponse struct {
	Network Network `json:"network"`
}

type CreateNetworkRequest struct {
	Network CreateNetworkOpts `json:"network,omitempty"`
}

type CreateNetworkOpts struct {
	Name         string `json:"name,omitempty"`
	Description  string `json:"description,omitempty"`
	AdminStateUp *bool  `json:"admin_state_up,omitempty"`
}

type CreateNetworkResponse struct {
	Network Network `json:"network"`
}

// ListOpts allows the filtering and sorting of paginated collections through
// the API. Filtering is achieved by passing in struct field values that map to
// the network attributes you want to see returned. SortKey allows you to sort
// by a particular network attribute. SortDir sets the direction, and is either
// `asc' or `desc'. Marker and Limit are used for pagination.
type ListNetworkParams struct {
	Status       string `q:"status"`
	Name         string `q:"name"`
	Description  string `q:"description"`
	AdminStateUp *bool  `q:"admin_state_up"`
	TenantID     string `q:"tenant_id"`
	ProjectID    string `q:"project_id"`
	Shared       *bool  `q:"shared"`
	ID           string `q:"id"`
	Marker       string `q:"marker"`
	Limit        int    `q:"limit"`
	SortKey      string `q:"sort_key"`
	SortDir      string `q:"sort_dir"`
	Tags         string `q:"tags"`
	TagsAny      string `q:"tags-any"`
	NotTags      string `q:"not-tags"`
	NotTagsAny   string `q:"not-tags-any"`
}

type ListNetworksResponse struct {
	Networks []Network `json:"networks"`
}
