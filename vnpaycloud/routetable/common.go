package routetable

import (
	"context"
	"net/http"
	"terraform-provider-vnpaycloud/vnpaycloud/dto"
	"terraform-provider-vnpaycloud/vnpaycloud/helper/client"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/retry"
)

func routeTableStateRefreshFunc(ctx context.Context, c *client.Client, projectID, routeTableID string) retry.StateRefreshFunc {
	return func() (interface{}, string, error) {
		resp := &dto.RouteTableResponse{}
		_, err := c.Get(ctx, client.ApiPath.RouteTableWithID(projectID, routeTableID), resp, nil)
		if err != nil {
			if client.ResponseCodeIs(err, http.StatusNotFound) {
				return &dto.RouteTable{}, "deleted", nil
			}
			return nil, "", err
		}

		if resp.RouteTable.ID == "" {
			return &dto.RouteTable{}, "deleted", nil
		}

		return &resp.RouteTable, resp.RouteTable.Status, nil
	}
}
