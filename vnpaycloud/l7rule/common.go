package l7rule

import (
	"context"
	"fmt"
	"net/http"
	"terraform-provider-vnpaycloud/vnpaycloud/dto"
	"terraform-provider-vnpaycloud/vnpaycloud/helper/client"
	"terraform-provider-vnpaycloud/vnpaycloud/util"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/retry"
)

func l7RuleStateRefreshFunc(ctx context.Context, c *client.Client, projectID, l7policyID, id string) retry.StateRefreshFunc {
	return func() (interface{}, string, error) {
		resp := &dto.L7RuleResponse{}
		_, err := c.Get(ctx, client.ApiPath.L7RuleWithID(projectID, l7policyID, id), resp, nil)

		if err != nil {
			if util.ResponseCodeIs(err, http.StatusNotFound) {
				return resp.L7Rule, "deleted", nil
			}
			return nil, "", err
		}

		prov := resp.L7Rule.ProvisioningStatus
		if prov == "" {
			if resp.L7Rule.Status == "failed" || resp.L7Rule.Status == "error" {
				return resp.L7Rule, resp.L7Rule.Status, fmt.Errorf("The L7 rule is in error status. " +
					"Please check with your cloud admin or check the API logs.")
			}
		} else if prov == "ERROR" {
			return resp.L7Rule, "error", fmt.Errorf("The L7 rule is in error status. " +
				"Please check with your cloud admin or check the API logs.")
		}

		return resp.L7Rule, resp.L7Rule.Status, nil
	}
}
