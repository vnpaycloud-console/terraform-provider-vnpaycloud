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

		if status := niResp.NetworkInterface.Status; status == "failed" || status == "error" {
			return niResp.NetworkInterface, status, fmt.Errorf("the network interface entered %q status; "+
				"please check with your cloud admin or the API logs", status)
		}

		return niResp.NetworkInterface, niResp.NetworkInterface.Status, nil
	}
}
