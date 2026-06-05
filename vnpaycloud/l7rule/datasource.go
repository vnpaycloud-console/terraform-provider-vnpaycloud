package l7rule

import (
	"context"
	"terraform-provider-vnpaycloud/vnpaycloud/config"
	"terraform-provider-vnpaycloud/vnpaycloud/dto"
	"terraform-provider-vnpaycloud/vnpaycloud/helper/client"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

func DataSourceL7Rule() *schema.Resource {
	return &schema.Resource{
		ReadContext: dataSourceL7RuleRead,
		Schema: map[string]*schema.Schema{
			"id": {
				Type:     schema.TypeString,
				Required: true,
			},
			"l7policy_id": {
				Type:     schema.TypeString,
				Required: true,
			},
			"rule_type": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"compare_type": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"value": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"key": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"invert": {
				Type:     schema.TypeBool,
				Computed: true,
			},
			"status": {
				Type:     schema.TypeString,
				Computed: true,
			},
		},
	}
}

func dataSourceL7RuleRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	cfg := meta.(*config.Config)
	id := d.Get("id").(string)
	l7policyID := d.Get("l7policy_id").(string)

	resp := &dto.L7RuleResponse{}
	_, err := cfg.Client.Get(ctx, client.ApiPath.L7RuleWithID(cfg.ProjectID, l7policyID, id), resp, nil)
	if err != nil {
		return diag.Errorf("Error retrieving vnpaycloud_lb_l7rule %s: %s", id, err)
	}

	d.SetId(resp.L7Rule.ID)
	d.Set("l7policy_id", resp.L7Rule.L7PolicyID)
	d.Set("rule_type", resp.L7Rule.RuleType)
	d.Set("compare_type", resp.L7Rule.CompareType)
	d.Set("value", resp.L7Rule.Value)
	d.Set("key", resp.L7Rule.Key)
	d.Set("invert", resp.L7Rule.Invert)
	d.Set("status", resp.L7Rule.Status)

	return nil
}
