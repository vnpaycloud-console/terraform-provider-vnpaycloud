package dto

// NetworkInterfaceAddressPair matches the iac-proxy-v2 NetworkInterfaceAddressPair proto message.
type NetworkInterfaceAddressPair struct {
	IPAddress  string `json:"ipAddress"`
	MACAddress string `json:"macAddress"`
}

// NetworkInterface matches the iac-proxy-v2 NetworkInterface proto message.
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
	CreatedAt           string                        `json:"createdAt"`
	ProjectID           string                        `json:"projectId"`
	ZoneID              string                        `json:"zoneId"`
}

// CreateNetworkInterfaceRequest matches the iac-proxy-v2 CreateNetworkInterfaceRequest proto message.
// project_id is passed via URL path, not in the body.
type CreateNetworkInterfaceRequest struct {
	Name        string `json:"name"`
	SubnetID    string `json:"subnetId"`
	IPAddress   string `json:"ipAddress,omitempty"`
	Description string `json:"description,omitempty"`
}

// AttachNetworkInterfaceRequest matches the iac-proxy-v2 AttachNetworkInterfaceRequest proto message.
// project_id and id are passed via URL path.
type AttachNetworkInterfaceRequest struct {
	ServerID string `json:"serverId"`
}

// DetachNetworkInterfaceRequest matches the iac-proxy-v2 DetachNetworkInterfaceRequest proto message.
// project_id and id are passed via URL path.
type DetachNetworkInterfaceRequest struct {
	ServerID string `json:"serverId"`
}

// NetworkInterfaceResponse matches the iac-proxy-v2 NetworkInterfaceResponse proto message.
type NetworkInterfaceResponse struct {
	NetworkInterface NetworkInterface `json:"networkInterface"`
}

// ListNetworkInterfacesResponse matches the iac-proxy-v2 ListNetworkInterfacesResponse proto message.
type ListNetworkInterfacesResponse struct {
	NetworkInterfaces []NetworkInterface `json:"networkInterfaces"`
}
