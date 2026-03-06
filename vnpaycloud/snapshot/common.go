package snapshot

import (
	"context"
	"fmt"
	"net/http"
	"terraform-provider-vnpaycloud/vnpaycloud/dto"
	"terraform-provider-vnpaycloud/vnpaycloud/helper/client"
	"terraform-provider-vnpaycloud/vnpaycloud/util"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/retry"
)

func snapshotStateRefreshFunc(ctx context.Context, c *client.Client, projectID, snapshotID string) retry.StateRefreshFunc {
	return func() (interface{}, string, error) {
		snapResp := &dto.SnapshotResponse{}
		_, err := c.Get(ctx, client.ApiPath.SnapshotWithID(projectID, snapshotID), snapResp, nil)

		if err != nil {
			if util.ResponseCodeIs(err, http.StatusNotFound) {
				return snapResp.Snapshot, "deleted", nil
			}
			return nil, "", err
		}

		if snapResp.Snapshot.Status == "failed" || snapResp.Snapshot.Status == "error" {
			return snapResp.Snapshot, snapResp.Snapshot.Status, fmt.Errorf("The snapshot is in error status. " +
				"Please check with your cloud admin or check the API logs.")
		}

		return snapResp.Snapshot, snapResp.Snapshot.Status, nil
	}
}
