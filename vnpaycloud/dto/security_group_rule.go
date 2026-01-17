package dto

type SecGroupRule struct {
	// The UUID for this security group rule.
	ID string

	// The direction in which the security group rule is applied. The only values
	// allowed are "ingress" or "egress". For a compute instance, an ingress
	// security group rule is applied to incoming (ingress) traffic for that
	// instance. An egress rule is applied to traffic leaving the instance.
	Direction string

	// Description of the rule
	Description string `json:"description"`

	// Must be IPv4 or IPv6, and addresses represented in CIDR must match the
	// ingress or egress rules.
	EtherType string `json:"ethertype"`

	// The security group ID to associate with this security group rule.
	SecGroupID string `json:"security_group_id"`

	// The minimum port number in the range that is matched by the security group
	// rule. If the protocol is TCP or UDP, this value must be less than or equal
	// to the value of the PortRangeMax attribute. If the protocol is ICMP, this
	// value must be an ICMP type.
	PortRangeMin int `json:"port_range_min"`

	// The maximum port number in the range that is matched by the security group
	// rule. The PortRangeMin attribute constrains the PortRangeMax attribute. If
	// the protocol is ICMP, this value must be an ICMP type.
	PortRangeMax int `json:"port_range_max"`

	// The protocol that is matched by the security group rule. Valid values are
	// "tcp", "udp", "icmp" or an empty string.
	Protocol string

	// The remote group ID to be associated with this security group rule. You
	// can specify either RemoteGroupID or RemoteIPPrefix.
	RemoteGroupID string `json:"remote_group_id"`

	// The remote IP prefix to be associated with this security group rule. You
	// can specify either RemoteGroupID or RemoteIPPrefix . This attribute
	// matches the specified IP prefix as the source IP address of the IP packet.
	RemoteIPPrefix string `json:"remote_ip_prefix"`

	// TenantID is the project owner of this security group rule.
	TenantID string `json:"tenant_id"`

	// ProjectID is the project owner of this security group rule.
	ProjectID string `json:"project_id"`
}

type RuleDirection string
type RuleProtocol string
type RuleEtherType string

const (
	RuleDirIngress        RuleDirection = "ingress"
	RuleDirEgress         RuleDirection = "egress"
	RuleEtherType4        RuleEtherType = "IPv4"
	RuleEtherType6        RuleEtherType = "IPv6"
	RuleProtocolAH        RuleProtocol  = "ah"
	RuleProtocolDCCP      RuleProtocol  = "dccp"
	RuleProtocolEGP       RuleProtocol  = "egp"
	RuleProtocolESP       RuleProtocol  = "esp"
	RuleProtocolGRE       RuleProtocol  = "gre"
	RuleProtocolICMP      RuleProtocol  = "icmp"
	RuleProtocolIGMP      RuleProtocol  = "igmp"
	RuleProtocolIPIP      RuleProtocol  = "ipip"
	RuleProtocolIPv6Encap RuleProtocol  = "ipv6-encap"
	RuleProtocolIPv6Frag  RuleProtocol  = "ipv6-frag"
	RuleProtocolIPv6ICMP  RuleProtocol  = "ipv6-icmp"
	RuleProtocolIPv6NoNxt RuleProtocol  = "ipv6-nonxt"
	RuleProtocolIPv6Opts  RuleProtocol  = "ipv6-opts"
	RuleProtocolIPv6Route RuleProtocol  = "ipv6-route"
	RuleProtocolOSPF      RuleProtocol  = "ospf"
	RuleProtocolPGM       RuleProtocol  = "pgm"
	RuleProtocolRSVP      RuleProtocol  = "rsvp"
	RuleProtocolSCTP      RuleProtocol  = "sctp"
	RuleProtocolTCP       RuleProtocol  = "tcp"
	RuleProtocolUDP       RuleProtocol  = "udp"
	RuleProtocolUDPLite   RuleProtocol  = "udplite"
	RuleProtocolVRRP      RuleProtocol  = "vrrp"
	RuleProtocolAny       RuleProtocol  = ""
)

type CreateSecurityGroupRuleOpts struct {
	// Must be either "ingress" or "egress": the direction in which the security
	// group rule is applied.
	Direction RuleDirection `json:"direction" required:"true"`

	// String description of each rule, optional
	Description string `json:"description,omitempty"`

	// Must be "IPv4" or "IPv6", and addresses represented in CIDR must match the
	// ingress or egress rules.
	EtherType RuleEtherType `json:"ethertype" required:"true"`

	// The security group ID to associate with this security group rule.
	SecGroupID string `json:"security_group_id" required:"true"`

	// The maximum port number in the range that is matched by the security group
	// rule. The PortRangeMin attribute constrains the PortRangeMax attribute. If
	// the protocol is ICMP, this value must be an ICMP type.
	PortRangeMax int `json:"port_range_max,omitempty"`

	// The minimum port number in the range that is matched by the security group
	// rule. If the protocol is TCP or UDP, this value must be less than or equal
	// to the value of the PortRangeMax attribute. If the protocol is ICMP, this
	// value must be an ICMP type.
	PortRangeMin int `json:"port_range_min,omitempty"`

	// The protocol that is matched by the security group rule. Valid values are
	// "tcp", "udp", "icmp" or an empty string.
	Protocol RuleProtocol `json:"protocol,omitempty"`

	// The remote group ID to be associated with this security group rule. You can
	// specify either RemoteGroupID or RemoteIPPrefix.
	RemoteGroupID string `json:"remote_group_id,omitempty"`

	// The remote IP prefix to be associated with this security group rule. You can
	// specify either RemoteGroupID or RemoteIPPrefix. This attribute matches the
	// specified IP prefix as the source IP address of the IP packet.
	RemoteIPPrefix string `json:"remote_ip_prefix,omitempty"`

	// TenantID is the UUID of the project who owns the Rule.
	// Only administrative users can specify a project UUID other than their own.
	ProjectID string `json:"project_id,omitempty"`
}

type CreateSecurityGroupRuleRequest struct {
	SecurityGroupRule CreateSecurityGroupRuleOpts `json:"security_group_rule,omitempty"`
}

type CreateSecurityGroupRuleResponse struct {
	SecurityGroupRule SecGroupRule `json:"security_group_rule"`
}

type GetSecurityGroupRuleResponse struct {
	SecurityGroupRule SecGroupRule `json:"security_group_rule"`
}
