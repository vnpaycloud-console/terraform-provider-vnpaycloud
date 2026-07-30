package networkacl

import (
	"context"
	"net/url"
	"terraform-provider-vnpaycloud/vnpaycloud/config"
	"terraform-provider-vnpaycloud/vnpaycloud/dto"
	"terraform-provider-vnpaycloud/vnpaycloud/helper/client"

	"github.com/hashicorp/terraform-plugin-log/tflog"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

func DataSourceNetworkACL() *schema.Resource {
	return &schema.Resource{
		ReadContext: dataSourceNetworkACLRead,
		Schema: map[string]*schema.Schema{
			"id": {
				Type:         schema.TypeString,
				Optional:     true,
				Computed:     true,
				AtLeastOneOf: []string{"id", "name", "vpc_id"},
			},
			"name": {
				Type:         schema.TypeString,
				Optional:     true,
				Computed:     true,
				AtLeastOneOf: []string{"id", "name", "vpc_id"},
			},
			"description": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"vpc_id": {
				Type:         schema.TypeString,
				Optional:     true,
				Computed:     true,
				AtLeastOneOf: []string{"id", "name", "vpc_id"},
			},
			"subnet_ids": {
				Type:     schema.TypeSet,
				Computed: true,
				Elem:     &schema.Schema{Type: schema.TypeString},
			},
			"total_rules": {
				Type:     schema.TypeInt,
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
		},
	}
}

func dataSourceNetworkACLRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	cfg := meta.(*config.Config)

	if id, ok := d.GetOk("id"); ok && id.(string) != "" {
		resp := &dto.NetworkACLResponse{}
		_, err := cfg.Client.Get(ctx, client.ApiPath.NetworkACLWithID(cfg.ProjectID, id.(string)), resp, nil)
		if err != nil {
			return diag.Errorf("Error retrieving vnpaycloud_network_acl %s: %s", id, err)
		}
		if v, ok := d.GetOk("name"); ok && resp.NetworkACL.Name != v.(string) {
			return diag.Errorf("vnpaycloud_network_acl %q does not match name %q", id.(string), v.(string))
		}
		if v, ok := d.GetOk("vpc_id"); ok && resp.NetworkACL.VpcID != v.(string) {
			return diag.Errorf("vnpaycloud_network_acl %q does not match vpc_id %q", id.(string), v.(string))
		}
		setNetworkACLAttributes(d, resp.NetworkACL)
		return nil
	}

	path := client.ApiPath.NetworkACLs(cfg.ProjectID)
	if vpcID, ok := d.GetOk("vpc_id"); ok && vpcID.(string) != "" {
		path += "?vpc_id=" + url.QueryEscape(vpcID.(string))
	}

	listResp := &dto.ListNetworkACLsResponse{}
	_, err := cfg.Client.Get(ctx, path, listResp, nil)
	if err != nil {
		return diag.Errorf("Unable to query vnpaycloud_network_acl: %s", err)
	}

	name := d.Get("name").(string)
	var matched []dto.NetworkACL
	for _, acl := range listResp.NetworkACLs {
		if name != "" && acl.Name != name {
			continue
		}
		matched = append(matched, acl)
	}

	if len(matched) < 1 {
		return diag.Errorf("Your vnpaycloud_network_acl query returned no results")
	}
	if len(matched) > 1 {
		return diag.Errorf("Your vnpaycloud_network_acl query returned multiple results")
	}

	tflog.Debug(ctx, "Retrieved vnpaycloud_network_acl datasource", map[string]interface{}{"network_acl": matched[0]})
	setNetworkACLAttributes(d, matched[0])

	return nil
}
