package loadbalancer

import (
	"context"
	"fmt"
	"net/http"
	"terraform-provider-vnpaycloud/vnpaycloud/dto"
	"terraform-provider-vnpaycloud/vnpaycloud/helper/client"
	"terraform-provider-vnpaycloud/vnpaycloud/util"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/retry"
)

func loadBalancerStateRefreshFunc(ctx context.Context, c *client.Client, projectID, lbID string) retry.StateRefreshFunc {
	return func() (interface{}, string, error) {
		lbResp := &dto.LoadBalancerResponse{}
		_, err := c.Get(ctx, client.ApiPath.LoadBalancerWithID(projectID, lbID), lbResp, nil)

		if err != nil {
			if util.ResponseCodeIs(err, http.StatusNotFound) {
				return lbResp.LoadBalancer, "deleted", nil
			}
			return nil, "", err
		}

		if lbResp.LoadBalancer.Status == "failed" || lbResp.LoadBalancer.Status == "error" {
			return lbResp.LoadBalancer, lbResp.LoadBalancer.Status, fmt.Errorf("The load balancer is in error status. " +
				"Please check with your cloud admin or check the API logs.")
		}

		return lbResp.LoadBalancer, lbResp.LoadBalancer.Status, nil
	}
}
