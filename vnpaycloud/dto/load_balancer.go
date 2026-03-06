package dto

// LoadBalancer matches the iac-proxy-v2 LoadBalancer proto message.
type LoadBalancer struct {
	ID          string   `json:"id"`
	Name        string   `json:"name"`
	Description string   `json:"description"`
	VipAddress  string   `json:"vipAddress"`
	VipSubnetID string   `json:"vipSubnetId"`
	Status      string   `json:"status"`
	ListenerIDs []string `json:"listenerIds"`
	CreatedAt   string   `json:"createdAt"`
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

// LoadBalancerResponse matches the iac-proxy-v2 LoadBalancerResponse proto message.
type LoadBalancerResponse struct {
	LoadBalancer LoadBalancer `json:"loadBalancer"`
}

// ListLoadBalancersResponse matches the iac-proxy-v2 ListLoadBalancersResponse proto message.
type ListLoadBalancersResponse struct {
	LoadBalancers []LoadBalancer `json:"loadBalancers"`
}
