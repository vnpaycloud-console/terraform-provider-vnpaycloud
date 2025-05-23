package shared

import (
	"context"
	"fmt"
	"terraform-provider-vnpaycloud/vnpaycloud/config"
	"terraform-provider-vnpaycloud/vnpaycloud/helper/client"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

type PeeringConnection struct {
	ID             string `json:"id"`
	PeerStatus     string `json:"peering_status"`
	Description    string `json:"description"`
	ConnectionType string `json:"connection_type"`
	Status         string `json:"status"`
	VpcId          string `json:"src_vpc_id"`
	PeerOrgId      string `json:"dest_org_id"`
	PeerVpcId      string `json:"dest_vpc_id"`
	PortId         string `json:"port_peering_connection_id"`
}

type GetPeeringConnectionResponse struct {
	PeeringConnection PeeringConnection `json:"peering_connection"`
}

func PeeringConnectionId2PortId(ctx context.Context, d *schema.ResourceData, meta interface{}, id string) (string, error) {
	config := meta.(*config.Config)
	c, err := client.NewClient(ctx, config.ConsoleClientConfig)

	if err != nil {
		return "", fmt.Errorf("Error creating VNPAY Cloud Peering Connection client: %s", err)
	}

	getPeeringConnectionResp := &GetPeeringConnectionResponse{}

	_, err = c.Get(ctx, client.ApiPath.PeeringConnectionWithId(id), getPeeringConnectionResp, nil)

	if err != nil {
		return "", fmt.Errorf("Error retrieving peering connection: %s", err)
	}

	return getPeeringConnectionResp.PeeringConnection.PortId, nil
}
