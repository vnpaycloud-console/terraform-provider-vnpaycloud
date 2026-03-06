package internetgateway

import (
	"context"
	"fmt"
	"terraform-provider-vnpaycloud/vnpaycloud/config"
	"terraform-provider-vnpaycloud/vnpaycloud/dto"
	"terraform-provider-vnpaycloud/vnpaycloud/helper/client"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

func DataSourceInternetGateway() *schema.Resource {
	return &schema.Resource{
		ReadContext: dataSourceInternetGatewayRead,
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
			"vpc_id": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"status": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"created_at": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"zone_id": {
				Type:     schema.TypeString,
				Computed: true,
			},
		},
	}
}

func dataSourceInternetGatewayRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	cfg := meta.(*config.Config)

	if id, ok := d.GetOk("id"); ok {
		igwResp := &dto.InternetGatewayResponse{}
		_, err := cfg.Client.Get(ctx, client.ApiPath.InternetGatewayWithID(cfg.ProjectID, id.(string)), igwResp, nil)
		if err != nil {
			return diag.Errorf("Error fetching vnpaycloud_internet_gateway %s: %s", id, err)
		}
		return setInternetGatewayData(d, &igwResp.InternetGateway)
	}

	// List and filter client-side
	listResp := &dto.ListInternetGatewaysResponse{}
	_, err := cfg.Client.Get(ctx, client.ApiPath.InternetGateways(cfg.ProjectID), listResp, nil)
	if err != nil {
		return diag.Errorf("Error listing vnpaycloud_internet_gateway: %s", err)
	}

	nameFilter, nameOk := d.GetOk("name")

	for _, igw := range listResp.InternetGateways {
		if nameOk && igw.Name != nameFilter.(string) {
			continue
		}
		return setInternetGatewayData(d, &igw)
	}

	return diag.Errorf("No vnpaycloud_internet_gateway found matching the criteria")
}

func setInternetGatewayData(d *schema.ResourceData, igw *dto.InternetGateway) diag.Diagnostics {
	d.SetId(igw.ID)
	d.Set("name", igw.Name)
	d.Set("description", igw.Description)
	d.Set("vpc_id", igw.VPCID)
	d.Set("status", igw.Status)
	d.Set("created_at", igw.CreatedAt)
	d.Set("zone_id", igw.ZoneID)
	return nil
}

func DataSourceInternetGateways() *schema.Resource {
	return &schema.Resource{
		ReadContext: dataSourceInternetGatewaysRead,
		Schema: map[string]*schema.Schema{
			"internet_gateways": {
				Type:     schema.TypeList,
				Computed: true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"id":          {Type: schema.TypeString, Computed: true},
						"name":        {Type: schema.TypeString, Computed: true},
						"description": {Type: schema.TypeString, Computed: true},
						"vpc_id":      {Type: schema.TypeString, Computed: true},
						"status":      {Type: schema.TypeString, Computed: true},
						"created_at":  {Type: schema.TypeString, Computed: true},
						"zone_id":     {Type: schema.TypeString, Computed: true},
					},
				},
			},
		},
	}
}

func dataSourceInternetGatewaysRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	cfg := meta.(*config.Config)

	listResp := &dto.ListInternetGatewaysResponse{}
	_, err := cfg.Client.Get(ctx, client.ApiPath.InternetGateways(cfg.ProjectID), listResp, nil)
	if err != nil {
		return diag.Errorf("Error listing vnpaycloud_internet_gateways: %s", err)
	}

	var internetGateways []map[string]interface{}
	for _, igw := range listResp.InternetGateways {
		internetGateways = append(internetGateways, map[string]interface{}{
			"id":          igw.ID,
			"name":        igw.Name,
			"description": igw.Description,
			"vpc_id":      igw.VPCID,
			"status":      igw.Status,
			"created_at":  igw.CreatedAt,
			"zone_id":     igw.ZoneID,
		})
	}

	d.SetId(fmt.Sprintf("internet-gateways-%s", cfg.ProjectID))
	d.Set("internet_gateways", internetGateways)

	return nil
}
