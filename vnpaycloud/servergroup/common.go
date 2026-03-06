package servergroup

import (
	"context"
	"net/http"
	"terraform-provider-vnpaycloud/vnpaycloud/dto"
	"terraform-provider-vnpaycloud/vnpaycloud/helper/client"
	"terraform-provider-vnpaycloud/vnpaycloud/util"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/retry"
)

func serverGroupStateRefreshFunc(ctx context.Context, c *client.Client, projectID, serverGroupID string) retry.StateRefreshFunc {
	return func() (interface{}, string, error) {
		sgResp := &dto.ServerGroupResponse{}
		_, err := c.Get(ctx, client.ApiPath.ServerGroupWithID(projectID, serverGroupID), sgResp, nil)

		if err != nil {
			if util.ResponseCodeIs(err, http.StatusNotFound) {
				return sgResp.ServerGroup, "deleted", nil
			}

			return nil, "", err
		}

		if sgResp.ServerGroup.ID == "" {
			return sgResp.ServerGroup, "deleted", nil
		}

		// Server groups are typically immediately active after creation
		return sgResp.ServerGroup, "active", nil
	}
}
