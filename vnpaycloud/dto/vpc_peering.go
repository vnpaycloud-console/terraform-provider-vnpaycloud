package dto

// PeeringConnection matches the backend PeeringConnection proto message.
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

// CreatePeeringConnectionRequest matches the backend CreatePeeringConnectionRequest proto message.
type CreatePeeringConnectionRequest struct {
	SrcVpcID    string `json:"srcVpcId"`
	DestVpcID   string `json:"destVpcId"`
	Description string `json:"description,omitempty"`
}

// UpdatePeeringConnectionRequest matches the backend UpdatePeeringConnectionRequest proto message.
type UpdatePeeringConnectionRequest struct {
	Name string `json:"name"`
}

// PeeringConnectionResponse matches the backend PeeringConnectionResponse proto message.
type PeeringConnectionResponse struct {
	PeeringConnection PeeringConnection `json:"peeringConnection"`
}

// ListPeeringConnectionsResponse matches the backend ListPeeringConnectionsResponse proto message.
type ListPeeringConnectionsResponse struct {
	PeeringConnections []PeeringConnection `json:"peeringConnections"`
}
