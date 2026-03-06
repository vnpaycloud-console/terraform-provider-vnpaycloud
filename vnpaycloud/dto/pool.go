package dto

// Pool matches the iac-proxy-v2 Pool proto message.
type Pool struct {
	ID          string       `json:"id"`
	Name        string       `json:"name"`
	ListenerID  string       `json:"listenerId"`
	LBAlgorithm string       `json:"lbAlgorithm"`
	Protocol    string       `json:"protocol"`
	Members     []PoolMember `json:"members"`
	Status      string       `json:"status"`
	CreatedAt   string       `json:"createdAt"`
}

// PoolMember matches the iac-proxy-v2 PoolMember proto message.
type PoolMember struct {
	ID           string `json:"id"`
	Name         string `json:"name,omitempty"`
	Address      string `json:"address"`
	ProtocolPort int    `json:"protocolPort"`
	Weight       int    `json:"weight"`
	Status       string `json:"status"`
}

// CreatePoolRequest matches the iac-proxy-v2 CreatePoolRequest proto message.
// project_id is passed via URL path.
type CreatePoolRequest struct {
	Name        string       `json:"name"`
	ListenerID  string       `json:"listenerId"`
	LBAlgorithm string       `json:"lbAlgorithm"`
	Protocol    string       `json:"protocol"`
	Members     []PoolMember `json:"members,omitempty"`
}

// UpdatePoolRequest matches the iac-proxy-v2 UpdatePoolRequest proto message.
// project_id and id are passed via URL path.
type UpdatePoolRequest struct {
	Name        string       `json:"name,omitempty"`
	LBAlgorithm string       `json:"lbAlgorithm,omitempty"`
	Members     []PoolMember `json:"members,omitempty"`
}

// PoolResponse matches the iac-proxy-v2 PoolResponse proto message.
type PoolResponse struct {
	Pool Pool `json:"pool"`
}

// ListPoolsResponse matches the iac-proxy-v2 ListPoolsResponse proto message.
type ListPoolsResponse struct {
	Pools []Pool `json:"pools"`
}
