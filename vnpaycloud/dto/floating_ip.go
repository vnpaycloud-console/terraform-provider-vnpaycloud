package dto

// FloatingIP matches the iac-proxy-v2 FloatingIP proto message.
type FloatingIP struct {
	ID           string `json:"id"`
	Address      string `json:"address"`
	Status       string `json:"status"`
	PortID       string `json:"portId"`
	Type         string `json:"type"`
	VpcID        string `json:"vpcId"`
	InstanceID   string `json:"instanceId"`
	InstanceName string `json:"instanceName"`
	CreatedAt    string `json:"createdAt"`
	ProjectID    string `json:"projectId"`
	ZoneID       string `json:"zoneId"`
}

// CreateFloatingIPRequest matches the iac-proxy-v2 CreateFloatingIPRequest proto message.
// project_id is passed via URL path, not in the body.
// zone is resolved server-side from project_id.
type CreateFloatingIPRequest struct{}

// AssociateFloatingIPRequest matches the iac-proxy-v2 AssociateFloatingIPRequest proto message.
// project_id and id are passed via URL path.
// port_id and vpc_id are mutually exclusive.
type AssociateFloatingIPRequest struct {
	PortID string `json:"portId,omitempty"`
	VpcID  string `json:"vpcId,omitempty"`
}

// DisassociateFloatingIPRequest is an empty body.
// project_id and id are passed via URL path.
type DisassociateFloatingIPRequest struct{}

// FloatingIPResponse matches the iac-proxy-v2 FloatingIPResponse proto message.
type FloatingIPResponse struct {
	FloatingIP FloatingIP `json:"floatingIp"`
}

// ListFloatingIPsResponse matches the iac-proxy-v2 ListFloatingIPsResponse proto message.
type ListFloatingIPsResponse struct {
	FloatingIPs []FloatingIP `json:"floatingIps"`
}
