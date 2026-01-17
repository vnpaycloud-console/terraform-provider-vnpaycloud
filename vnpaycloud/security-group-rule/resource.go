package securityGroupRule

import (
	"context"
	"log"
	"strings"
	"time"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/retry"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"

	"terraform-provider-vnpaycloud/vnpaycloud/config"
	"terraform-provider-vnpaycloud/vnpaycloud/dto"
	"terraform-provider-vnpaycloud/vnpaycloud/helper/client"
	"terraform-provider-vnpaycloud/vnpaycloud/util"
)

func ResourceNetworkingSecGroupRule() *schema.Resource {
	return &schema.Resource{
		CreateContext: resourceNetworkingSecGroupRuleCreate,
		ReadContext:   resourceNetworkingSecGroupRuleRead,
		DeleteContext: resourceNetworkingSecGroupRuleDelete,
		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},

		Timeouts: &schema.ResourceTimeout{
			Delete: schema.DefaultTimeout(10 * time.Minute),
		},

		Schema: map[string]*schema.Schema{
			"region": {
				Type:     schema.TypeString,
				Optional: true,
				Computed: true,
				ForceNew: true,
			},

			"description": {
				Type:     schema.TypeString,
				Optional: true,
				ForceNew: true,
			},

			"direction": {
				Type:         schema.TypeString,
				Required:     true,
				ForceNew:     true,
				ValidateFunc: ResourceNetworkingSecGroupRuleDirection,
			},

			"ethertype": {
				Type:         schema.TypeString,
				Required:     true,
				ForceNew:     true,
				ValidateFunc: ResourceNetworkingSecGroupRuleEtherType,
			},

			"port_range_min": {
				Type:         schema.TypeInt,
				Optional:     true,
				ForceNew:     true,
				RequiredWith: []string{"protocol", "port_range_max"},
				ValidateFunc: validation.IntBetween(0, 65535),
			},

			"port_range_max": {
				Type:         schema.TypeInt,
				Optional:     true,
				ForceNew:     true,
				RequiredWith: []string{"protocol", "port_range_min"},
				ValidateFunc: validation.IntBetween(0, 65535),
			},

			"protocol": {
				Type:         schema.TypeString,
				Optional:     true,
				ForceNew:     true,
				ValidateFunc: resourceNetworkingSecGroupRuleProtocol,
			},

			"remote_ip_prefix": {
				Type:     schema.TypeString,
				Optional: true,
				ForceNew: true,
				Computed: true,
				StateFunc: func(v interface{}) string {
					return strings.ToLower(v.(string))
				},
			},

			"security_group_id": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},

			"tenant_id": {
				Type:     schema.TypeString,
				Optional: true,
				ForceNew: true,
				Computed: true,
			},
		},
	}
}

func resourceNetworkingSecGroupRuleCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	config := meta.(*config.Config)
	tfClient, err := client.NewClient(ctx, config.ConsoleClientConfig)
	if err != nil {
		return diag.Errorf("Error creating VNPAY Cloud networking client: %s", err)
	}

	securityGroupID := d.Get("security_group_id").(string)
	config.MutexKV.Lock(securityGroupID)
	defer config.MutexKV.Unlock(securityGroupID)

	protocol := d.Get("protocol").(string)
	direction := d.Get("direction").(string)
	etherType := d.Get("ethertype").(string)
	createOpts := dto.CreateSecurityGroupRuleOpts{
		Direction:      dto.RuleDirection(direction),
		EtherType:      dto.RuleEtherType(etherType),
		Protocol:       dto.RuleProtocol(protocol),
		PortRangeMin:   d.Get("port_range_min").(int),
		PortRangeMax:   d.Get("port_range_max").(int),
		Description:    d.Get("description").(string),
		SecGroupID:     securityGroupID,
		RemoteIPPrefix: d.Get("remote_ip_prefix").(string),
		ProjectID:      d.Get("tenant_id").(string),
	}

	log.Printf("[DEBUG] vnpaycloud_networking_secgroup_rule create options: %#v", createOpts)

	createResp := &dto.CreateSecurityGroupRuleResponse{}
	_, err = tfClient.Post(ctx, client.ApiPath.SecurityGroupRule, dto.CreateSecurityGroupRuleRequest{SecurityGroupRule: createOpts}, createResp, nil)
	if err != nil {
		return diag.Errorf("Error creating vnpaycloud_networking_secgroup_rule: %s", err)
	}

	d.SetId(createResp.SecurityGroupRule.ID)

	log.Printf("[DEBUG] Created vnpaycloud_networking_secgroup_rule %s: %#v", createResp.SecurityGroupRule.ID, createResp.SecurityGroupRule)
	return resourceNetworkingSecGroupRuleRead(ctx, d, meta)
}

func resourceNetworkingSecGroupRuleRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	config := meta.(*config.Config)
	tfClient, err := client.NewClient(ctx, config.ConsoleClientConfig)
	if err != nil {
		return diag.Errorf("Error creating VNPAY Cloud networking client: %s", err)
	}

	securityGroupRuleResp := &dto.GetSecurityGroupRuleResponse{}
	_, err = tfClient.Get(ctx, client.ApiPath.SecurityGroupRuleWithId(d.Id()), securityGroupRuleResp, nil)
	if err != nil {
		return diag.FromErr(util.CheckDeleted(d, err, "Error getting vnpaycloud_networking_secgroup_rule"))
	}

	log.Printf("[DEBUG] Retrieved vnpaycloud_networking_secgroup_rule %s: %#v", d.Id(), securityGroupRuleResp.SecurityGroupRule)

	d.Set("description", securityGroupRuleResp.SecurityGroupRule.Description)
	d.Set("direction", securityGroupRuleResp.SecurityGroupRule.Direction)
	d.Set("ethertype", securityGroupRuleResp.SecurityGroupRule.EtherType)
	d.Set("protocol", securityGroupRuleResp.SecurityGroupRule.Protocol)
	d.Set("port_range_min", securityGroupRuleResp.SecurityGroupRule.PortRangeMin)
	d.Set("port_range_max", securityGroupRuleResp.SecurityGroupRule.PortRangeMax)
	d.Set("remote_ip_prefix", securityGroupRuleResp.SecurityGroupRule.RemoteIPPrefix)
	d.Set("security_group_id", securityGroupRuleResp.SecurityGroupRule.SecGroupID)
	d.Set("tenant_id", securityGroupRuleResp.SecurityGroupRule.TenantID)
	d.Set("region", util.GetRegion(d, config))

	return nil
}

func resourceNetworkingSecGroupRuleDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	config := meta.(*config.Config)
	tfClient, err := client.NewClient(ctx, config.ConsoleClientConfig)
	if err != nil {
		return diag.Errorf("Error creating VNPAY Cloud networking client: %s", err)
	}

	securityGroupID := d.Get("security_group_id").(string)
	config.MutexKV.Lock(securityGroupID)
	defer config.MutexKV.Unlock(securityGroupID)

	if _, err := tfClient.Delete(ctx, client.ApiPath.SecurityGroupWithId(securityGroupID), nil); err != nil {
		return diag.FromErr(util.CheckDeleted(d, err, "Error deleting vnpaycloud_networking_secgroup_rule"))
	}

	stateConf := &retry.StateChangeConf{
		Pending:    []string{"ACTIVE"},
		Target:     []string{"DELETED"},
		Refresh:    ResourceNetworkingSecGroupRuleStateRefreshFunc(ctx, tfClient, d.Id()),
		Timeout:    d.Timeout(schema.TimeoutDelete),
		Delay:      5 * time.Second,
		MinTimeout: 3 * time.Second,
	}

	_, err = stateConf.WaitForStateContext(ctx)
	if err != nil {
		return diag.Errorf("Error waiting for vnpaycloud_networking_secgroup_rule %s to Delete:  %s", d.Id(), err)
	}

	d.SetId("")
	return nil
}
