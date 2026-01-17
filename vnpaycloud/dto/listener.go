package dto

type Listener struct {
	ID                      string            `json:"id"`
	ProjectID               string            `json:"project_id"`
	Name                    string            `json:"name"`
	Description             string            `json:"description"`
	Protocol                string            `json:"protocol"`
	ProtocolPort            int               `json:"protocol_port"`
	DefaultPoolID           string            `json:"default_pool_id"`
	DefaultPool             *Pool             `json:"default_pool"`
	Loadbalancers           []LoadBalancerID  `json:"loadbalancers"`
	ConnLimit               int               `json:"connection_limit"`
	SniContainerRefs        []string          `json:"sni_container_refs"`
	DefaultTlsContainerRef  string            `json:"default_tls_container_ref"`
	AdminStateUp            bool              `json:"admin_state_up"`
	Pools                   []Pool            `json:"pools"`
	L7Policies              []L7Policy        `json:"l7policies"`
	ProvisioningStatus      string            `json:"provisioning_status"`
	TimeoutClientData       int               `json:"timeout_client_data"`
	TimeoutMemberData       int               `json:"timeout_member_data"`
	TimeoutMemberConnect    int               `json:"timeout_member_connect"`
	TimeoutTCPInspect       int               `json:"timeout_tcp_inspect"`
	InsertHeaders           map[string]string `json:"insert_headers"`
	AllowedCIDRs            []string          `json:"allowed_cidrs"`
	TLSCiphers              string            `json:"tls_ciphers"`
	TLSVersions             []string          `json:"tls_versions"`
	Tags                    []string          `json:"tags"`
	ALPNProtocols           []string          `json:"alpn_protocols"`
	ClientAuthentication    string            `json:"client_authentication"`
	ClientCATLSContainerRef string            `json:"client_ca_tls_container_ref"`
	ClientCRLContainerRef   string            `json:"client_crl_container_ref"`
	HSTSIncludeSubdomains   bool              `json:"hsts_include_subdomains"`
	HSTSMaxAge              int               `json:"hsts_max_age"`
	HSTSPreload             bool              `json:"hsts_preload"`
	OperatingStatus         string            `json:"operating_status"`
}

type GetListenerResponse struct {
	Listener Listener `json:"listener"`
}

type CreateListenerRequest struct {
	Listener CreateListenerOpts `json:"listener,omitempty"`
}

type CreateListenerOpts struct {
	LoadbalancerID          string               `json:"loadbalancer_id,omitempty"`
	Protocol                Protocol             `json:"protocol" required:"true"`
	ProtocolPort            int                  `json:"protocol_port" required:"true"`
	ProjectID               string               `json:"project_id,omitempty"`
	Name                    string               `json:"name,omitempty"`
	DefaultPoolID           string               `json:"default_pool_id,omitempty"`
	DefaultPool             *CreatePoolOpts      `json:"default_pool,omitempty"`
	Description             string               `json:"description,omitempty"`
	ConnLimit               *int                 `json:"connection_limit,omitempty"`
	DefaultTlsContainerRef  string               `json:"default_tls_container_ref,omitempty"`
	SniContainerRefs        []string             `json:"sni_container_refs,omitempty"`
	AdminStateUp            *bool                `json:"admin_state_up,omitempty"`
	L7Policies              []CreateL7PolicyOpts `json:"l7policies,omitempty"`
	TimeoutClientData       *int                 `json:"timeout_client_data,omitempty"`
	TimeoutMemberData       *int                 `json:"timeout_member_data,omitempty"`
	TimeoutMemberConnect    *int                 `json:"timeout_member_connect,omitempty"`
	TimeoutTCPInspect       *int                 `json:"timeout_tcp_inspect,omitempty"`
	InsertHeaders           map[string]string    `json:"insert_headers,omitempty"`
	AllowedCIDRs            []string             `json:"allowed_cidrs,omitempty"`
	ALPNProtocols           []string             `json:"alpn_protocols,omitempty"`
	ClientAuthentication    ClientAuthentication `json:"client_authentication,omitempty"`
	ClientCATLSContainerRef string               `json:"client_ca_tls_container_ref,omitempty"`
	ClientCRLContainerRef   string               `json:"client_crl_container_ref,omitempty"`
	HSTSIncludeSubdomains   bool                 `json:"hsts_include_subdomains,omitempty"`
	HSTSMaxAge              int                  `json:"hsts_max_age,omitempty"`
	HSTSPreload             bool                 `json:"hsts_preload,omitempty"`
	TLSCiphers              string               `json:"tls_ciphers,omitempty"`
	TLSVersions             []TLSVersion         `json:"tls_versions,omitempty"`
	Tags                    []string             `json:"tags,omitempty"`
}

type CreateListenerResponse struct {
	Listener Listener `json:"listener"`
}

type Protocol string

const (
	ProtocolTCP             Protocol = "TCP"
	ProtocolUDP             Protocol = "UDP"
	ProtocolPROXY           Protocol = "PROXY"
	ProtocolHTTP            Protocol = "HTTP"
	ProtocolHTTPS           Protocol = "HTTPS"
	ProtocolSCTP            Protocol = "SCTP"
	ProtocolPrometheus      Protocol = "PROMETHEUS"
	ProtocolTerminatedHTTPS Protocol = "TERMINATED_HTTPS"
)

type ClientAuthentication string

const (
	ClientAuthenticationNone      ClientAuthentication = "NONE"
	ClientAuthenticationOptional  ClientAuthentication = "OPTIONAL"
	ClientAuthenticationMandatory ClientAuthentication = "MANDATORY"
)

type TLSVersion string

const (
	TLSVersionSSLv3   TLSVersion = "SSLv3"
	TLSVersionTLSv1   TLSVersion = "TLSv1"
	TLSVersionTLSv1_1 TLSVersion = "TLSv1.1"
	TLSVersionTLSv1_2 TLSVersion = "TLSv1.2"
	TLSVersionTLSv1_3 TLSVersion = "TLSv1.3"
)

type ListenerID struct {
	ID string `json:"id"`
}
