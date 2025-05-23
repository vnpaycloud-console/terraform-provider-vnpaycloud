package routetable

type RouteTable struct {
	ID            string `json:"id"`
	Name          string `json:"name"`
	DestCidr      string `json:"dest_cidr"`
	Status        string `json:"status"`
	TargetId      string `json:"target_id"`
	VpcId         string `json:"vpc_id"`
	RouterRouteId string `json:"router_route_id"`
	RouteStatus   string `json:"router_status"`
}

type CreateRouteTableOpts struct {
	VpcId    string `json:"src_vpc_id"`
	Cidr     string `json:"dest_cidr"`
	TargetId string `json:"target_id"`
}

type GetRouteTableResponse struct {
	RouteTable RouteTable `json:"route_table"`
}

type CreateRouteTableRequest struct {
	RouteTable CreateRouteTableOpts `json:"route_table"`
}

type CreateRouteTableResponse struct {
	RouteTable RouteTable `json:"route_table"`
}
