package networkaclrule

import (
	"context"
	"terraform-provider-vnpaycloud/vnpaycloud/config"
	"terraform-provider-vnpaycloud/vnpaycloud/dto"
	"terraform-provider-vnpaycloud/vnpaycloud/helper/client"
	"terraform-provider-vnpaycloud/vnpaycloud/util"
	"time"

	"github.com/hashicorp/terraform-plugin-log/tflog"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/retry"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
)

func ResourceNetworkACLRule() *schema.Resource {
	return &schema.Resource{
		CreateContext: resourceNetworkACLRuleCreate,
		ReadContext:   resourceNetworkACLRuleRead,
		DeleteContext: resourceNetworkACLRuleDelete,
		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},
		Timeouts: &schema.ResourceTimeout{
			Create: schema.DefaultTimeout(10 * time.Minute),
			Delete: schema.DefaultTimeout(10 * time.Minute),
		},
		Schema: map[string]*schema.Schema{
			"nacl_id": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},
			"name": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},
			"priority": {
				Type:         schema.TypeInt,
				Required:     true,
				ForceNew:     true,
				ValidateFunc: validation.IntBetween(1, 1000),
			},
			"type": {
				Type:         schema.TypeString,
				Required:     true,
				ForceNew:     true,
				ValidateFunc: validation.StringInSlice(networkACLRuleTypes, false),
			},
			"action": {
				Type:         schema.TypeString,
				Required:     true,
				ForceNew:     true,
				ValidateFunc: validation.StringInSlice([]string{"allow", "drop"}, false),
			},
			"port_start": {
				Type:         schema.TypeInt,
				Optional:     true,
				Computed:     true,
				ForceNew:     true,
				ValidateFunc: validation.IntBetween(0, 65535),
			},
			"port_end": {
				Type:         schema.TypeInt,
				Optional:     true,
				Computed:     true,
				ForceNew:     true,
				ValidateFunc: validation.IntBetween(0, 65535),
			},
			"source": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},
			"destination": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},
			"icmp_type": {
				Type:     schema.TypeString,
				Optional: true,
				Computed: true,
				ForceNew: true,
			},
			"description": {
				Type:     schema.TypeString,
				Optional: true,
				Computed: true,
				ForceNew: true,
			},
			"status": {
				Type:     schema.TypeString,
				Computed: true,
			},
		},
	}
}

func resourceNetworkACLRuleCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	if err := validateNetworkACLRuleConfig(d); err != nil {
		return diag.FromErr(err)
	}

	cfg := meta.(*config.Config)

	createOpts := dto.CreateNetworkACLRuleRequest{
		NaclID:      d.Get("nacl_id").(string),
		Name:        d.Get("name").(string),
		Priority:    int64(d.Get("priority").(int)),
		Type:        d.Get("type").(string),
		Action:      d.Get("action").(string),
		PortStart:   int32(d.Get("port_start").(int)),
		PortEnd:     int32(d.Get("port_end").(int)),
		Source:      d.Get("source").(string),
		Destination: d.Get("destination").(string),
		IcmpType:    d.Get("icmp_type").(string),
		Description: d.Get("description").(string),
	}

	tflog.Debug(ctx, "vnpaycloud_network_acl_rule create options", map[string]interface{}{"create_opts": createOpts})

	createResp := &dto.NetworkACLRuleResponse{}
	_, err := cfg.Client.Post(ctx, client.ApiPath.NetworkACLRules(cfg.ProjectID), createOpts, createResp, nil)
	if err != nil {
		return diag.Errorf("Error creating vnpaycloud_network_acl_rule: %s", err)
	}

	d.SetId(createResp.Rule.ID)

	// The backend provisions the rule asynchronously (status "creating") and only
	// fills in the computed protocol ports for preset types once it becomes "active".
	stateConf := &retry.StateChangeConf{
		Pending:    []string{"creating", "initiating", "unknown"},
		Target:     []string{"active", "created"},
		Refresh:    networkACLRuleStateRefreshFunc(ctx, cfg.Client, cfg.ProjectID, d.Id()),
		Timeout:    d.Timeout(schema.TimeoutCreate),
		Delay:      3 * time.Second,
		MinTimeout: 2 * time.Second,
	}

	if _, err = stateConf.WaitForStateContext(ctx); err != nil {
		return diag.Errorf("Error waiting for vnpaycloud_network_acl_rule %s to become ready: %s", d.Id(), err)
	}

	return resourceNetworkACLRuleRead(ctx, d, meta)
}

func resourceNetworkACLRuleRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	cfg := meta.(*config.Config)

	resp := &dto.NetworkACLRuleResponse{}
	_, err := cfg.Client.Get(ctx, client.ApiPath.NetworkACLRuleWithID(cfg.ProjectID, d.Id()), resp, nil)
	if err != nil {
		return diag.FromErr(util.CheckNotFound(d, err, "Error retrieving vnpaycloud_network_acl_rule"))
	}

	tflog.Debug(ctx, "Retrieved vnpaycloud_network_acl_rule "+d.Id(), map[string]interface{}{"rule": resp.Rule})
	setNetworkACLRuleAttributes(d, resp.Rule)

	return nil
}

func resourceNetworkACLRuleDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	cfg := meta.(*config.Config)

	_, err := cfg.Client.Delete(ctx, client.ApiPath.NetworkACLRuleWithID(cfg.ProjectID, d.Id()), nil)
	if err != nil {
		return diag.FromErr(util.CheckDeleted(d, err, "Error deleting vnpaycloud_network_acl_rule"))
	}

	return nil
}
