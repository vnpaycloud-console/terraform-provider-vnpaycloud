package client

import (
	"fmt"
)

var ApiPath = struct {
	Auth                                  string
	VPC                                   string
	VPCWithId                             func(id string) string
	VPCWithParams                         func(params any) string
	PeeringConnectionRequest              string
	PeeringConnectionRequestWithId        func(id string) string
	PeeringConnectionApproval             string
	PeeringConnectionApprovalWithId       func(id string) string
	ListPeeringConnectionApproval         func(params any) string
	PeeringConnection                     string
	PeeringConnectionWithId               func(id string) string
	RouteTable                            string
	RouteTableWithId                      func(id string) string
	Network                               string
	NetworkWithId                         func(id string) string
	NetworkWithParams                     func(params any) string
	Volume                                func(projectId string) string
	VolumeWithParams                      func(projectId string, params any) string
	VolumeWithId                          func(projectId string, id string) string
	VolumeAttachment                      func(projectId string) string
	VolumeAttachmentWithId                func(projectId string, id string) string
	VolumeAction                          func(projectId string, id string) string
	Subnet                                string
	SubnetWithId                          func(id string) string
	SubnetWithParams                      func(params any) string
	Port                                  string
	PortWithId                            func(id string) string
	PortWithParams                        func(params any) string
	LbaasLoadBalancer                     string
	LbaasLoadBalancerWithParams           func(params any) string
	LbaasLoadBalancerStatus               func(id string) string
	LbaasListener                         string
	LbaasPool                             string
	LbaasPoolMember                       func(id string) string
	LbaasHealthMonitor                    string
	LbaasLoadBalancerWithId               func(id string) string
	LbaasListenerWithId                   func(id string) string
	LbaasPoolWithId                       func(id string) string
	LbaasPoolMemberWithId                 func(poolId string, memberId string) string
	LbaasHealthMonitorWithId              func(id string) string
	LbaasL7Policy                         func(id string) string
	LbaasL7PolicyWithId                   func(id string) string
	LbaasL7Rule                           func(policyId string) string
	LbaasL7RuleWithId                     func(policyId string, id string) string
	LbaasL7RuleWithParams                 func(policyId string, params any) string
	FloatingIP                            string
	FloatingIPWithId                      func(id string) string
	FloatingIPWithParams                  func(params any) string
	KeyPair                               string
	KeyPairWithParams                     func(params any) string
	KeyPairWithName                       func(name string) string
	KeyPairWithIdAndParams                func(id string, params any) string
	Flavor                                string
	FlavorDetail                          string
	FlavorDetailWithParams                func(params any) string
	FlavorWithId                          func(id string) string
	FlavorWithParams                      func(params any) string
	FlavorExtraSpecs                      func(id string) string
	ServerGroup                           string
	ServerGroupWithId                     func(id string) string
	SecurityGroup                         string
	SecurityGroupWithId                   func(id string) string
	SecurityGroupWithParams               func(params any) string
	SecurityGroupRule                     string
	SecurityGroupRuleWithId               func(id string) string
	ServerInterfaceAttach                 func(serverId string) string
	ServerInterfaceAttachWithId           func(serverId string, portId string) string
	ApplicationCredential                 func(userId string) string
	ApplicationCredentialWithId           func(userId string, id string) string
	ApplicationCredentialAccessRule       func(userId string) string
	ApplicationCredentialAccessRuleWithId func(userId string, id string) string
	Image                                 string
	ImageWithParams                       func(params any) string
	ImageWithId                           func(id string) string
	Server                                string
	ServerWithId                          func(id string) string
	ServerActionWithId                    func(id string) string
	ServerMetadataWithId                  func(id string) string
}{
	Auth: "/v3/auth/tokens",
	VPC:  "/v2.0/vpcs",
	VPCWithId: func(id string) string {
		return fmt.Sprintf("/v2.0/vpcs/%s", id)
	},
	VPCWithParams: func(params any) string {
		query, _ := BuildQueryString(params)
		return "/v2.0/vpcs?" + query.Query().Encode()
	},
	PeeringConnectionRequest: "/v2.0/peering-connection-requests",
	PeeringConnectionRequestWithId: func(id string) string {
		return fmt.Sprintf("/v2.0/peering-connection-requests/%s", id)
	},
	PeeringConnectionApproval: "/v2.0/peering-connection-approvals",
	PeeringConnectionApprovalWithId: func(id string) string {
		return fmt.Sprintf("/v2.0/peering-connection-approvals/%s", id)
	},
	ListPeeringConnectionApproval: func(params any) string {
		query, _ := BuildQueryString(params)
		return "/v2.0/peering-connection-approvals?" + query.Query().Encode()
	},
	PeeringConnection: "/v2.0/peering-connections",
	PeeringConnectionWithId: func(id string) string {
		return fmt.Sprintf("/v2.0/peering-connections/%s", id)
	},
	RouteTable: "/v2.0/route-tables",
	RouteTableWithId: func(id string) string {
		return fmt.Sprintf("/v2.0/route-tables/%s", id)
	},
	Network: "/v2.0/networks",
	NetworkWithId: func(id string) string {
		return fmt.Sprintf("/v2.0/networks/%s", id)
	},
	NetworkWithParams: func(params any) string {
		query, _ := BuildQueryString(params)
		return "/v2.0/networks?" + query.Query().Encode()
	},
	Subnet: "/v2.0/subnets",
	SubnetWithId: func(id string) string {
		return fmt.Sprintf("/v2.0/subnets/%s", id)
	},
	SubnetWithParams: func(params any) string {
		query, _ := BuildQueryString(params)
		return "/v2.0/subnets?" + query.Query().Encode()
	},
	Port: "/v2.0/ports",
	PortWithId: func(id string) string {
		return fmt.Sprintf("/v2.0/ports/%s", id)
	},
	PortWithParams: func(params any) string {
		query, _ := BuildQueryString(params)
		return "/v2.0/ports?" + query.Query().Encode()
	},
	LbaasLoadBalancer: "/v2.0/lbaas/loadbalancers",
	LbaasLoadBalancerWithId: func(id string) string {
		return fmt.Sprintf("/v2.0/lbaas/loadbalancers/%s", id)
	},
	LbaasLoadBalancerWithParams: func(params any) string {
		query, _ := BuildQueryString(params)
		return "/v2.0/lbaas/loadbalancers?" + query.Query().Encode()
	},
	LbaasLoadBalancerStatus: func(id string) string {
		return fmt.Sprintf("/v2.0/lbaas/loadbalancers/%s/status", id)
	},
	LbaasListener: "/v2.0/lbaas/listeners",
	LbaasListenerWithId: func(id string) string {
		return fmt.Sprintf("/v2.0/lbaas/listeners/%s", id)
	},
	LbaasPool: "/v2.0/lbaas/pools",
	LbaasPoolWithId: func(id string) string {
		return fmt.Sprintf("/v2.0/lbaas/pools/%s", id)
	},
	LbaasPoolMember: func(poolId string) string {
		return fmt.Sprintf("/v2.0/lbaas/pools/%s/members", poolId)
	},
	LbaasPoolMemberWithId: func(poolId string, memberId string) string {
		return fmt.Sprintf("/v2.0/lbaas/pools/%s/members/%s", poolId, memberId)
	},
	LbaasHealthMonitor: "/v2.0/lbaas/healthmonitors",
	LbaasHealthMonitorWithId: func(id string) string {
		return fmt.Sprintf("/v2.0/lbaas/healthmonitors/%s", id)
	},
	LbaasL7Policy: func(policyId string) string {
		return fmt.Sprintf("/v2.0/lbaas/l7policies/%s", policyId)
	},
	LbaasL7PolicyWithId: func(id string) string {
		return fmt.Sprintf("/v2.0/lbaas/l7policies/%s", id)
	},
	LbaasL7Rule: func(policyId string) string {
		return fmt.Sprintf("/v2.0/lbaas/l7policies/%s/rules", policyId)
	},
	LbaasL7RuleWithId: func(policyId string, id string) string {
		return fmt.Sprintf("/v2.0/lbaas/l7policies/%s/rules/%s", policyId, id)
	},
	LbaasL7RuleWithParams: func(policyId string, params any) string {
		query, _ := BuildQueryString(params)
		return fmt.Sprintf("/v2.0/lbaas/l7policies/%s/rules?%s", policyId, query.Query().Encode())
	},
	Volume: func(projectId string) string {
		return fmt.Sprintf("/v3/%s/volumes", projectId)
	},
	VolumeWithId: func(projectId string, id string) string {
		return fmt.Sprintf("/v3/%s/volumes/%s", projectId, id)
	},
	VolumeWithParams: func(projectId string, params any) string {
		query, _ := BuildQueryString(params)
		return fmt.Sprintf("/v3/%s/volumes/detail?%s", projectId, query.Query().Encode())
	},
	VolumeAttachment: func(projectId string) string {
		return fmt.Sprintf("/v3/%s/attachments", projectId)
	},
	VolumeAttachmentWithId: func(projectId string, id string) string {
		return fmt.Sprintf("/v3/%s/attachments/%s", projectId, id)
	},
	VolumeAction: func(projectId string, id string) string {
		return fmt.Sprintf("/v3/%s/volumes/%s/action", projectId, id)
	},
	FloatingIP: "/v2.0/floatingips",
	FloatingIPWithId: func(id string) string {
		return fmt.Sprintf("/v2.0/floatingips/%s", id)
	},
	FloatingIPWithParams: func(params any) string {
		query, _ := BuildQueryString(params)
		return "/v2.0/floatingips?" + query.Query().Encode()
	},
	KeyPair: "/v2.1/os-keypairs",
	KeyPairWithParams: func(params any) string {
		query, _ := BuildQueryString(params)
		return "/v2.1/os-keypairs?" + query.Query().Encode()
	},
	KeyPairWithName: func(name string) string {
		return fmt.Sprintf("/v2.1/os-keypairs/%s", name)
	},
	KeyPairWithIdAndParams: func(id string, params any) string {
		query, _ := BuildQueryString(params)
		return fmt.Sprintf("/v2.1/os-keypairs/%s?%s", id, query.Query().Encode())
	},
	Flavor:       "/v2.1/flavors",
	FlavorDetail: "/v2.1/flavors/detail",
	FlavorDetailWithParams: func(params any) string {
		query, _ := BuildQueryString(params)
		return "/v2.1/flavors/detail?" + query.Query().Encode()
	},
	FlavorWithId: func(id string) string {
		return fmt.Sprintf("/v2.1/flavors/%s", id)
	},
	FlavorWithParams: func(params any) string {
		query, _ := BuildQueryString(params)
		return "/v2.1/flavors?" + query.Query().Encode()
	},
	FlavorExtraSpecs: func(id string) string {
		return fmt.Sprintf("/v2.1/flavors/%s/os-extra_specs", id)
	},
	ServerGroup: "/v2.1/os-server-groups",
	ServerGroupWithId: func(id string) string {
		return fmt.Sprintf("/v2.1/os-server-groups/%s", id)
	},
	SecurityGroup: "/v2.0/security-groups",
	SecurityGroupWithId: func(id string) string {
		return fmt.Sprintf("/v2.0/security-groups/%s", id)
	},
	SecurityGroupWithParams: func(params any) string {
		query, _ := BuildQueryString(params)
		return "/v2.0/security-groups?" + query.Query().Encode()
	},
	SecurityGroupRule: "/v2.0/security-group-rules",
	SecurityGroupRuleWithId: func(id string) string {
		return fmt.Sprintf("/v2.0/security-group-rules/%s", id)
	},
	ServerInterfaceAttach: func(serverId string) string {
		return fmt.Sprintf("/servers/%s/os-interface", serverId)
	},
	ServerInterfaceAttachWithId: func(serverId string, portId string) string {
		return fmt.Sprintf("/servers/%s/os-interface/%s", serverId, portId)
	},
	ApplicationCredentialAccessRule: func(userId string) string {
		return fmt.Sprintf("/v3/%s/access_rules", userId)
	},
	ApplicationCredentialAccessRuleWithId: func(userId string, id string) string {
		return fmt.Sprintf("/v3/%s/access_rules/%s", userId, id)
	},
	ApplicationCredential: func(userId string) string {
		return fmt.Sprintf("/v3/%s/application_credentials", userId)
	},
	ApplicationCredentialWithId: func(userId string, id string) string {
		return fmt.Sprintf("/v3/%s/application_credentials/%s", userId, id)
	},
	Image: "/images",
	ImageWithParams: func(params any) string {
		query, _ := BuildQueryString(params)
		return "/images?" + query.Query().Encode()
	},
	ImageWithId: func(id string) string {
		return fmt.Sprintf("/images/%s", id)
	},
	Server: "/v2.1/servers",
	ServerWithId: func(id string) string {
		return fmt.Sprintf("/v2.1/servers/%s", id)
	},
	ServerActionWithId: func(id string) string {
		return fmt.Sprintf("/v2.1/servers/%s/action", id)
	},
	ServerMetadataWithId: func(id string) string {
		return fmt.Sprintf("/v2.1/servers/%s/metadata", id)
	},
}
