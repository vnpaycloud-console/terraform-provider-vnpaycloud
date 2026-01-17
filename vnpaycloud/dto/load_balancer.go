package dto

import (
	"time"
)

type LoadBalancer struct {
	Description        string          `json:"description"`
	AdminStateUp       bool            `json:"admin_state_up"`
	ProjectID          string          `json:"project_id"`
	UpdatedAt          time.Time       `json:"-"`
	CreatedAt          time.Time       `json:"-"`
	ProvisioningStatus string          `json:"provisioning_status"`
	VipAddress         string          `json:"vip_address"`
	VipPortID          string          `json:"vip_port_id"`
	VipSubnetID        string          `json:"vip_subnet_id"`
	VipNetworkID       string          `json:"vip_network_id"`
	VipQosPolicyID     string          `json:"vip_qos_policy_id"`
	ID                 string          `json:"id"`
	OperatingStatus    string          `json:"operating_status"`
	Name               string          `json:"name"`
	FlavorID           string          `json:"flavor_id"`
	AvailabilityZone   string          `json:"availability_zone"`
	Provider           string          `json:"provider"`
	Listeners          []Listener      `json:"listeners"`
	Pools              []Pool          `json:"pools"`
	Tags               []string        `json:"tags"`
	AdditionalVips     []AdditionalVip `json:"additional_vips"`
}

type AdditionalVip struct {
	SubnetID  string `json:"subnet_id"`
	IPAddress string `json:"ip_address,omitempty"`
}

type GetLoadBalancerResponse struct {
	LoadBalancer LoadBalancer `json:"loadbalancer"`
}

type CreateLoadBalancerRequest struct {
	LoadBalancer CreateLoadBalancerOpts `json:"loadbalancer,omitempty"`
}

type CreateLoadBalancerOpts struct {
	Name        string `json:"name,omitempty"`
	Description string `json:"description,omitempty"`
	FlavorID    string `json:"flavor_id,omitempty"`
	VipSubnetID string `json:"vip_subnet_id,omitempty"`
}

type CreateLoadBalancerResponse struct {
	LoadBalancer LoadBalancer `json:"loadbalancer"`
}

type LoadBalancerID struct {
	ID string `json:"id"`
}

type ListLoadBalancerParams struct {
	Description        string   `q:"description"`
	AdminStateUp       *bool    `q:"admin_state_up"`
	ProjectID          string   `q:"project_id"`
	ProvisioningStatus string   `q:"provisioning_status"`
	VipAddress         string   `q:"vip_address"`
	VipPortID          string   `q:"vip_port_id"`
	VipSubnetID        string   `q:"vip_subnet_id"`
	VipNetworkID       string   `q:"vip_network_id"`
	ID                 string   `q:"id"`
	OperatingStatus    string   `q:"operating_status"`
	Name               string   `q:"name"`
	FlavorID           string   `q:"flavor_id"`
	AvailabilityZone   string   `q:"availability_zone"`
	Provider           string   `q:"provider"`
	Limit              int      `q:"limit"`
	Marker             string   `q:"marker"`
	SortKey            string   `q:"sort_key"`
	SortDir            string   `q:"sort_dir"`
	Tags               []string `q:"tags"`
	TagsAny            []string `q:"tags-any"`
	TagsNot            []string `q:"not-tags"`
	TagsNotAny         []string `q:"not-tags-any"`
}

type ListLoadBalancerResponse struct {
	LoadBalancers []LoadBalancer `json:"loadbalancers"`
}

type LoadBalancerStatusTree struct {
	Loadbalancer *LoadBalancer `json:"loadbalancer"`
}

type GetLoadBalancerStatusResponse struct {
	Statuses *LoadBalancerStatusTree `json:"statuses"`
}

