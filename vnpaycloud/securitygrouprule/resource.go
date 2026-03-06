package securitygrouprule

import (
	"context"
	"terraform-provider-vnpaycloud/vnpaycloud/config"
	"terraform-provider-vnpaycloud/vnpaycloud/dto"
	"terraform-provider-vnpaycloud/vnpaycloud/helper/client"
	"terraform-provider-vnpaycloud/vnpaycloud/util"

	"github.com/hashicorp/terraform-plugin-log/tflog"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

func ResourceSecurityGroupRule() *schema.Resource {
	return &schema.Resource{
		CreateContext: resourceSecurityGroupRuleCreate,
		ReadContext:   resourceSecurityGroupRuleRead,
		DeleteContext: resourceSecurityGroupRuleDelete,
		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},
		Schema: map[string]*schema.Schema{
			"security_group_id": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},
			"direction": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},
			"protocol": {
				Type:     schema.TypeString,
				Optional: true,
				ForceNew: true,
			},
			"ethertype": {
				Type:     schema.TypeString,
				Optional: true,
				ForceNew: true,
				Default:  "IPv4",
			},
			"port_range_min": {
				Type:     schema.TypeInt,
				Optional: true,
				ForceNew: true,
			},
			"port_range_max": {
				Type:     schema.TypeInt,
				Optional: true,
				ForceNew: true,
			},
			"remote_ip_prefix": {
				Type:     schema.TypeString,
				Optional: true,
				ForceNew: true,
			},
			"remote_group_id": {
				Type:     schema.TypeString,
				Optional: true,
				ForceNew: true,
			},
		},
	}
}

func resourceSecurityGroupRuleCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	cfg := meta.(*config.Config)

	createOpts := dto.CreateSecurityGroupRuleRequest{
		SecurityGroupID: d.Get("security_group_id").(string),
		Direction:       d.Get("direction").(string),
		Protocol:        d.Get("protocol").(string),
		EtherType:       d.Get("ethertype").(string),
		PortRangeMin:    int32(d.Get("port_range_min").(int)),
		PortRangeMax:    int32(d.Get("port_range_max").(int)),
		RemoteIPPrefix:  d.Get("remote_ip_prefix").(string),
		RemoteGroupID:   d.Get("remote_group_id").(string),
	}

	tflog.Debug(ctx, "vnpaycloud_security_group_rule create options", map[string]interface{}{"create_opts": createOpts})

	createResp := &dto.SecurityGroupRuleResponse{}
	_, err := cfg.Client.Post(ctx, client.ApiPath.SecurityGroupRules(cfg.ProjectID), createOpts, createResp, nil)
	if err != nil {
		return diag.Errorf("Error creating vnpaycloud_security_group_rule: %s", err)
	}

	d.SetId(createResp.Rule.ID)

	return resourceSecurityGroupRuleRead(ctx, d, meta)
}

func resourceSecurityGroupRuleRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	cfg := meta.(*config.Config)

	ruleResp := &dto.SecurityGroupRuleResponse{}
	_, err := cfg.Client.Get(ctx, client.ApiPath.SecurityGroupRuleWithID(cfg.ProjectID, d.Id()), ruleResp, nil)
	if err != nil {
		return diag.FromErr(util.CheckNotFound(d, err, "Error retrieving vnpaycloud_security_group_rule"))
	}

	tflog.Debug(ctx, "Retrieved vnpaycloud_security_group_rule "+d.Id(), map[string]interface{}{"rule": ruleResp.Rule})

	d.Set("security_group_id", ruleResp.Rule.SecurityGroupID)
	d.Set("direction", ruleResp.Rule.Direction)
	d.Set("protocol", ruleResp.Rule.Protocol)
	d.Set("ethertype", ruleResp.Rule.EtherType)
	d.Set("port_range_min", ruleResp.Rule.PortRangeMin)
	d.Set("port_range_max", ruleResp.Rule.PortRangeMax)
	d.Set("remote_ip_prefix", ruleResp.Rule.RemoteIPPrefix)
	d.Set("remote_group_id", ruleResp.Rule.RemoteGroupID)

	return nil
}

func resourceSecurityGroupRuleDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	cfg := meta.(*config.Config)

	_, err := cfg.Client.Delete(ctx, client.ApiPath.SecurityGroupRuleWithID(cfg.ProjectID, d.Id()), nil)
	if err != nil {
		return diag.FromErr(util.CheckDeleted(d, err, "Error deleting vnpaycloud_security_group_rule"))
	}

	return nil
}
