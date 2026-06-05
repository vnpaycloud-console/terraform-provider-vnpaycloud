package dto

// SessionPersistence matches the iac-proxy-v2 SessionPersistence proto message.
type SessionPersistence struct {
	Type       string `json:"type"`
	CookieName string `json:"cookieName,omitempty"`
}

// Pool matches the iac-proxy-v2 Pool proto message.
type Pool struct {
	ID                 string              `json:"id"`
	Name               string              `json:"name"`
	Description        string              `json:"description"`
	LoadBalancerID     string              `json:"loadBalancerId"`
	ListenerID         string              `json:"listenerId"`
	LBAlgorithm        string              `json:"lbAlgorithm"`
	Protocol           string              `json:"protocol"`
	SessionPersistence *SessionPersistence `json:"sessionPersistence,omitempty"`
	TlsEnabled         bool                `json:"tlsEnabled"`
	Members            []PoolMember        `json:"members"`
	Status             string              `json:"status"`
	ProvisioningStatus string              `json:"provisioningStatus"`
	OperatingStatus    string              `json:"operatingStatus"`
	CreatedAt          string              `json:"createdAt"`
}

// PoolMember matches the iac-proxy-v2 PoolMember proto message.
type PoolMember struct {
	ID                 string `json:"id"`
	Name               string `json:"name,omitempty"`
	Address            string `json:"address"`
	ProtocolPort       int    `json:"protocolPort"`
	Weight             int    `json:"weight"`
	Status             string `json:"status"`
	ProvisioningStatus string `json:"provisioningStatus"`
	OperatingStatus    string `json:"operatingStatus"`
}

// CreatePoolRequest matches the iac-proxy-v2 CreatePoolRequest proto message.
// project_id is passed via URL path.
type CreatePoolRequest struct {
	Name               string              `json:"name"`
	Description        string              `json:"description,omitempty"`
	LoadBalancerID     string              `json:"loadBalancerId"`
	ListenerID         string              `json:"listenerId,omitempty"`
	LBAlgorithm        string              `json:"lbAlgorithm"`
	Protocol           string              `json:"protocol"`
	SessionPersistence *SessionPersistence `json:"sessionPersistence,omitempty"`
	TlsEnabled         bool                `json:"tlsEnabled,omitempty"`
	Members            []PoolMember        `json:"members,omitempty"`
}

// UpdatePoolRequest matches the iac-proxy-v2 UpdatePoolRequest proto message.
// project_id and id are passed via URL path.
type UpdatePoolRequest struct {
	Name               string              `json:"name,omitempty"`
	Description        string              `json:"description,omitempty"`
	LBAlgorithm        string              `json:"lbAlgorithm,omitempty"`
	SessionPersistence *SessionPersistence `json:"sessionPersistence,omitempty"`
	TlsEnabled         bool                `json:"tlsEnabled"`
	Members            []PoolMember        `json:"members,omitempty"`
}

// PoolResponse matches the iac-proxy-v2 PoolResponse proto message.
type PoolResponse struct {
	Pool Pool `json:"pool"`
}

// ListPoolsResponse matches the iac-proxy-v2 ListPoolsResponse proto message.
type ListPoolsResponse struct {
	Pools []Pool `json:"pools"`
}
