package registryproject

import (
	"context"
	"terraform-provider-vnpaycloud/vnpaycloud/dto"
	"terraform-provider-vnpaycloud/vnpaycloud/helper/client"

	"github.com/hashicorp/terraform-plugin-log/tflog"
)

// registryProjectStateRefreshFunc returns the current status of a registry project.
// If the resource returns 404, it is treated as "deleted".
func registryProjectStateRefreshFunc(ctx context.Context, c *client.Client, projectID, id string) func() (interface{}, string, error) {
	return func() (interface{}, string, error) {
		resp := &dto.RegistryProjectResponse{}
		httpResp, err := c.Get(ctx, client.ApiPath.RegistryProjectWithID(projectID, id), resp, nil)
		if err != nil {
			if httpResp != nil && httpResp.StatusCode == 404 {
				return resp, "deleted", nil
			}
			return nil, "", err
		}

		status := resp.Registry.Status
		if status == "" {
			status = "unknown"
		}

		tflog.Trace(ctx, "registryProjectStateRefreshFunc", map[string]interface{}{
			"id":     id,
			"status": status,
		})

		return resp, status, nil
	}
}
