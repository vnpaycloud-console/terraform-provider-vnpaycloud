package vpc

import (
	"context"
	"terraform-provider-vnpaycloud/vnpaycloud/config"
	"terraform-provider-vnpaycloud/vnpaycloud/util"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/vnpaycloud-console/gophercloud/v2/openstack/networking/v2/vpcs"
)

func DataSourceVpc() *schema.Resource {
	return &schema.Resource{
		ReadContext: dataSourceVpcRead,
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
			"description": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"cidr_block": {
				Type:     schema.TypeString,
				Optional: true,
				Computed: true,
			},
			"enable_snat": {
				Type:     schema.TypeBool,
				Computed: true,
			},
		},
	}
}

func dataSourceVpcRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	config := meta.(*config.Config)
	client, err := config.NetworkingV2Client(ctx, util.GetRegion(d, config))

	if err != nil {
		return diag.Errorf("Error creating VNPAY Cloud VPC client: %s", err)
	}

	listOpts := vpcs.ListOpts{
		ID:   d.Get("id").(string),
		Name: d.Get("name").(string),
		CIDR: d.Get("cidr_block").(string),
	}

	allPages, err := vpcs.List(client, listOpts).AllPages(ctx)

	if err != nil {
		return diag.Errorf("Unable to query vnpaycloud_vpc: %s", err)
	}

	var allVpcs []vpcs.VPC
	err = vpcs.ExtractVPCsInto(allPages, &allVpcs)

	if err != nil {
		return diag.Errorf("Unable to retrieve vnpaycloud_vpc: %s", err)
	}

	if len(allVpcs) > 1 {
		return diag.Errorf("Your vnpaycloud_vpc query returned multiple results")
	}

	if len(allVpcs) < 1 {
		return diag.Errorf("Your vnpaycloud_vpc query returned no results")
	}

	dataSourceVPCAttributes(ctx, d, allVpcs[0])

	return nil
}

func dataSourceVPCAttributes(ctx context.Context, d *schema.ResourceData, volume vpcs.VPC) {
	d.SetId(volume.ID)
	d.Set("name", volume.Name)
	d.Set("description", volume.Name)
	d.Set("cidr_block", volume.CIDR)
	d.Set("enable_snat", volume.EnableSNAT)
}
