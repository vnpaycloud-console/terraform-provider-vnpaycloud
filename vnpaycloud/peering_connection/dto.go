package peeringconnection

// PeeringConnectionRequest
type PeeringConnectionRequest struct {
	ID             string `json:"id"`
	RequestStatus  string `json:"request_status"`
	PeerId         string `json:"peering_connection_id"`
	Description    string `json:"description"`
	ConnectionType string `json:"connection_type"`
	Status         string `json:"status"`
	VpcId          string `json:"src_vpc_id"`
	PeerOrgId      string `json:"dest_org_id"`
	PeerVpcId      string `json:"dest_vpc_id"`
}

type GetPeeringConnectionRequestResponse struct {
	PeeringConnectionRequest PeeringConnectionRequest `json:"peering_connection_request"`
}

type CreatePeeringConnectionRequestOpts struct {
	PeerVPCId   string `json:"dest_vpc_id,omitempty"`
	PeerOrgId   string `json:"dest_org_id,omitempty"`
	VPCId       string `json:"src_vpc_id,omitempty"`
	Description string `json:"description,omitempty"`
}

type CreatePeeringConnectionRequest struct {
	PeeringConnectionRequest CreatePeeringConnectionRequestOpts `json:"peering_connection_request,omitempty"`
}

type CreatePeeringConnectionRequestResponse struct {
	PeeringConnectionRequest PeeringConnectionRequest `json:"peering_connection_request"`
}

// PeeringConnectionApproval
type PeeringConnectApproval struct {
	ID          string `json:"id"`
	PeerId      string `json:"peering_connection_id"`
	Name        string `json:"name"`
	Description string `json:"description"`
	VpcId       string `json:"src_vpc_id"`
	PeerOrgId   string `json:"dest_org_id"`
	PeerVpcId   string `json:"dest_vpc_id"`
	Status      string `json:"status"`
}

type ListPeeringConnectApprovalRequest struct {
	PeerVPCId string `q:"src_vpc_id,omitempty"`
	PeerOrgId string `q:"src_org_id,omitempty"`
	VPCId     string `q:"dest_vpc_id,omitempty"`
	Status    string `q:"status,omitempty"`
}

type ListPeeringConnectApprovalResponse struct {
	PeeringConnectApprovals []*PeeringConnectApproval `json:"peering_connection_approvals,omitempty"`
}

type PeeringConnectApprovalOpts struct {
	Accept bool `json:"is_allowed,omitempty"`
}

type UpdatePeeringConnectApprovalRequest struct {
	PeeringConnectApproval PeeringConnectApprovalOpts `json:"peering_connection_approval,omitempty"`
}

type UpdatePeeringConnectApprovalResponse struct {
	PeeringConnectApproval PeeringConnectApproval `json:"peering_connection_approval,omitempty"`
}

// PeeringConnection
type PeeringConnection struct {
	ID             string `json:"id"`
	PeerStatus     string `json:"peering_status"`
	Description    string `json:"description"`
	ConnectionType string `json:"connection_type"`
	Status         string `json:"status"`
	VpcId          string `json:"src_vpc_id"`
	PeerOrgId      string `json:"dest_org_id"`
	PeerVpcId      string `json:"dest_vpc_id"`
	PortId         string `json:"port_peering_connection_id"`
}

type GetPeeringConnectionResponse struct {
	PeeringConnection PeeringConnection `json:"peering_connection"`
}
