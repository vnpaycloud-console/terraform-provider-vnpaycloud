package robotaccount

import (
	"context"
	"terraform-provider-vnpaycloud/vnpaycloud/dto"
	"terraform-provider-vnpaycloud/vnpaycloud/helper/client"

	"github.com/hashicorp/terraform-plugin-log/tflog"
)

// robotAccountExists checks if a robot account still exists.
// Returns "active" if found, "deleted" if 404, or error.
func robotAccountStateRefreshFunc(ctx context.Context, c *client.Client, projectID, registryID, id string) func() (interface{}, string, error) {
	return func() (interface{}, string, error) {
		resp := &dto.RobotAccountResponse{}
		httpResp, err := c.Get(ctx, client.ApiPath.RobotAccountWithID(projectID, registryID, id), resp, nil)
		if err != nil {
			if httpResp != nil && httpResp.StatusCode == 404 {
				return resp, "deleted", nil
			}
			return nil, "", err
		}

		// Robot accounts don't have a status field that transitions.
		// If we can read it, it's active.
		status := "active"
		if !resp.RobotAccount.Enabled {
			status = "disabled"
		}

		tflog.Trace(ctx, "robotAccountStateRefreshFunc", map[string]interface{}{
			"id":     id,
			"status": status,
		})

		return resp, status, nil
	}
}
