package vpc

import (
	"context"
	"fmt"
	"net/http"
	"terraform-provider-vnpaycloud/vnpaycloud/dto"
	"terraform-provider-vnpaycloud/vnpaycloud/helper/client"
	"terraform-provider-vnpaycloud/vnpaycloud/util"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/retry"
)

func vpcStateRefreshFunc(ctx context.Context, c *client.Client, projectID, vpcID string) retry.StateRefreshFunc {
	return func() (interface{}, string, error) {
		vpcResp := &dto.VPCResponse{}
		_, err := c.Get(ctx, client.ApiPath.VPCWithID(projectID, vpcID), vpcResp, nil)

		if err != nil {
			if util.ResponseCodeIs(err, http.StatusNotFound) {
				return vpcResp.VPC, "deleted", nil
			}

			return nil, "", err
		}

		if vpcResp.VPC.Status == "failed" {
			return vpcResp.VPC, vpcResp.VPC.Status, fmt.Errorf("The VPC is in error status. " +
				"Please check with your cloud admin or check the VPC " +
				"API logs to see why this error occurred.")
		}

		return vpcResp.VPC, vpcResp.VPC.Status, nil
	}
}
