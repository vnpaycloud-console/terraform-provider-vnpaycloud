package dto

// VPC matches the iac-proxy-v2 VPC proto message.
type VPC struct {
	ID          string   `json:"id"`
	Name        string   `json:"name"`
	Description string   `json:"description"`
	CIDR        string   `json:"cidr"`
	Status      string   `json:"status"`
	SubnetIDs   []string `json:"subnetIds"`
	EnableSnat  bool     `json:"enableSnat"`
	SnatAddress string   `json:"snatAddress"`
	CreatedAt   string   `json:"createdAt"`
	ProjectID   string   `json:"projectId"`
	ZoneID      string   `json:"zoneId"`
}

// CreateVPCRequest matches the iac-proxy-v2 CreateVPCRequest proto message.
// project_id is passed via URL path, not in the body.
type CreateVPCRequest struct {
	Name        string `json:"name"`
	Description string `json:"description,omitempty"`
	CIDR        string `json:"cidr"`
}

// UpdateVPCRequest matches the iac-proxy-v2 UpdateVPCRequest proto message.
// project_id and id are passed via URL path.
type UpdateVPCRequest struct {
	Name        string `json:"name,omitempty"`
	Description string `json:"description,omitempty"`
}

// VPCResponse matches the iac-proxy-v2 VPCResponse proto message.
type VPCResponse struct {
	VPC VPC `json:"vpc"`
}

// ListVPCsResponse matches the iac-proxy-v2 ListVPCsResponse proto message.
type ListVPCsResponse struct {
	VPCs []VPC `json:"vpcs"`
}

// SetVPCRouterSNATRequest matches the iac-proxy-v2 SetVPCRouterSNATRequest proto message.
type SetVPCRouterSNATRequest struct {
	EnableSnat bool `json:"enableSnat"`
}
