package dto

// PrivateGateway matches the iac-proxy-v2 PrivateGateway proto message.
type PrivateGateway struct {
	ID             string `json:"id"`
	Name           string `json:"name"`
	Description    string `json:"description"`
	LoadBalancerID string `json:"loadBalancerId"`
	SubnetID       string `json:"subnetId"`
	FlavorID       string `json:"flavorId"`
	Status         string `json:"status"`
	CreatedAt      string `json:"createdAt"`
	ProjectID      string `json:"projectId"`
	ZoneID         string `json:"zoneId"`
}

// CreatePrivateGatewayRequest matches the iac-proxy-v2 CreatePrivateGatewayRequest proto message.
// project_id is passed via URL path.
type CreatePrivateGatewayRequest struct {
	Name        string `json:"name"`
	Description string `json:"description,omitempty"`
}

// UpdatePrivateGatewayRequest matches the iac-proxy-v2 UpdatePrivateGatewayRequest proto message.
// project_id and id are passed via URL path.
type UpdatePrivateGatewayRequest struct {
	Name        string `json:"name"`
	Description string `json:"description,omitempty"`
}

// PrivateGatewayResponse matches the iac-proxy-v2 PrivateGatewayResponse proto message.
type PrivateGatewayResponse struct {
	PrivateGateway PrivateGateway `json:"privateGateway"`
}

// ListPrivateGatewaysResponse matches the iac-proxy-v2 ListPrivateGatewaysResponse proto message.
type ListPrivateGatewaysResponse struct {
	PrivateGateways []PrivateGateway `json:"privateGateways"`
}
