package listener

import (
	"context"
	"fmt"
	"net/http"
	"terraform-provider-vnpaycloud/vnpaycloud/dto"
	"terraform-provider-vnpaycloud/vnpaycloud/helper/client"
	"terraform-provider-vnpaycloud/vnpaycloud/util"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/retry"
)

func listenerStateRefreshFunc(ctx context.Context, c *client.Client, projectID, listenerID string) retry.StateRefreshFunc {
	return func() (interface{}, string, error) {
		resp := &dto.ListenerResponse{}
		_, err := c.Get(ctx, client.ApiPath.ListenerWithID(projectID, listenerID), resp, nil)

		if err != nil {
			if util.ResponseCodeIs(err, http.StatusNotFound) {
				return resp.Listener, "deleted", nil
			}
			return nil, "", err
		}

		if resp.Listener.Status == "failed" || resp.Listener.Status == "error" {
			return resp.Listener, resp.Listener.Status, fmt.Errorf("The listener is in error status. " +
				"Please check with your cloud admin or check the API logs.")
		}

		return resp.Listener, resp.Listener.Status, nil
	}
}
