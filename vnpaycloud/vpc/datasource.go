package vpc

import (
	"context"
	"terraform-provider-vnpaycloud/vnpaycloud/config"
	"terraform-provider-vnpaycloud/vnpaycloud/dto"
	"terraform-provider-vnpaycloud/vnpaycloud/helper/client"

	"github.com/hashicorp/terraform-plugin-log/tflog"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
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
	vpcClient, err := client.NewClient(ctx, config.ConsoleClientConfig)

	if err != nil {
		return diag.Errorf("Error creating VNPAY Cloud VPC client: %s", err)
	}

	listOpts := dto.ListVpcParams{
		ID:   d.Get("id").(string),
		Name: d.Get("name").(string),
		CIDR: d.Get("cidr_block").(string),
	}

	tflog.Debug(ctx, "vnpaycloud_vpc listOpts", map[string]interface{}{"listOpts": listOpts})

	listVpcResp := dto.ListVpcResponse{}

	_, err = vpcClient.Get(ctx, client.ApiPath.VPCWithParams(listOpts), &listVpcResp, nil)

	if err != nil {
		return diag.Errorf("Unable to query vnpaycloud_vpc: %s", err)
	}

	if len(listVpcResp.VPCs) > 1 {
		return diag.Errorf("Your vnpaycloud_vpc query returned multiple results")
	}

	if len(listVpcResp.VPCs) < 1 {
		return diag.Errorf("Your vnpaycloud_vpc query returned no results")
	}

	dataSourceVPCAttributes(ctx, d, listVpcResp.VPCs[0])

	return nil
}

func dataSourceVPCAttributes(ctx context.Context, d *schema.ResourceData, volume dto.Vpc) {
	d.SetId(volume.ID)
	d.Set("name", volume.Name)
	d.Set("description", volume.Name)
	d.Set("cidr_block", volume.CIDR)
	d.Set("enable_snat", volume.EnableSNAT)
}
