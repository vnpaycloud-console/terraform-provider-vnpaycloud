package backuppolicyserver

import (
	"context"
	"regexp"
	"terraform-provider-vnpaycloud/vnpaycloud/config"
	"terraform-provider-vnpaycloud/vnpaycloud/dto"
	"terraform-provider-vnpaycloud/vnpaycloud/helper/client"
	"terraform-provider-vnpaycloud/vnpaycloud/util"
	"time"

	"github.com/hashicorp/terraform-plugin-log/tflog"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
)

var nameRegex = regexp.MustCompile(`^[a-zA-Z0-9-_. ]*$`)

const purposeOnDemand = "on_demand"

func ResourceBackupPolicyServer() *schema.Resource {
	return &schema.Resource{
		CreateContext: resourceBackupPolicyServerCreate,
		ReadContext:   resourceBackupPolicyServerRead,
		UpdateContext: resourceBackupPolicyServerUpdate,
		DeleteContext: resourceBackupPolicyServerDelete,
		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},
		Timeouts: &schema.ResourceTimeout{
			Create: schema.DefaultTimeout(10 * time.Minute),
			Delete: schema.DefaultTimeout(10 * time.Minute),
		},
		CustomizeDiff: func(ctx context.Context, d *schema.ResourceDiff, meta interface{}) error {
			if err := validateMonthlyBlock("monthly", d.Get("monthly"), false); err != nil {
				return err
			}
			return validateMonthlyBlock("second_tier_monthly", d.Get("second_tier_monthly"), true)
		},
		Schema: map[string]*schema.Schema{
			"name": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
				ValidateFunc: validation.All(
					validation.StringLenBetween(3, 512),
					validation.StringMatch(nameRegex, "must contain only letters, digits, spaces, hyphens, underscores and dots"),
				),
			},
			"description": {
				Type:         schema.TypeString,
				Optional:     true,
				ValidateFunc: validation.StringLenBetween(0, 512),
			},
			"start_hour": {
				Type:         schema.TypeInt,
				Required:     true,
				ValidateFunc: validation.IntBetween(0, 23),
			},
			"resource_type": {
				Type:         schema.TypeString,
				Required:     true,
				ForceNew:     true,
				ValidateFunc: validation.StringInSlice([]string{"vm", "volume"}, false),
			},
			"purpose": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"is_auto": {
				Type:     schema.TypeBool,
				Optional: true,
				Default:  false,
			},
			"is_auto_apply_for_volume": {
				Type:     schema.TypeBool,
				Optional: true,
				Default:  false,
			},
			"backup_vault_ids": {
				Type:     schema.TypeList,
				Required: true,
				ForceNew: true,
				MinItems: 1,
				Elem:     &schema.Schema{Type: schema.TypeString},
			},

			"daily": {
				Type:     schema.TypeList,
				Required: true,
				MaxItems: 1,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"retentions": {
							Type:         schema.TypeInt,
							Required:     true,
							ValidateFunc: validation.IntBetween(7, 14),
						},
					},
				},
			},
			"weekly": {
				Type:     schema.TypeList,
				Optional: true,
				MaxItems: 1,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"retentions": {
							Type:         schema.TypeInt,
							Required:     true,
							ValidateFunc: validation.IntBetween(1, 14),
						},
						"day_of_week": {
							Type:         schema.TypeInt,
							Required:     true,
							ValidateFunc: validation.IntBetween(1, 7),
						},
					},
				},
			},
			"monthly": {
				Type:     schema.TypeList,
				Optional: true,
				MaxItems: 1,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"retentions": {
							Type:         schema.TypeInt,
							Required:     true,
							ValidateFunc: validation.IntBetween(1, 14),
						},
						"day_type": {
							Type:         schema.TypeString,
							Required:     true,
							ValidateFunc: validation.StringInSlice(monthlyDayTypes, false),
						},
						"day_of_month": {
							Type:         schema.TypeInt,
							Optional:     true,
							ValidateFunc: validation.IntBetween(1, 31),
						},
						"day_of_week": {
							Type:         schema.TypeInt,
							Optional:     true,
							ValidateFunc: validation.IntBetween(1, 7),
						},
					},
				},
			},

			"second_tier_enabled": {
				Type:     schema.TypeBool,
				Optional: true,
				Default:  false,
			},
			"second_tier_weekly": {
				Type:     schema.TypeList,
				Optional: true,
				MaxItems: 1,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"retentions": {
							Type:         schema.TypeInt,
							Required:     true,
							ValidateFunc: validation.IntBetween(1, 14),
						},
						"day_of_week": {
							Type:         schema.TypeInt,
							Required:     true,
							ValidateFunc: validation.IntBetween(0, 6),
						},
					},
				},
			},
			"second_tier_monthly": {
				Type:     schema.TypeList,
				Optional: true,
				MaxItems: 1,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"retentions": {
							Type:         schema.TypeInt,
							Required:     true,
							ValidateFunc: validation.IntBetween(1, 14),
						},
						"day_type": {
							Type:         schema.TypeString,
							Required:     true,
							ValidateFunc: validation.StringInSlice(monthlyDayTypes, false),
						},
						"day_of_month": {
							Type:         schema.TypeInt,
							Optional:     true,
							ValidateFunc: validation.IntBetween(1, 31),
						},
						"day_of_week": {
							Type:         schema.TypeInt,
							Optional:     true,
							ValidateFunc: validation.IntBetween(0, 6),
						},
					},
				},
			},
			"second_tier_yearly": {
				Type:     schema.TypeList,
				Optional: true,
				MaxItems: 1,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"retentions": {
							Type:         schema.TypeInt,
							Required:     true,
							ValidateFunc: validation.IntBetween(1, 14),
						},
						"month": {
							Type:         schema.TypeInt,
							Required:     true,
							ValidateFunc: validation.IntBetween(1, 12),
						},
						"day_of_month": {
							Type:         schema.TypeInt,
							Required:     true,
							ValidateFunc: validation.IntBetween(1, 31),
						},
					},
				},
			},

			"protected_servers":      {Type: schema.TypeInt, Computed: true},
			"protected_volumes":      {Type: schema.TypeInt, Computed: true},
			"protected_volume_sizes": {Type: schema.TypeInt, Computed: true},
			"compliance_state":       {Type: schema.TypeString, Computed: true},
			"compliance_msg":         {Type: schema.TypeString, Computed: true},
			"status":                 {Type: schema.TypeString, Computed: true},
			"created_at":             {Type: schema.TypeString, Computed: true},
		},
	}
}

func resourceBackupPolicyServerCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	cfg := meta.(*config.Config)

	createOpts := dto.CreateBackupPolicyServerRequest{
		Name:                 d.Get("name").(string),
		Description:          d.Get("description").(string),
		StartHour:            d.Get("start_hour").(int),
		ResourceType:         d.Get("resource_type").(string),
		Purpose:              purposeOnDemand,
		IsAuto:               d.Get("is_auto").(bool),
		IsAutoApplyForVolume: d.Get("is_auto_apply_for_volume").(bool),
		BackupVaultIDs:       expandStringList(d.Get("backup_vault_ids")),
		Daily:                expandDaily(d.Get("daily")),
		Weekly:               expandWeekly(d.Get("weekly")),
		Monthly:              expandMonthly(d.Get("monthly")),
		SecondTierEnabled:    d.Get("second_tier_enabled").(bool),
	}

	if createOpts.SecondTierEnabled {
		createOpts.SecondTierWeekly = expandSecondTierWeekly(d.Get("second_tier_weekly"))
		createOpts.SecondTierMonthly = expandSecondTierMonthly(d.Get("second_tier_monthly"))
		createOpts.SecondTierYearly = expandSecondTierYearly(d.Get("second_tier_yearly"))
	}

	tflog.Debug(ctx, "vnpaycloud_backup_policy_server create options", map[string]interface{}{
		"name":          createOpts.Name,
		"resource_type": createOpts.ResourceType,
		"purpose":       createOpts.Purpose,
	})

	createResp := &dto.BackupPolicyServerResponse{}
	if _, err := cfg.Client.Post(ctx, client.ApiPath.BackupPolicyServers(cfg.ProjectID), createOpts, createResp, nil); err != nil {
		return diag.Errorf("Error creating vnpaycloud_backup_policy_server: %s", err)
	}

	d.SetId(createResp.BackupPolicyServer.ID)

	return resourceBackupPolicyServerRead(ctx, d, meta)
}

func resourceBackupPolicyServerRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	cfg := meta.(*config.Config)

	resp := &dto.BackupPolicyServerResponse{}
	if _, err := cfg.Client.Get(ctx, client.ApiPath.BackupPolicyServerWithID(cfg.ProjectID, d.Id()), resp, nil); err != nil {
		return diag.FromErr(util.CheckNotFound(d, err, "Error retrieving vnpaycloud_backup_policy_server"))
	}

	setBackupPolicyServerData(d, &resp.BackupPolicyServer)

	return nil
}

func resourceBackupPolicyServerUpdate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	cfg := meta.(*config.Config)

	updateOpts := dto.UpdateBackupPolicyServerRequest{
		Description:          d.Get("description").(string),
		StartHour:            d.Get("start_hour").(int),
		IsAuto:               d.Get("is_auto").(bool),
		IsAutoApplyForVolume: d.Get("is_auto_apply_for_volume").(bool),
		BackupVaultIDs:       expandStringList(d.Get("backup_vault_ids")),
		Daily:                expandDaily(d.Get("daily")),
		Weekly:               expandWeekly(d.Get("weekly")),
		Monthly:              expandMonthly(d.Get("monthly")),
		SecondTierEnabled:    d.Get("second_tier_enabled").(bool),
	}

	if updateOpts.SecondTierEnabled {
		updateOpts.SecondTierWeekly = expandSecondTierWeekly(d.Get("second_tier_weekly"))
		updateOpts.SecondTierMonthly = expandSecondTierMonthly(d.Get("second_tier_monthly"))
		updateOpts.SecondTierYearly = expandSecondTierYearly(d.Get("second_tier_yearly"))
	}

	if _, err := cfg.Client.Put(ctx, client.ApiPath.BackupPolicyServerWithID(cfg.ProjectID, d.Id()), updateOpts, nil, nil); err != nil {
		return diag.Errorf("Error updating vnpaycloud_backup_policy_server %s: %s", d.Id(), err)
	}

	return resourceBackupPolicyServerRead(ctx, d, meta)
}

func resourceBackupPolicyServerDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	cfg := meta.(*config.Config)

	if _, err := cfg.Client.Delete(ctx, client.ApiPath.BackupPolicyServerWithID(cfg.ProjectID, d.Id()), nil); err != nil {
		return diag.FromErr(util.CheckDeleted(d, err, "Error deleting vnpaycloud_backup_policy_server"))
	}

	return nil
}

func expandStringList(v interface{}) []string {
	l, ok := v.([]interface{})
	if !ok {
		return nil
	}
	result := make([]string, 0, len(l))
	for _, item := range l {
		if s, ok := item.(string); ok {
			result = append(result, s)
		}
	}
	return result
}
