package subnet

import (
	"context"
	"fmt"
	"net/http"
	"terraform-provider-vnpaycloud/vnpaycloud/dto"
	"terraform-provider-vnpaycloud/vnpaycloud/helper/client"
	"terraform-provider-vnpaycloud/vnpaycloud/util"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/retry"
)

func subnetStateRefreshFunc(ctx context.Context, c *client.Client, projectID, subnetID string) retry.StateRefreshFunc {
	return func() (interface{}, string, error) {
		subnetResp := &dto.SubnetResponse{}
		_, err := c.Get(ctx, client.ApiPath.SubnetWithID(projectID, subnetID), subnetResp, nil)

		if err != nil {
			if util.ResponseCodeIs(err, http.StatusNotFound) {
				return subnetResp.Subnet, "deleted", nil
			}
			return nil, "", err
		}

		if subnetResp.Subnet.Status == "failed" {
			return subnetResp.Subnet, subnetResp.Subnet.Status, fmt.Errorf("The subnet is in error status. " +
				"Please check with your cloud admin or check the API logs.")
		}

		return subnetResp.Subnet, subnetResp.Subnet.Status, nil
	}
}
