package vpc

import (
	"context"
	"fmt"
	"net/http"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/retry"
	"github.com/vnpaycloud-console/gophercloud/v2"
	"github.com/vnpaycloud-console/gophercloud/v2/openstack/networking/v2/vpcs"
)

func vpcStateRefreshFunc(ctx context.Context, client *gophercloud.ServiceClient, vpcId string) retry.StateRefreshFunc {
	return func() (interface{}, string, error) {
		vpc, err := vpcs.Get(ctx, client, vpcId).Extract()
		if err != nil {
			if gophercloud.ResponseCodeIs(err, http.StatusNotFound) {
				return vpc, "OS_DELETED", nil
			}

			return nil, "", err
		}

		if vpc.Status == "OS_FAILED" {
			return vpc, vpc.Status, fmt.Errorf("The VPC is in error status. " +
				"Please check with your cloud admin or check the VPC " +
				"API logs to see why this error occurred.")
		}

		return vpc, vpc.Status, nil
	}
}
