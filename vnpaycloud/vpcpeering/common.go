package vpcpeering

import (
	"context"
	"fmt"
	"net/http"
	"terraform-provider-vnpaycloud/vnpaycloud/dto"
	"terraform-provider-vnpaycloud/vnpaycloud/helper/client"
	"terraform-provider-vnpaycloud/vnpaycloud/util"

	"github.com/hashicorp/terraform-plugin-log/tflog"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/retry"
)

func vpcPeeringStateRefreshFunc(ctx context.Context, c *client.Client, peeringID string) retry.StateRefreshFunc {
	return func() (interface{}, string, error) {
		peeringResp := &dto.PeeringConnectionResponse{}
		_, err := c.Get(ctx, client.ApiPath.PeeringConnectionWithID(peeringID), peeringResp, nil)

		if err != nil {
			if util.ResponseCodeIs(err, http.StatusNotFound) {
				return peeringResp.PeeringConnection, "deleted", nil
			}
			return nil, "", err
		}

		if peeringResp.PeeringConnection.Status == "failed" || peeringResp.PeeringConnection.Status == "error" {
			return peeringResp.PeeringConnection, peeringResp.PeeringConnection.Status, fmt.Errorf("The VPC peering connection is in error status. " +
				"Please check with your cloud admin or check the API logs.")
		}

		return peeringResp.PeeringConnection, peeringResp.PeeringConnection.Status, nil
	}
}

// findReversePeeringID finds the reverse direction peering by listing all peerings
// and matching src/dest VPC IDs in the opposite direction.
func findReversePeeringID(ctx context.Context, c *client.Client, peeringID string) string {
	// First get the primary peering to know src/dest VPCs
	peeringResp := &dto.PeeringConnectionResponse{}
	_, err := c.Get(ctx, client.ApiPath.PeeringConnectionWithID(peeringID), peeringResp, nil)
	if err != nil {
		tflog.Warn(ctx, "Failed to get peering for reverse lookup", map[string]interface{}{"peering_id": peeringID, "error": err.Error()})
		return ""
	}

	srcVpcID := peeringResp.PeeringConnection.SrcVpcID
	destVpcID := peeringResp.PeeringConnection.DestVpcID

	// List all peerings and find the reverse direction
	listResp := &dto.ListPeeringConnectionsResponse{}
	_, err = c.Get(ctx, client.ApiPath.PeeringConnections(), listResp, nil)
	if err != nil {
		tflog.Warn(ctx, "Failed to list peerings for reverse lookup", map[string]interface{}{"error": err.Error()})
		return ""
	}

	for _, p := range listResp.PeeringConnections {
		if p.ID != peeringID && p.SrcVpcID == destVpcID && p.DestVpcID == srcVpcID {
			tflog.Info(ctx, "Found reverse peering", map[string]interface{}{"primary_id": peeringID, "reverse_id": p.ID})
			return p.ID
		}
	}

	tflog.Warn(ctx, "Reverse peering not found", map[string]interface{}{"peering_id": peeringID})
	return ""
}
