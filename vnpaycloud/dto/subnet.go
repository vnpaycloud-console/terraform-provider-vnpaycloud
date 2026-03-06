package dto

type Subnet struct {
	ID           string `json:"id"`
	Name         string `json:"name"`
	VpcID        string `json:"vpcId"`
	CIDR         string `json:"cidr"`
	GatewayIP    string `json:"gatewayIp"`
	EnableDHCP   bool   `json:"enableDhcp"`
	EnableSnat   bool   `json:"enableSnat"`
	ExternalIpID string `json:"externalIpId"`
	UsedByK8S    bool   `json:"usedByK8s"`
	Status       string `json:"status"`
	CreatedAt    string `json:"createdAt"`
	ProjectID    string `json:"projectId"`
	ZoneID       string `json:"zoneId"`
}

type CreateSubnetRequest struct {
	Name       string `json:"name"`
	VpcID      string `json:"vpcId"`
	NetworkID  string `json:"networkId,omitempty"`
	CIDR       string `json:"cidr"`
	GatewayIP  string `json:"gatewayIp,omitempty"`
	EnableDHCP bool   `json:"enableDhcp"`
	UsedByK8S  bool   `json:"usedByK8s"`
}

type SubnetResponse struct {
	Subnet Subnet `json:"subnet"`
}

type ListSubnetsResponse struct {
	Subnets []Subnet `json:"subnets"`
}

// EnableSubnetSNATRequest matches the iac-proxy-v2 EnableSubnetSNATRequest proto message.
type EnableSubnetSNATRequest struct {
	FloatingIpID string `json:"floatingIpId"`
}
