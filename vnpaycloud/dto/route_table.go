package dto

// RouteTable matches the iac-proxy-v2 RouteTable proto message.
type RouteTable struct {
	ID         string `json:"id"`
	VpcID      string `json:"vpcId"`
	DestCIDR   string `json:"destCidr"`
	TargetID   string `json:"targetId"`
	TargetType string `json:"targetType"`
	TargetName string `json:"targetName"`
	Name       string `json:"name"`
	Status     string `json:"status"`
	CreatedAt  string `json:"createdAt"`
	ProjectID  string `json:"projectId"`
	ZoneID     string `json:"zoneId"`
}

// CreateRouteTableRequest matches the iac-proxy-v2 CreateRouteTableRequest proto message.
// project_id is passed via URL path.
type CreateRouteTableRequest struct {
	VpcID      string `json:"vpcId"`
	DestCIDR   string `json:"destCidr"`
	TargetID   string `json:"targetId"`
	TargetType string `json:"targetType"`
}

// RouteTableResponse matches the iac-proxy-v2 RouteTableResponse proto message.
type RouteTableResponse struct {
	RouteTable RouteTable `json:"routeTable"`
}

// ListRouteTablesResponse matches the iac-proxy-v2 ListRouteTablesResponse proto message.
type ListRouteTablesResponse struct {
	RouteTables []RouteTable `json:"routeTables"`
}
