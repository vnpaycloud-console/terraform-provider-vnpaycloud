package routetable

import (
	"context"
	"fmt"
	"net/http"
	"terraform-provider-vnpaycloud/vnpaycloud/helper/client"
	"terraform-provider-vnpaycloud/vnpaycloud/util"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/retry"
)

func routeTableStateRefreshFunc(ctx context.Context, consoleClient *client.Client, id string) retry.StateRefreshFunc {
	return func() (interface{}, string, error) {
		getResp := &GetRouteTableResponse{}
		_, err := consoleClient.Get(ctx, client.ApiPath.RouteTableWithId(id), getResp, nil)

		if err != nil {
			if util.ResponseCodeIs(err, http.StatusNotFound) {
				return getResp.RouteTable, "OS_DELETED", nil
			}

			return nil, "", err
		}

		if getResp.RouteTable.Status == "OS_FAILED" {
			return getResp.RouteTable, getResp.RouteTable.Status, fmt.Errorf("The Route Table is in error status. " +
				"Please check with your cloud admin or check the Route Table " +
				"API logs to see why this error occurred.")
		}

		return getResp.RouteTable, getResp.RouteTable.Status, nil
	}
}