type L7Policy struct {
	ID                 string   `json:"id"`
	Name               string   `json:"name"`
	ListenerID         string   `json:"listener_id"`
	Action             string   `json:"action"`
	Position           int32    `json:"position"`
	Description        string   `json:"description"`
	ProjectID          string   `json:"project_id"`
	RedirectPoolID     string   `json:"redirect_pool_id"`
	RedirectPrefix     string   `json:"redirect_prefix"`
	RedirectURL        string   `json:"redirect_url"`
	RedirectHttpCode   int32    `json:"redirect_http_code"`
	AdminStateUp       bool     `json:"admin_state_up"`
	ProvisioningStatus string   `json:"provisioning_status"`
	OperatingStatus    string   `json:"operating_status"`
	Rules              []Rule   `json:"rules"`
	Tags               []string `json:"tags"`
}

type Rule struct {
	ID                 string   `json:"id"`
	RuleType           string   `json:"type"`
	CompareType        string   `json:"compare_type"`
	Value              string   `json:"value"`
	ProjectID          string   `json:"project_id"`
	Key                string   `json:"key"`
	Invert             bool     `json:"invert"`
	AdminStateUp       bool     `json:"admin_state_up"`
	ProvisioningStatus string   `json:"provisioning_status"`
	OperatingStatus    string   `json:"operating_status"`
	Tags               []string `json:"tags"`
}

type GetRuleResponse struct {
	Rule Rule `json:"rule"`
}

type GetL7PolicyResponse struct {
	L7Policy L7Policy `json:"l7policies"`
}

type Action string
type RuleType string
type CompareType string

const (
	ActionRedirectPrefix Action = "REDIRECT_PREFIX"
	ActionRedirectToPool Action = "REDIRECT_TO_POOL"
	ActionRedirectToURL  Action = "REDIRECT_TO_URL"
	ActionReject         Action = "REJECT"

	TypeCookie          RuleType = "COOKIE"
	TypeFileType        RuleType = "FILE_TYPE"
	TypeHeader          RuleType = "HEADER"
	TypeHostName        RuleType = "HOST_NAME"
	TypePath            RuleType = "PATH"
	TypeSSLConnHasCert  RuleType = "SSL_CONN_HAS_CERT"
	TypeSSLVerifyResult RuleType = "SSL_VERIFY_RESULT"
	TypeSSLDNField      RuleType = "SSL_DN_FIELD"

	CompareTypeContains  CompareType = "CONTAINS"
	CompareTypeEndWith   CompareType = "ENDS_WITH"
	CompareTypeEqual     CompareType = "EQUAL_TO"
	CompareTypeRegex     CompareType = "REGEX"
	CompareTypeStartWith CompareType = "STARTS_WITH"
)

type CreateL7PolicyOpts struct {
	Name             string           `json:"name,omitempty"`
	ListenerID       string           `json:"listener_id,omitempty"`
	Action           Action           `json:"action" required:"true"`
	Position         int32            `json:"position,omitempty"`
	Description      string           `json:"description,omitempty"`
	ProjectID        string           `json:"project_id,omitempty"`
	RedirectPrefix   string           `json:"redirect_prefix,omitempty"`
	RedirectPoolID   string           `json:"redirect_pool_id,omitempty"`
	RedirectURL      string           `json:"redirect_url,omitempty"`
	RedirectHttpCode int32            `json:"redirect_http_code,omitempty"`
	AdminStateUp     *bool            `json:"admin_state_up,omitempty"`
	Rules            []CreateRuleOpts `json:"rules,omitempty"`
	Tags             []string         `json:"tags,omitempty"`
}

type CreateRuleOpts struct {
	RuleType     RuleType    `json:"type" required:"true"`
	CompareType  CompareType `json:"compare_type" required:"true"`
	Value        string      `json:"value" required:"true"`
	ProjectID    string      `json:"project_id,omitempty"`
	Key          string      `json:"key,omitempty"`
	Invert       bool        `json:"invert,omitempty"`
	AdminStateUp *bool       `json:"admin_state_up,omitempty"`
	Tags         []string    `json:"tags,omitempty"`
}
