package vpc

import (
	"context"
	"fmt"
	"net/http"
	"terraform-provider-vnpaycloud/vnpaycloud/helper/client"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/retry"
	"github.com/vnpaycloud-console/gophercloud/v2"
)

func vpcStateRefreshFunc(ctx context.Context, vpcClient *client.Client, vpcId string) retry.StateRefreshFunc {
	return func() (interface{}, string, error) {
		vpcResp := &GetVpcDtoResponse{}
		_, err := vpcClient.Get(ctx, client.ApiPath.VPCWithId(vpcId), vpcResp, nil)

		if err != nil {
			if gophercloud.ResponseCodeIs(err, http.StatusNotFound) {
				return vpcResp.VPC, "OS_DELETED", nil
			}

			return nil, "", err
		}

		if vpcResp.VPC.Status == "OS_FAILED" {
			return vpcResp.VPC, vpcResp.VPC.Status, fmt.Errorf("The VPC is in error status. " +
				"Please check with your cloud admin or check the VPC " +
				"API logs to see why this error occurred.")
		}

		return vpcResp.VPC, vpcResp.VPC.Status, nil
	}
}
