package dto

// Listener matches the iac-proxy-v2 Listener proto message.
type Listener struct {
	ID                     string   `json:"id"`
	Name                   string   `json:"name"`
	Description            string   `json:"description"`
	LoadBalancerID         string   `json:"loadBalancerId"`
	Protocol               string   `json:"protocol"`
	ProtocolPort           int      `json:"protocolPort"`
	DefaultPoolID          string   `json:"defaultPoolId"`
	InsertHeaders          []string `json:"insertHeaders,omitempty"`
	AllowedCidrs           []string `json:"allowedCidrs,omitempty"`
	ConnectionLimit        int      `json:"connectionLimit"`
	TimeoutClientData      int      `json:"timeoutClientData"`
	TimeoutMemberConnect   int      `json:"timeoutMemberConnect"`
	TimeoutMemberData      int      `json:"timeoutMemberData"`
	CertificateID          string   `json:"certificateId"`
	CertificateAuthorityID string   `json:"certificateAuthorityId"`
	SniCertificateIDs      []string `json:"sniCertificateIds,omitempty"`
	Status                 string   `json:"status"`
	ProvisioningStatus     string   `json:"provisioningStatus"`
	OperatingStatus        string   `json:"operatingStatus"`
	CreatedAt              string   `json:"createdAt"`
}

// CreateListenerRequest matches the iac-proxy-v2 CreateListenerRequest proto message.
// project_id is passed via URL path.
type CreateListenerRequest struct {
	Name                   string   `json:"name"`
	Description            string   `json:"description,omitempty"`
	LoadBalancerID         string   `json:"loadBalancerId"`
	Protocol               string   `json:"protocol"`
	ProtocolPort           int      `json:"protocolPort"`
	DefaultPoolID          string   `json:"defaultPoolId,omitempty"`
	InsertHeaders          []string `json:"insertHeaders,omitempty"`
	AllowedCidrs           []string `json:"allowedCidrs,omitempty"`
	ConnectionLimit        int      `json:"connectionLimit,omitempty"`
	TimeoutClientData      int      `json:"timeoutClientData,omitempty"`
	TimeoutMemberConnect   int      `json:"timeoutMemberConnect,omitempty"`
	TimeoutMemberData      int      `json:"timeoutMemberData,omitempty"`
	CertificateID          string   `json:"certificateId,omitempty"`
	CertificateAuthorityID string   `json:"certificateAuthorityId,omitempty"`
	SniCertificateIDs      []string `json:"sniCertificateIds,omitempty"`
}

// UpdateListenerRequest matches the iac-proxy-v2 UpdateListenerRequest proto message.
// project_id and id are passed via URL path.
type UpdateListenerRequest struct {
	Name                   string   `json:"name,omitempty"`
	Description            string   `json:"description"`
	DefaultPoolID          string   `json:"defaultPoolId,omitempty"`
	InsertHeaders          []string `json:"insertHeaders,omitempty"`
	AllowedCidrs           []string `json:"allowedCidrs"`
	ConnectionLimit        int      `json:"connectionLimit"`
	TimeoutClientData      int      `json:"timeoutClientData"`
	TimeoutMemberConnect   int      `json:"timeoutMemberConnect"`
	TimeoutMemberData      int      `json:"timeoutMemberData"`
	CertificateID          string   `json:"certificateId,omitempty"`
	CertificateAuthorityID string   `json:"certificateAuthorityId,omitempty"`
	SniCertificateIDs      []string `json:"sniCertificateIds"`
}

// ListenerResponse matches the iac-proxy-v2 ListenerResponse proto message.
type ListenerResponse struct {
	Listener Listener `json:"listener"`
}

// ListListenersResponse matches the iac-proxy-v2 ListListenersResponse proto message.
type ListListenersResponse struct {
	Listeners []Listener `json:"listeners"`
}
