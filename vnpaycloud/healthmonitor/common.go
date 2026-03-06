package healthmonitor

import (
	"context"
	"fmt"
	"net/http"
	"terraform-provider-vnpaycloud/vnpaycloud/dto"
	"terraform-provider-vnpaycloud/vnpaycloud/helper/client"
	"terraform-provider-vnpaycloud/vnpaycloud/util"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/retry"
)

func healthMonitorStateRefreshFunc(ctx context.Context, c *client.Client, projectID, hmID string) retry.StateRefreshFunc {
	return func() (interface{}, string, error) {
		resp := &dto.HealthMonitorResponse{}
		_, err := c.Get(ctx, client.ApiPath.HealthMonitorWithID(projectID, hmID), resp, nil)

		if err != nil {
			if util.ResponseCodeIs(err, http.StatusNotFound) {
				return resp.HealthMonitor, "deleted", nil
			}
			return nil, "", err
		}

		if resp.HealthMonitor.Status == "failed" || resp.HealthMonitor.Status == "error" {
			return resp.HealthMonitor, resp.HealthMonitor.Status, fmt.Errorf("The health monitor is in error status. " +
				"Please check with your cloud admin or check the API logs.")
		}

		return resp.HealthMonitor, resp.HealthMonitor.Status, nil
	}
}
