package networkinterface

import (
	"context"
	"fmt"
	"net/http"
	"terraform-provider-vnpaycloud/vnpaycloud/dto"
	"terraform-provider-vnpaycloud/vnpaycloud/helper/client"
	"terraform-provider-vnpaycloud/vnpaycloud/util"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/retry"
)

func networkInterfaceStateRefreshFunc(ctx context.Context, c *client.Client, projectID, networkInterfaceID string) retry.StateRefreshFunc {
	return func() (interface{}, string, error) {
		niResp := &dto.NetworkInterfaceResponse{}
		_, err := c.Get(ctx, client.ApiPath.NetworkInterfaceWithID(projectID, networkInterfaceID), niResp, nil)

		if err != nil {
			if util.ResponseCodeIs(err, http.StatusNotFound) {
				return niResp.NetworkInterface, "deleted", nil
			}
			return nil, "", err
		}

		if niResp.NetworkInterface.Status == "failed" {
			return niResp.NetworkInterface, niResp.NetworkInterface.Status, fmt.Errorf("The network interface is in error status. " +
				"Please check with your cloud admin or check the API logs.")
		}

		return niResp.NetworkInterface, niResp.NetworkInterface.Status, nil
	}
}
