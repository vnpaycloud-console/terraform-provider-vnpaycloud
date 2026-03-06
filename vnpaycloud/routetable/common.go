package routetable

import (
	"context"
	"terraform-provider-vnpaycloud/vnpaycloud/dto"
	"terraform-provider-vnpaycloud/vnpaycloud/helper/client"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/retry"
)

// findRouteTableByID searches the route table list for a specific ID.
// Returns nil if not found (no GET single endpoint exists for route tables).
func findRouteTableByID(ctx context.Context, c *client.Client, projectID, routeTableID string) (*dto.RouteTable, error) {
	listResp := &dto.ListRouteTablesResponse{}
	_, err := c.Get(ctx, client.ApiPath.RouteTables(projectID), listResp, nil)
	if err != nil {
		return nil, err
	}

	for _, rt := range listResp.RouteTables {
		if rt.ID == routeTableID {
			return &rt, nil
		}
	}

	return nil, nil
}

func routeTableStateRefreshFunc(ctx context.Context, c *client.Client, projectID, routeTableID string) retry.StateRefreshFunc {
	return func() (interface{}, string, error) {
		rt, err := findRouteTableByID(ctx, c, projectID, routeTableID)
		if err != nil {
			return nil, "", err
		}

		if rt == nil {
			return &dto.RouteTable{}, "deleted", nil
		}

		return rt, rt.Status, nil
	}
}
