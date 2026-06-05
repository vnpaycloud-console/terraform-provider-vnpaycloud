package dto

// LoadBalancer matches the iac-proxy-v2 LoadBalancer proto message.
type LoadBalancer struct {
	ID                 string `json:"id"`
	Name               string `json:"name"`
	Description        string `json:"description"`
	Flavor             string `json:"flavor"`
	VipAddress         string `json:"vipAddress"`
	VipSubnetID        string `json:"vipSubnetId"`
	Status             string `json:"status"`
	ProvisioningStatus string `json:"provisioningStatus"`
	OperatingStatus    string `json:"operatingStatus"`
	CreatedAt          string `json:"createdAt"`
	FloatingIPID       string `json:"floatingIpId"`
}

// CreateLoadBalancerRequest matches the iac-proxy-v2 CreateLoadBalancerRequest proto message.
// project_id is passed via URL path.
type CreateLoadBalancerRequest struct {
	Name         string `json:"name"`
	Description  string `json:"description,omitempty"`
	SubnetID     string `json:"subnetId"`
	Flavor       string `json:"flavor"`
	External     bool   `json:"external"`
	FloatingIPID string `json:"floatingIpId,omitempty"`
}

// UpdateLoadBalancerRequest matches the iac-proxy-v2 UpdateLoadBalancerRequest proto message.
// project_id and id are passed via URL path.
type UpdateLoadBalancerRequest struct {
	Name        string `json:"name,omitempty"`
	Description string `json:"description,omitempty"`
}

// ChangeFlavorLoadBalancerRequest matches the iac-proxy-v2 ChangeFlavorLoadBalancerRequest proto message.
// project_id and id are passed via URL path. flavor is the flavor name (resolved to ID by the backend).
type ChangeFlavorLoadBalancerRequest struct {
	Flavor string `json:"flavor"`
}

// LoadBalancerResponse matches the iac-proxy-v2 LoadBalancerResponse proto message.
type LoadBalancerResponse struct {
	LoadBalancer LoadBalancer `json:"loadBalancer"`
}

// ListLoadBalancersResponse matches the iac-proxy-v2 ListLoadBalancersResponse proto message.
type ListLoadBalancersResponse struct {
	LoadBalancers []LoadBalancer `json:"loadBalancers"`
}

// LoadBalancerFlavor matches the iac-proxy-v2 LoadBalancerFlavor proto message.
type LoadBalancerFlavor struct {
	ID          string `json:"id"`
	Name        string `json:"name"`
	Description string `json:"description"`
	ZoneID      string `json:"zoneId"`
}

// ListLoadBalancerFlavorsResponse matches the iac-proxy-v2 ListLoadBalancerFlavorsResponse proto message.
type ListLoadBalancerFlavorsResponse struct {
	Flavors []LoadBalancerFlavor `json:"flavors"`
}
