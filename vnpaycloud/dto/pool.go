package dto

type Pool struct {
	LBMethod           string             `json:"lb_algorithm"`
	Protocol           string             `json:"protocol"`
	Description        string             `json:"description"`
	Listeners          []ListenerID       `json:"listeners"` //[]map[string]any
	Members            []Member           `json:"members"`
	MonitorID          string             `json:"healthmonitor_id"`
	SubnetID           string             `json:"subnet_id"`
	ProjectID          string             `json:"project_id"`
	AdminStateUp       bool               `json:"admin_state_up"`
	Name               string             `json:"name"`
	ID                 string             `json:"id"`
	Loadbalancers      []LoadBalancerID   `json:"loadbalancers"`
	Persistence        SessionPersistence `json:"session_persistence"`
	ALPNProtocols      []string           `json:"alpn_protocols"`
	CATLSContainerRef  string             `json:"ca_tls_container_ref"`
	CRLContainerRef    string             `json:"crl_container_ref"`
	TLSEnabled         bool               `json:"tls_enabled"`
	TLSCiphers         string             `json:"tls_ciphers"`
	TLSContainerRef    string             `json:"tls_container_ref"`
	TLSVersions        []string           `json:"tls_versions"`
	Provider           string             `json:"provider"`
	Monitor            Monitor            `json:"healthmonitor"`
	ProvisioningStatus string             `json:"provisioning_status"`
	OperatingStatus    string             `json:"operating_status"`
	Tags               []string           `json:"tags"`
}

type SessionPersistence struct {
	Type       string `json:"type"`                  // "SOURCE_IP", "HTTP_COOKIE", "APP_COOKIE"
	CookieName string `json:"cookie_name,omitempty"` // APP_COOKIE only
}

type GetPoolResponse struct {
	Pool Pool `json:"pool"`
}

type CreatePoolRequest struct {
	Pool CreatePoolOpts `json:"pool"`
}

type CreatePoolOpts struct {
	Name           string `json:"name,omitempty"`
	Description    string `json:"description,omitempty"`
	Protocol       string `json:"protocol"`        // "HTTP", "HTTPS", "TCP", "UDP"
	LBMethod       string `json:"lb_algorithm"`    // "ROUND_ROBIN", "LEAST_CONNECTIONS", ...
	LoadbalancerID string `json:"loadbalancer_id"` // ID LB
	ListenerID     string `json:"listener_id,omitempty"`
	//SessionPersistence *SessionPersistence `json:"session_persistence,omitempty"`
	//Tags               []string            `json:"tags,omitempty"`
}

type CreatePoolResponse struct {
	Pool Pool `json:"pool"`
}

type PoolID struct {
	ID string `json:"id"`
}
