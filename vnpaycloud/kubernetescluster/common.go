package kubernetescluster

import (
	"context"
	"terraform-provider-vnpaycloud/vnpaycloud/dto"
	"terraform-provider-vnpaycloud/vnpaycloud/helper/client"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/retry"
)

const maxTransientClusterErrors = 20

func clusterStateRefreshFunc(ctx context.Context, c *client.Client, projectID, id string) retry.StateRefreshFunc {
	transientErrors := 0
	return func() (interface{}, string, error) {
		resp := &dto.K8sClusterResponse{}
		httpResp, err := c.Get(ctx, client.ApiPath.ClusterWithID(projectID, id), resp, nil)
		if err != nil {
			if httpResp != nil && httpResp.StatusCode == 404 {
				return resp, "deleted", nil
			}
			transientErrors++
			if transientErrors > maxTransientClusterErrors {
				return nil, "", err
			}

			return resp, "unknown", nil
		}

		transientErrors = 0
		status := resp.Cluster.Status
		if status == "" {
			status = "unknown"
		}
		return resp, status, nil
	}
}
