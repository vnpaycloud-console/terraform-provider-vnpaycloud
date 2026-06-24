package dto

// InternetGateway matches the backend InternetGateway proto message.
type InternetGateway struct {
	ID          string `json:"id"`
	Name        string `json:"name"`
	Description string `json:"description"`
	VPCID       string `json:"vpcId"`
	Status      string `json:"status"`
	CreatedAt   string `json:"createdAt"`
	ProjectID   string `json:"projectId"`
	ZoneID      string `json:"zoneId"`
}

// CreateInternetGatewayRequest matches the backend CreateInternetGatewayRequest proto message.
// project_id is passed via URL path.
type CreateInternetGatewayRequest struct {
	Name        string `json:"name"`
	Description string `json:"description,omitempty"`
}

// AttachInternetGatewayToVPCRequest matches the backend AttachInternetGatewayToVPCRequest proto message.
// project_id and id are passed via URL path.
type AttachInternetGatewayToVPCRequest struct {
	VPCID string `json:"vpcId"`
}

// DetachInternetGatewayFromVPCRequest matches the backend DetachInternetGatewayFromVPCRequest proto message.
// project_id and id are passed via URL path.
type DetachInternetGatewayFromVPCRequest struct {
	VPCID string `json:"vpcId"`
}

// InternetGatewayResponse matches the backend InternetGatewayResponse proto message.
type InternetGatewayResponse struct {
	InternetGateway InternetGateway `json:"internetGateway"`
}

// ListInternetGatewaysResponse matches the backend ListInternetGatewaysResponse proto message.
type ListInternetGatewaysResponse struct {
	InternetGateways []InternetGateway `json:"internetGateways"`
}
