package dto

// PeeringConnection matches the iac-proxy-v2 PeeringConnection proto message.
type PeeringConnection struct {
	ID            string `json:"id"`
	Name          string `json:"name"`
	Status        string `json:"status"`        // lifecycle: active, creating, deleting, deleted, error
	PeeringStatus string `json:"peeringStatus"` // peering: established, deleted, unknown
	SrcVpcID      string `json:"srcVpcId"`
	SrcVpcCIDR    string `json:"srcVpcCidr"`
	DestVpcID     string `json:"destVpcId"`
	DestVpcCIDR   string `json:"destVpcCidr"`
	CreatedAt     string `json:"createdAt"`
}

// CreatePeeringConnectionRequest matches the iac-proxy-v2 CreatePeeringConnectionRequest proto message.
type CreatePeeringConnectionRequest struct {
	SrcVpcID    string `json:"srcVpcId"`
	DestVpcID   string `json:"destVpcId"`
	Description string `json:"description,omitempty"`
}

// UpdatePeeringConnectionRequest matches the iac-proxy-v2 UpdatePeeringConnectionRequest proto message.
type UpdatePeeringConnectionRequest struct {
	Name string `json:"name"`
}

// PeeringConnectionResponse matches the iac-proxy-v2 PeeringConnectionResponse proto message.
type PeeringConnectionResponse struct {
	PeeringConnection PeeringConnection `json:"peeringConnection"`
}

// ListPeeringConnectionsResponse matches the iac-proxy-v2 ListPeeringConnectionsResponse proto message.
type ListPeeringConnectionsResponse struct {
	PeeringConnections []PeeringConnection `json:"peeringConnections"`
}
