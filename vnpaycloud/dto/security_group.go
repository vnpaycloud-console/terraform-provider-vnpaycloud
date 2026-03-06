package dto

// SecurityGroupRule matches the iac-proxy-v2 SecurityGroupRule proto message.
type SecurityGroupRule struct {
	ID              string `json:"id"`
	SecurityGroupID string `json:"securityGroupId"`
	Direction       string `json:"direction"`
	Protocol        string `json:"protocol"`
	EtherType       string `json:"ethertype"`
	PortRangeMin    int32  `json:"portRangeMin"`
	PortRangeMax    int32  `json:"portRangeMax"`
	RemoteIPPrefix  string `json:"remoteIpPrefix"`
	RemoteGroupID   string `json:"remoteGroupId"`
	ProjectID       string `json:"projectId"`
	ZoneID          string `json:"zoneId"`
}

// SecurityGroup matches the iac-proxy-v2 SecurityGroup proto message.
type SecurityGroup struct {
	ID          string              `json:"id"`
	Name        string              `json:"name"`
	Description string              `json:"description"`
	Rules       []SecurityGroupRule `json:"rules"`
	CreatedAt   string              `json:"createdAt"`
	Status      string              `json:"status"`
	ProjectID   string              `json:"projectId"`
	ZoneID      string              `json:"zoneId"`
}

// CreateSecurityGroupRequest matches the iac-proxy-v2 CreateSecurityGroupRequest proto message.
// project_id is passed via URL path, not in the body.
type CreateSecurityGroupRequest struct {
	Name        string `json:"name"`
	Description string `json:"description,omitempty"`
}

// UpdateSecurityGroupRequest matches the iac-proxy-v2 UpdateSecurityGroupRequest proto message.
// project_id and id are passed via URL path.
type UpdateSecurityGroupRequest struct {
	Name        string `json:"name,omitempty"`
	Description string `json:"description,omitempty"`
}

// SecurityGroupResponse matches the iac-proxy-v2 SecurityGroupResponse proto message.
type SecurityGroupResponse struct {
	SecurityGroup SecurityGroup `json:"securityGroup"`
}

// ListSecurityGroupsResponse matches the iac-proxy-v2 ListSecurityGroupsResponse proto message.
type ListSecurityGroupsResponse struct {
	SecurityGroups []SecurityGroup `json:"securityGroups"`
}

// CreateSecurityGroupRuleRequest matches the iac-proxy-v2 CreateSecurityGroupRuleRequest proto message.
// project_id is passed via URL path, not in the body.
type CreateSecurityGroupRuleRequest struct {
	SecurityGroupID string `json:"securityGroupId"`
	Direction       string `json:"direction"`
	Protocol        string `json:"protocol,omitempty"`
	EtherType       string `json:"ethertype,omitempty"`
	PortRangeMin    int32  `json:"portRangeMin,omitempty"`
	PortRangeMax    int32  `json:"portRangeMax,omitempty"`
	RemoteIPPrefix  string `json:"remoteIpPrefix,omitempty"`
	RemoteGroupID   string `json:"remoteGroupId,omitempty"`
}

// SecurityGroupRuleResponse matches the iac-proxy-v2 SecurityGroupRuleResponse proto message.
type SecurityGroupRuleResponse struct {
	Rule SecurityGroupRule `json:"rule"`
}
