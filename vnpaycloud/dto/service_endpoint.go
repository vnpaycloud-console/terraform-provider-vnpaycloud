package dto

// ServiceEndpoint matches the iac-proxy-v2 ServiceEndpoint proto message.
type ServiceEndpoint struct {
	ID                 string   `json:"id"`
	Name               string   `json:"name"`
	Description        string   `json:"description"`
	ProviderID         string   `json:"providerId"`
	ServiceID          string   `json:"serviceId"`
	ServiceGatewayID   string   `json:"serviceGatewayId"`
	Port               int      `json:"port"`
	AllowedCIDRs       []string `json:"allowedCidrs"`
	ListenerID         string   `json:"listenerId"`
	PoolID             string   `json:"poolId"`
	HealthMonitorID    string   `json:"healthMonitorId"`
	PoolMemberIDs      []string `json:"poolMemberIds"`
	OperatingStatus    string   `json:"operatingStatus"`
	ProvisioningStatus string   `json:"provisioningStatus"`
	ZoneID             string   `json:"zoneId"`
	Status             string   `json:"status"`
	CreatedAt          string   `json:"createdAt"`
	ProjectID          string   `json:"projectId"`
}

// CreateServiceEndpointRequest matches the iac-proxy-v2 CreateServiceEndpointRequest proto message.
// project_id is passed via URL path.
type CreateServiceEndpointRequest struct {
	Name             string   `json:"name"`
	Description      string   `json:"description,omitempty"`
	ProviderID       string   `json:"providerId"`
	ServiceID        string   `json:"serviceId"`
	ServiceGatewayID string   `json:"serviceGatewayId"`
	Port             int      `json:"port"`
	AllowedCIDRs     []string `json:"allowedCidrs,omitempty"`
}

// UpdateServiceEndpointRequest matches the iac-proxy-v2 UpdateServiceEndpointRequest proto message.
// project_id and id are passed via URL path.
type UpdateServiceEndpointRequest struct {
	Name         string   `json:"name,omitempty"`
	Description  string   `json:"description,omitempty"`
	AllowedCIDRs []string `json:"allowedCidrs"`
}

// ServiceEndpointResponse matches the iac-proxy-v2 ServiceEndpointResponse proto message.
type ServiceEndpointResponse struct {
	ServiceEndpoint ServiceEndpoint `json:"serviceEndpoint"`
}

// ListServiceEndpointsResponse matches the iac-proxy-v2 ListServiceEndpointsResponse proto message.
type ListServiceEndpointsResponse struct {
	ServiceEndpoints []ServiceEndpoint `json:"serviceEndpoints"`
}

// ServiceProvider matches the iac-proxy-v2 ServiceProvider proto message.
type ServiceProvider struct {
	ID     string `json:"id"`
	Name   string `json:"name"`
	Status string `json:"status"`
}

// ListServiceProvidersResponse matches the iac-proxy-v2 ListServiceProvidersResponse proto message.
type ListServiceProvidersResponse struct {
	Providers []ServiceProvider `json:"providers"`
}

// Service matches the iac-proxy-v2 Service proto message.
type Service struct {
	ID            string `json:"id"`
	Name          string `json:"name"`
	Description   string `json:"description"`
	ProviderID    string `json:"providerId"`
	ZoneID        string `json:"zoneId"`
	ServiceDomain string `json:"serviceDomain"`
	Status        string `json:"status"`
}

// ListServicesResponse matches the iac-proxy-v2 ListServicesResponse proto message.
type ListServicesResponse struct {
	Services []Service `json:"services"`
}
