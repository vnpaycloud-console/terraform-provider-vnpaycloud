package kubernetescluster

import (
	"context"
	"terraform-provider-vnpaycloud/vnpaycloud/dto"
	"terraform-provider-vnpaycloud/vnpaycloud/helper/client"
	"time"

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

// WaitClusterActive blocks until the cluster leaves an updating state. Day-2
// cluster actions (whitelist, rotate credentials, rolling update, audit log)
// all move the cluster to "updating" and back to "active" when done.
func WaitClusterActive(ctx context.Context, c *client.Client, projectID, id string, timeout time.Duration) error {
	stateConf := &retry.StateChangeConf{
		Pending:    []string{"updating", "creating", "initiating", "pending_update", "unknown"},
		Target:     []string{"active"},
		Refresh:    clusterStateRefreshFunc(ctx, c, projectID, id),
		Timeout:    timeout,
		Delay:      15 * time.Second,
		MinTimeout: 10 * time.Second,
	}
	_, err := stateConf.WaitForStateContext(ctx)
	return err
}
