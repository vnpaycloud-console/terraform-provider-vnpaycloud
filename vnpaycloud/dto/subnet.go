package dto

type Subnet struct {
	ID             string      `json:"id"`
	Name           string      `json:"name"`
	VpcID          string      `json:"vpcId"`
	CIDR           string      `json:"cidr"`
	GatewayIP      string      `json:"gatewayIp"`
	EnableDHCP     bool        `json:"enableDhcp"`
	EnableSnat     bool        `json:"enableSnat"`
	ExternalIpID   string      `json:"externalIpId"`
	UsedByK8S      bool        `json:"usedByK8s"`
	DNSNameservers []string    `json:"dnsNameservers,omitempty"`
	Routes         []HostRoute `json:"routes,omitempty"`
	Status         string      `json:"status"`
	CreatedAt      string      `json:"createdAt"`
	ProjectID      string      `json:"projectId"`
	ZoneID         string      `json:"zoneId"`
}

type HostRoute struct {
	Destination string `json:"destination"`
	Nexthop     string `json:"nexthop"`
}

type CreateSubnetRequest struct {
	Name           string   `json:"name"`
	VpcID          string   `json:"vpcId"`
	NetworkID      string   `json:"networkId,omitempty"`
	CIDR           string   `json:"cidr,omitempty"`
	GatewayIP      string   `json:"gatewayIp,omitempty"`
	EnableDHCP     bool     `json:"enableDhcp"`
	UsedByK8S      bool     `json:"usedByK8s"`
	UsedBySI       bool     `json:"usedBySi,omitempty"`
	DNSNameservers []string `json:"dnsNameservers,omitempty"`
}

type UpdateSubnetRequest struct {
	Name           string   `json:"name"`
	DNSNameservers []string `json:"dnsNameservers,omitempty"`
}

type UpdateSubnetRoutesRequest struct {
	Routes []HostRoute `json:"routes"`
}

type SubnetResponse struct {
	Subnet Subnet `json:"subnet"`
}

type ListSubnetsResponse struct {
	Subnets []Subnet `json:"subnets"`
}

// EnableSubnetSNATRequest matches the backend EnableSubnetSNATRequest proto message.
type EnableSubnetSNATRequest struct {
	FloatingIpID string `json:"floatingIpId"`
}
