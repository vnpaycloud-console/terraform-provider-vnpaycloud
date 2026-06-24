package dto

// FloatingIP matches the backend FloatingIP proto message.
type FloatingIP struct {
	ID           string `json:"id"`
	Address      string `json:"address"`
	Status       string `json:"status"`
	PortID       string `json:"portId"`
	Type         string `json:"type"`
	VpcID        string `json:"vpcId"`
	FixedIP      string `json:"fixedIp"`
	InstanceID   string `json:"instanceId"`
	InstanceName string `json:"instanceName"`
	CreatedAt    string `json:"createdAt"`
	ProjectID    string `json:"projectId"`
	ZoneID       string `json:"zoneId"`
}

// CreateFloatingIPRequest matches the backend CreateFloatingIPRequest proto message.
// project_id is passed via URL path, not in the body.
// zone is resolved server-side from project_id.
type CreateFloatingIPRequest struct{}

// AssociateFloatingIPRequest matches the backend AssociateFloatingIPRequest proto message.
// project_id and id are passed via URL path.
// port_id and vpc_id are mutually exclusive.
type AssociateFloatingIPRequest struct {
	PortID string `json:"portId,omitempty"`
	VpcID  string `json:"vpcId,omitempty"`
}

// DisassociateFloatingIPRequest is an empty body.
// project_id and id are passed via URL path.
type DisassociateFloatingIPRequest struct{}

// FloatingIPResponse matches the backend FloatingIPResponse proto message.
type FloatingIPResponse struct {
	FloatingIP FloatingIP `json:"floatingIp"`
}

// ListFloatingIPsResponse matches the backend ListFloatingIPsResponse proto message.
type ListFloatingIPsResponse struct {
	FloatingIPs []FloatingIP `json:"floatingIps"`
}
