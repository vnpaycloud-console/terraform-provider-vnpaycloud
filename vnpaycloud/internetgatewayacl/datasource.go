package internetgatewayacl

import (
	"context"
	"terraform-provider-vnpaycloud/vnpaycloud/config"
	"terraform-provider-vnpaycloud/vnpaycloud/dto"
	"terraform-provider-vnpaycloud/vnpaycloud/helper/client"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

func DataSourceInternetGatewayACL() *schema.Resource {
	return &schema.Resource{
		ReadContext: dataSourceInternetGatewayACLRead,
		Schema: map[string]*schema.Schema{
			"internet_gateway_id": {
				Type:     schema.TypeString,
				Required: true,
			},
			"match": {
				Type:     schema.TypeSet,
				Computed: true,
				Elem:     &schema.Schema{Type: schema.TypeString},
			},
			"total_match": {
				Type:     schema.TypeInt,
				Computed: true,
			},
			"zone_id": {
				Type:     schema.TypeString,
				Computed: true,
			},
		},
	}
}

func dataSourceInternetGatewayACLRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	cfg := meta.(*config.Config)

	igwID := d.Get("internet_gateway_id").(string)

	aclResp := &dto.InternetGatewayBlacklistResponse{}
	_, err := cfg.Client.Get(ctx, client.ApiPath.InternetGatewayBlacklist(cfg.ProjectID, igwID), aclResp, nil)
	if err != nil {
		return diag.Errorf("Error fetching vnpaycloud_internet_gateway_acl of internet gateway %s: %s", igwID, err)
	}

	d.SetId(igwID)
	d.Set("internet_gateway_id", igwID)
	d.Set("match", aclResp.Blacklist.Match)
	d.Set("total_match", len(aclResp.Blacklist.Match))
	d.Set("zone_id", aclResp.Blacklist.ZoneID)

	return nil
}
