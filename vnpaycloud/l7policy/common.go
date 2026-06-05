package l7policy

import (
	"context"
	"fmt"
	"net/http"
	"terraform-provider-vnpaycloud/vnpaycloud/dto"
	"terraform-provider-vnpaycloud/vnpaycloud/helper/client"
	"terraform-provider-vnpaycloud/vnpaycloud/util"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/retry"
)

func l7PolicyStateRefreshFunc(ctx context.Context, c *client.Client, projectID, id string) retry.StateRefreshFunc {
	return func() (interface{}, string, error) {
		resp := &dto.L7PolicyResponse{}
		_, err := c.Get(ctx, client.ApiPath.L7PolicyWithID(projectID, id), resp, nil)

		if err != nil {
			if util.ResponseCodeIs(err, http.StatusNotFound) {
				return resp.L7Policy, "deleted", nil
			}
			return nil, "", err
		}

		prov := resp.L7Policy.ProvisioningStatus
		if prov == "" {
			if resp.L7Policy.Status == "failed" || resp.L7Policy.Status == "error" {
				return resp.L7Policy, resp.L7Policy.Status, fmt.Errorf("The L7 policy is in error status. " +
					"Please check with your cloud admin or check the API logs.")
			}
		} else if prov == "ERROR" {
			return resp.L7Policy, "error", fmt.Errorf("The L7 policy is in error status. " +
				"Please check with your cloud admin or check the API logs.")
		}

		return resp.L7Policy, resp.L7Policy.Status, nil
	}
}
