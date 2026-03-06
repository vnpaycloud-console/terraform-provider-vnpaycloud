package dto

// Listener matches the iac-proxy-v2 Listener proto message.
type Listener struct {
	ID             string `json:"id"`
	Name           string `json:"name"`
	LoadBalancerID string `json:"loadBalancerId"`
	Protocol       string `json:"protocol"`
	ProtocolPort   int    `json:"protocolPort"`
	DefaultPoolID  string `json:"defaultPoolId"`
	Status         string `json:"status"`
	CreatedAt      string `json:"createdAt"`
}

// CreateListenerRequest matches the iac-proxy-v2 CreateListenerRequest proto message.
// project_id is passed via URL path.
type CreateListenerRequest struct {
	Name           string `json:"name"`
	LoadBalancerID string `json:"loadBalancerId"`
	Protocol       string `json:"protocol"`
	ProtocolPort   int    `json:"protocolPort"`
	DefaultPoolID  string `json:"defaultPoolId,omitempty"`
}

// UpdateListenerRequest matches the iac-proxy-v2 UpdateListenerRequest proto message.
// project_id and id are passed via URL path.
type UpdateListenerRequest struct {
	Name          string `json:"name,omitempty"`
	DefaultPoolID string `json:"defaultPoolId,omitempty"`
}

// ListenerResponse matches the iac-proxy-v2 ListenerResponse proto message.
type ListenerResponse struct {
	Listener Listener `json:"listener"`
}

// ListListenersResponse matches the iac-proxy-v2 ListListenersResponse proto message.
type ListListenersResponse struct {
	Listeners []Listener `json:"listeners"`
}
