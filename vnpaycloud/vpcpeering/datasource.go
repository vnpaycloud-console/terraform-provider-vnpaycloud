package vpcpeering

import (
	"context"
	"fmt"
	"terraform-provider-vnpaycloud/vnpaycloud/config"
	"terraform-provider-vnpaycloud/vnpaycloud/dto"
	"terraform-provider-vnpaycloud/vnpaycloud/helper/client"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

func DataSourceVPCPeering() *schema.Resource {
	return &schema.Resource{
		ReadContext: dataSourceVPCPeeringRead,
		Schema: map[string]*schema.Schema{
			"id": {
				Type:     schema.TypeString,
				Optional: true,
				Computed: true,
			},
			"name": {
				Type:     schema.TypeString,
				Optional: true,
				Computed: true,
			},
			"src_vpc_id": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"dest_vpc_id": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"status": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"peering_status": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"src_vpc_cidr": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"dest_vpc_cidr": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"created_at": {
				Type:     schema.TypeString,
				Computed: true,
			},
		},
	}
}

func dataSourceVPCPeeringRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	cfg := meta.(*config.Config)

	if id, ok := d.GetOk("id"); ok {
		peeringResp := &dto.PeeringConnectionResponse{}
		_, err := cfg.Client.Get(ctx, client.ApiPath.PeeringConnectionWithID(id.(string)), peeringResp, nil)
		if err != nil {
			return diag.Errorf("Error fetching vnpaycloud_vpc_peering %s: %s", id, err)
		}
		return setVPCPeeringData(d, &peeringResp.PeeringConnection)
	}

	// List and filter client-side
	listResp := &dto.ListPeeringConnectionsResponse{}
	_, err := cfg.Client.Get(ctx, client.ApiPath.PeeringConnections(), listResp, nil)
	if err != nil {
		return diag.Errorf("Error listing vnpaycloud_vpc_peering: %s", err)
	}

	nameFilter, nameOk := d.GetOk("name")

	for _, p := range listResp.PeeringConnections {
		if nameOk && p.Name != nameFilter.(string) {
			continue
		}
		return setVPCPeeringData(d, &p)
	}

	return diag.Errorf("No vnpaycloud_vpc_peering found matching the criteria")
}

func setVPCPeeringData(d *schema.ResourceData, peering *dto.PeeringConnection) diag.Diagnostics {
	d.SetId(peering.ID)
	d.Set("name", peering.Name)
	d.Set("src_vpc_id", peering.SrcVpcID)
	d.Set("dest_vpc_id", peering.DestVpcID)
	d.Set("status", peering.Status)
	d.Set("peering_status", peering.PeeringStatus)
	d.Set("src_vpc_cidr", peering.SrcVpcCIDR)
	d.Set("dest_vpc_cidr", peering.DestVpcCIDR)
	d.Set("created_at", peering.CreatedAt)
	return nil
}

func DataSourceVPCPeerings() *schema.Resource {
	return &schema.Resource{
		ReadContext: dataSourceVPCPeeringsRead,
		Schema: map[string]*schema.Schema{
			"vpc_peerings": {
				Type:     schema.TypeList,
				Computed: true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"id":             {Type: schema.TypeString, Computed: true},
						"name":           {Type: schema.TypeString, Computed: true},
						"src_vpc_id":     {Type: schema.TypeString, Computed: true},
						"dest_vpc_id":    {Type: schema.TypeString, Computed: true},
						"status":         {Type: schema.TypeString, Computed: true},
						"peering_status": {Type: schema.TypeString, Computed: true},
						"src_vpc_cidr":   {Type: schema.TypeString, Computed: true},
						"dest_vpc_cidr":  {Type: schema.TypeString, Computed: true},
						"created_at":     {Type: schema.TypeString, Computed: true},
					},
				},
			},
		},
	}
}

func dataSourceVPCPeeringsRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	cfg := meta.(*config.Config)

	listResp := &dto.ListPeeringConnectionsResponse{}
	_, err := cfg.Client.Get(ctx, client.ApiPath.PeeringConnections(), listResp, nil)
	if err != nil {
		return diag.Errorf("Error listing vnpaycloud_vpc_peerings: %s", err)
	}

	var vpcPeerings []map[string]interface{}
	for _, p := range listResp.PeeringConnections {
		vpcPeerings = append(vpcPeerings, map[string]interface{}{
			"id":             p.ID,
			"name":           p.Name,
			"src_vpc_id":     p.SrcVpcID,
			"dest_vpc_id":    p.DestVpcID,
			"status":         p.Status,
			"peering_status": p.PeeringStatus,
			"src_vpc_cidr":   p.SrcVpcCIDR,
			"dest_vpc_cidr":  p.DestVpcCIDR,
			"created_at":     p.CreatedAt,
		})
	}

	d.SetId(fmt.Sprintf("vpc-peerings-%s", cfg.ProjectID))
	d.Set("vpc_peerings", vpcPeerings)

	return nil
}
