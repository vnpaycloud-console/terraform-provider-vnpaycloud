package client

import (
	"fmt"
)

var ApiPath = struct {
	Auth                            string
	VPC                             string
	VPCWithId                       func(id string) string
	PeeringConnectionRequest        string
	PeeringConnectionRequestWithId  func(id string) string
	PeeringConnectionApproval       string
	PeeringConnectionApprovalWithId func(id string) string
	ListPeeringConnectionApproval   func(params any) string
	PeeringConnection               string
	PeeringConnectionWithId         func(id string) string
	RouteTable                      string
	RouteTableWithId                func(id string) string
}{
	Auth: "/v3/auth/tokens",
	VPC:  "/v2.0/vpcs",
	VPCWithId: func(id string) string {
		return fmt.Sprintf("/v2.0/vpcs/%s", id)
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
}
