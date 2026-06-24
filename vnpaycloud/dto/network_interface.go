package dto

// NetworkInterfaceAddressPair matches the backend NetworkInterfaceAddressPair proto message.
type NetworkInterfaceAddressPair struct {
	IPAddress  string `json:"ipAddress"`
	MACAddress string `json:"macAddress"`
}

// NetworkInterface matches the backend NetworkInterface proto message.
type NetworkInterface struct {
	ID                  string                        `json:"id"`
	Name                string                        `json:"name"`
	NetworkID           string                        `json:"networkId"`
	SubnetID            string                        `json:"subnetId"`
	IPAddress           string                        `json:"ipAddress"`
	MACAddress          string                        `json:"macAddress"`
	Status              string                        `json:"status"`
	SecurityGroups      []string                      `json:"securityGroups"`
	PortSecurityEnabled bool                          `json:"portSecurityEnabled"`
	AllowedAddressPairs []NetworkInterfaceAddressPair `json:"allowedAddressPairs"`
	NetworkType         string                        `json:"networkType"`
	Description         string                        `json:"description"`
	Reserved            bool                          `json:"reserved"`
	VirtualIP           bool                          `json:"virtualIp"`
	CreatedAt           string                        `json:"createdAt"`
	ProjectID           string                        `json:"projectId"`
	ZoneID              string                        `json:"zoneId"`
}

// CreateNetworkInterfaceRequest matches the backend CreateNetworkInterfaceRequest proto message.
// project_id is passed via URL path, not in the body.
type CreateNetworkInterfaceRequest struct {
	Name        string `json:"name"`
	SubnetID    string `json:"subnetId"`
	IPAddress   string `json:"ipAddress,omitempty"`
	Description string `json:"description,omitempty"`
	Reserved    bool   `json:"reserved,omitempty"`
	VirtualIP   bool   `json:"virtualIp,omitempty"`
}

type UpdateNetworkInterfaceReservedRequest struct {
	Reserved    bool   `json:"reserved"`
	Description string `json:"description,omitempty"`
}

type UpdateNetworkInterfaceVirtualIpRequest struct {
	VirtualIP bool `json:"virtualIp"`
}

type UpdateNetworkInterfaceAllowedAddressPairsRequest struct {
	AllowedAddressPairs []NetworkInterfaceAddressPair `json:"allowedAddressPairs"`
}

type UpdateNetworkInterfacePortSecurityRequest struct {
	PortSecurityEnabled bool `json:"portSecurityEnabled"`
}

type UpdateNetworkInterfaceSecurityGroupsRequest struct {
	SecurityGroupIDs []string `json:"securityGroupIds"`
}

// AttachNetworkInterfaceRequest matches the backend AttachNetworkInterfaceRequest proto message.
// project_id and id are passed via URL path.
type AttachNetworkInterfaceRequest struct {
	ServerID string `json:"serverId"`
}

// DetachNetworkInterfaceRequest matches the backend DetachNetworkInterfaceRequest proto message.
// project_id and id are passed via URL path.
type DetachNetworkInterfaceRequest struct {
	ServerID string `json:"serverId"`
}

// NetworkInterfaceResponse matches the backend NetworkInterfaceResponse proto message.
type NetworkInterfaceResponse struct {
	NetworkInterface NetworkInterface `json:"networkInterface"`
}

// ListNetworkInterfacesResponse matches the backend ListNetworkInterfacesResponse proto message.
type ListNetworkInterfacesResponse struct {
	NetworkInterfaces []NetworkInterface `json:"networkInterfaces"`
}
