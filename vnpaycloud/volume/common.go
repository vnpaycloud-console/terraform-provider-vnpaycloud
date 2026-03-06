package volume

import (
	"context"
	"fmt"
	"net/http"
	"terraform-provider-vnpaycloud/vnpaycloud/dto"
	"terraform-provider-vnpaycloud/vnpaycloud/helper/client"
	"terraform-provider-vnpaycloud/vnpaycloud/util"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/retry"
)

func volumeStateRefreshFunc(ctx context.Context, c *client.Client, projectID, volumeID string) retry.StateRefreshFunc {
	return func() (interface{}, string, error) {
		volResp := &dto.VolumeResponse{}
		_, err := c.Get(ctx, client.ApiPath.VolumeWithID(projectID, volumeID), volResp, nil)

		if err != nil {
			if util.ResponseCodeIs(err, http.StatusNotFound) {
				return volResp.Volume, "deleted", nil
			}
			return nil, "", err
		}

		if volResp.Volume.Status == "failed" {
			return volResp.Volume, volResp.Volume.Status, fmt.Errorf("The volume is in error status. " +
				"Please check with your cloud admin or check the API logs.")
		}

		return volResp.Volume, volResp.Volume.Status, nil
	}
}
