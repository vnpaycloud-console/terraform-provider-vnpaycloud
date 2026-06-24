package dto

// NetworkACL matches the backend NetworkACL proto message.
type NetworkACL struct {
	ID          string `json:"id"`
	Name        string `json:"name"`
	Description string `json:"description"`
	VpcID       string `json:"vpcId"`
	// The proxy wire field is networkIds, but Terraform exposes subnet_ids.
	SubnetIDs  []string `json:"networkIds"`
	TotalRules int64    `json:"totalRules,string"`
	Status     string   `json:"status"`
	CreatedAt  string   `json:"createdAt"`
	ProjectID  string   `json:"projectId"`
	ZoneID     string   `json:"zoneId"`
}

// CreateNetworkACLRequest — project_id is passed via URL path.
type CreateNetworkACLRequest struct {
	Name        string `json:"name"`
	VpcID       string `json:"vpcId"`
	Description string `json:"description,omitempty"`
}

// NetworkACLResponse wraps a single NetworkACL.
type NetworkACLResponse struct {
	NetworkACL NetworkACL `json:"networkAcl"`
}

// ListNetworkACLsResponse matches ListNetworkACLsResponse proto message.
type ListNetworkACLsResponse struct {
	NetworkACLs []NetworkACL `json:"networkAcls"`
}

// NetworkACLRule matches the backend NetworkACLRule proto message.
type NetworkACLRule struct {
	ID          string `json:"id"`
	NaclID      string `json:"naclId"`
	Name        string `json:"name"`
	Priority    int64  `json:"priority,string"`
	Type        string `json:"type"`
	Action      string `json:"action"`
	PortStart   int32  `json:"portStart"`
	PortEnd     int32  `json:"portEnd"`
	Source      string `json:"source"`
	Destination string `json:"destination"`
	IcmpType    string `json:"icmpType"`
	Description string `json:"description"`
	Status      string `json:"status"`
	ProjectID   string `json:"projectId"`
	ZoneID      string `json:"zoneId"`
}

// CreateNetworkACLRuleRequest — project_id is passed via URL path.
type CreateNetworkACLRuleRequest struct {
	NaclID      string `json:"naclId"`
	Name        string `json:"name"`
	Priority    int64  `json:"priority"`
	Type        string `json:"type"`
	Action      string `json:"action"`
	PortStart   int32  `json:"portStart,omitempty"`
	PortEnd     int32  `json:"portEnd,omitempty"`
	Source      string `json:"source,omitempty"`
	Destination string `json:"destination,omitempty"`
	IcmpType    string `json:"icmpType,omitempty"`
	Description string `json:"description,omitempty"`
}

// NetworkACLRuleResponse wraps a single NetworkACLRule.
type NetworkACLRuleResponse struct {
	Rule NetworkACLRule `json:"rule"`
}

// ListNetworkACLRulesResponse matches ListNetworkACLRulesResponse proto message.
type ListNetworkACLRulesResponse struct {
	NetworkACLRules []NetworkACLRule `json:"networkAclRules"`
}
