package instance

import (
	"context"
	"fmt"
	"net/http"
	"terraform-provider-vnpaycloud/vnpaycloud/dto"
	"terraform-provider-vnpaycloud/vnpaycloud/helper/client"
	"terraform-provider-vnpaycloud/vnpaycloud/util"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/retry"
)

func instanceStateRefreshFunc(ctx context.Context, c *client.Client, projectID, instanceID string) retry.StateRefreshFunc {
	return func() (interface{}, string, error) {
		instResp := &dto.InstanceResponse{}
		_, err := c.Get(ctx, client.ApiPath.InstanceWithID(projectID, instanceID), instResp, nil)

		if err != nil {
			if util.ResponseCodeIs(err, http.StatusNotFound) {
				return instResp.Instance, "deleted", nil
			}
			return nil, "", err
		}

		if instResp.Instance.Status == "error" || instResp.Instance.Status == "failed" {
			return instResp.Instance, instResp.Instance.Status, fmt.Errorf("The instance is in error status. " +
				"Please check with your cloud admin or check the API logs.")
		}

		return instResp.Instance, instResp.Instance.Status, nil
	}
}
