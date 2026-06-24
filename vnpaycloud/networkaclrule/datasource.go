package networkaclrule

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

func DataSourceNetworkACLRule() *schema.Resource {
	return &schema.Resource{
		ReadContext: dataSourceNetworkACLRuleRead,
		Schema: map[string]*schema.Schema{
			"id": {
				Type:     schema.TypeString,
				Optional: true,
				Computed: true,
			},
			"nacl_id": {
				Type:     schema.TypeString,
				Optional: true,
				Computed: true,
			},
			"name": {
				Type:     schema.TypeString,
				Optional: true,
				Computed: true,
			},
			"priority": {
				Type:     schema.TypeInt,
				Computed: true,
			},
			"type": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"action": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"port_start": {
				Type:     schema.TypeInt,
				Computed: true,
			},
			"port_end": {
				Type:     schema.TypeInt,
				Computed: true,
			},
			"source": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"destination": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"icmp_type": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"description": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"status": {
				Type:     schema.TypeString,
				Computed: true,
			},
		},
	}
}

func dataSourceNetworkACLRuleRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	cfg := meta.(*config.Config)

	if id, ok := d.GetOk("id"); ok && id.(string) != "" {
		resp := &dto.NetworkACLRuleResponse{}
		_, err := cfg.Client.Get(ctx, client.ApiPath.NetworkACLRuleWithID(cfg.ProjectID, id.(string)), resp, nil)
		if err != nil {
			return diag.Errorf("Error retrieving vnpaycloud_network_acl_rule %s: %s", id, err)
		}
		setNetworkACLRuleAttributes(d, resp.Rule)
		return nil
	}

	naclID, hasNaclID := d.GetOk("nacl_id")
	if !hasNaclID || naclID.(string) == "" {
		return diag.Errorf("nacl_id is required when querying vnpaycloud_network_acl_rule by name")
	}

	path := client.ApiPath.NetworkACLRules(cfg.ProjectID) + "?nacl_id=" + url.QueryEscape(naclID.(string))

	listResp := &dto.ListNetworkACLRulesResponse{}
	_, err := cfg.Client.Get(ctx, path, listResp, nil)
	if err != nil {
		return diag.Errorf("Unable to query vnpaycloud_network_acl_rule: %s", err)
	}

	name := d.Get("name").(string)
	var matched []dto.NetworkACLRule
	for _, rule := range listResp.NetworkACLRules {
		if name != "" && rule.Name != name {
			continue
		}
		matched = append(matched, rule)
	}

	if len(matched) < 1 {
		return diag.Errorf("Your vnpaycloud_network_acl_rule query returned no results")
	}
	if len(matched) > 1 {
		return diag.Errorf("Your vnpaycloud_network_acl_rule query returned multiple results")
	}

	tflog.Debug(ctx, "Retrieved vnpaycloud_network_acl_rule datasource", map[string]interface{}{"rule": matched[0]})
	setNetworkACLRuleAttributes(d, matched[0])

	return nil
}
