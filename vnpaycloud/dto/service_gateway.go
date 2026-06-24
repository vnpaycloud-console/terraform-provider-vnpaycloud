package dto

// ServiceGateway matches the backend ServiceGateway proto message.
type ServiceGateway struct {
	ID                 string `json:"id"`
	Name               string `json:"name"`
	Description        string `json:"description"`
	VPCID              string `json:"vpcId"`
	SubnetID           string `json:"subnetId"`
	FlavorID           string `json:"flavorId"`
	LoadBalancerID     string `json:"loadBalancerId"`
	VipAddress         string `json:"vipAddress"`
	PortID             string `json:"portId"`
	AllowedICMP        bool   `json:"allowedIcmp"`
	OperatingStatus    string `json:"operatingStatus"`
	ProvisioningStatus string `json:"provisioningStatus"`
	ZoneID             string `json:"zoneId"`
	Status             string `json:"status"`
	CreatedAt          string `json:"createdAt"`
	ProjectID          string `json:"projectId"`
}

// CreateServiceGatewayRequest matches the backend CreateServiceGatewayRequest proto message.
// project_id is passed via URL path.
type CreateServiceGatewayRequest struct {
	Name        string `json:"name"`
	Description string `json:"description,omitempty"`
	VPCID       string `json:"vpcId,omitempty"`
	SubnetID    string `json:"subnetId"`
	FlavorID    string `json:"flavorId"`
}

// UpdateServiceGatewayRequest matches the backend UpdateServiceGatewayRequest proto message.
// project_id and id are passed via URL path.
type UpdateServiceGatewayRequest struct {
	Name        string `json:"name,omitempty"`
	Description string `json:"description,omitempty"`
}

// SetServiceGatewayICMPRequest matches the backend SetServiceGatewayICMPRequest proto message.
// project_id and id are passed via URL path.
type SetServiceGatewayICMPRequest struct {
	AllowedICMP bool `json:"allowedIcmp"`
}

// ChangeFlavorServiceGatewayRequest matches the backend ChangeFlavorServiceGatewayRequest proto message.
// project_id and id are passed via URL path.
type ChangeFlavorServiceGatewayRequest struct {
	FlavorID string `json:"flavorId"`
}

// ServiceGatewayResponse matches the backend ServiceGatewayResponse proto message.
type ServiceGatewayResponse struct {
	ServiceGateway ServiceGateway `json:"serviceGateway"`
}

// ListServiceGatewaysResponse matches the backend ListServiceGatewaysResponse proto message.
type ListServiceGatewaysResponse struct {
	ServiceGateways []ServiceGateway `json:"serviceGateways"`
}

// ServiceGatewayFlavor matches the backend ServiceGatewayFlavor proto message.
type ServiceGatewayFlavor struct {
	ID          string `json:"id"`
	Name        string `json:"name"`
	Description string `json:"description"`
	ZoneID      string `json:"zoneId"`
}

// ListServiceGatewayFlavorsResponse matches the backend ListServiceGatewayFlavorsResponse proto message.
type ListServiceGatewayFlavorsResponse struct {
	Flavors []ServiceGatewayFlavor `json:"flavors"`
}
