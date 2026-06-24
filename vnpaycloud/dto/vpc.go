package dto

// VPC matches the backend VPC proto message.
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

// CreateVPCRequest matches the backend CreateVPCRequest proto message.
// project_id is passed via URL path, not in the body.
type CreateVPCRequest struct {
	Name        string `json:"name"`
	Description string `json:"description,omitempty"`
	CIDR        string `json:"cidr,omitempty"`
}

// UpdateVPCRequest matches the backend UpdateVPCRequest proto message.
// project_id and id are passed via URL path.
type UpdateVPCRequest struct {
	Name        string `json:"name,omitempty"`
	Description string `json:"description,omitempty"`
}

// VPCResponse matches the backend VPCResponse proto message.
type VPCResponse struct {
	VPC VPC `json:"vpc"`
}

// ListVPCsResponse matches the backend ListVPCsResponse proto message.
type ListVPCsResponse struct {
	VPCs []VPC `json:"vpcs"`
}

// SetVPCRouterSNATRequest matches the backend SetVPCRouterSNATRequest proto message.
type SetVPCRouterSNATRequest struct {
	EnableSnat bool `json:"enableSnat"`
}
