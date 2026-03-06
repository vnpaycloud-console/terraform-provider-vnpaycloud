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
			"cidr": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"status": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"enable_snat": {
				Type:     schema.TypeBool,
				Computed: true,
			},
			"snat_address": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"subnet_ids": {
				Type:     schema.TypeList,
				Computed: true,
				Elem:     &schema.Schema{Type: schema.TypeString},
			},
			"created_at": {
				Type:     schema.TypeString,
				Computed: true,
			},
		},
	}
}

func dataSourceVpcRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	cfg := meta.(*config.Config)

	// If ID is provided, fetch directly
	if id, ok := d.GetOk("id"); ok && id.(string) != "" {
		vpcResp := &dto.VPCResponse{}
		_, err := cfg.Client.Get(ctx, client.ApiPath.VPCWithID(cfg.ProjectID, id.(string)), vpcResp, nil)
		if err != nil {
			return diag.Errorf("Error retrieving vnpaycloud_vpc %s: %s", id, err)
		}
		setVPCDataSourceAttributes(d, vpcResp.VPC)
		return nil
	}

	// Otherwise, list and filter by name
	listResp := &dto.ListVPCsResponse{}
	_, err := cfg.Client.Get(ctx, client.ApiPath.VPCs(cfg.ProjectID), listResp, nil)
	if err != nil {
		return diag.Errorf("Unable to query vnpaycloud_vpc: %s", err)
	}

	name := d.Get("name").(string)
	var matched []dto.VPC
	for _, v := range listResp.VPCs {
		if name != "" && v.Name != name {
			continue
		}
		matched = append(matched, v)
	}

	if len(matched) < 1 {
		return diag.Errorf("Your vnpaycloud_vpc query returned no results")
	}

	if len(matched) > 1 {
		return diag.Errorf("Your vnpaycloud_vpc query returned multiple results")
	}

	tflog.Debug(ctx, "Retrieved vnpaycloud_vpc datasource", map[string]interface{}{"vpc": matched[0]})
	setVPCDataSourceAttributes(d, matched[0])

	return nil
}

func setVPCDataSourceAttributes(d *schema.ResourceData, vpc dto.VPC) {
	d.SetId(vpc.ID)
	d.Set("name", vpc.Name)
	d.Set("description", vpc.Description)
	d.Set("cidr", vpc.CIDR)
	d.Set("status", vpc.Status)
	d.Set("enable_snat", vpc.EnableSnat)
	d.Set("snat_address", vpc.SnatAddress)
	d.Set("subnet_ids", vpc.SubnetIDs)
	d.Set("created_at", vpc.CreatedAt)
}
