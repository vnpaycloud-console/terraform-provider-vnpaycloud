package securitygroup

import (
	"context"
	"fmt"
	"net/http"
	"terraform-provider-vnpaycloud/vnpaycloud/dto"
	"terraform-provider-vnpaycloud/vnpaycloud/helper/client"
	"terraform-provider-vnpaycloud/vnpaycloud/util"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/retry"
)

func securityGroupStateRefreshFunc(ctx context.Context, c *client.Client, projectID, sgID string) retry.StateRefreshFunc {
	return func() (interface{}, string, error) {
		sgResp := &dto.SecurityGroupResponse{}
		_, err := c.Get(ctx, client.ApiPath.SecurityGroupWithID(projectID, sgID), sgResp, nil)

		if err != nil {
			if util.ResponseCodeIs(err, http.StatusNotFound) {
				return sgResp.SecurityGroup, "deleted", nil
			}
			return nil, "", err
		}

		if sgResp.SecurityGroup.Status == "failed" {
			return sgResp.SecurityGroup, sgResp.SecurityGroup.Status, fmt.Errorf("The security group is in error status. " +
				"Please check with your cloud admin or check the API logs.")
		}

		return sgResp.SecurityGroup, sgResp.SecurityGroup.Status, nil
	}
}
