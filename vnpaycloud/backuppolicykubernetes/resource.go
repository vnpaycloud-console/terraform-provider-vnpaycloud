package backuppolicykubernetes

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

var nameRegex = regexp.MustCompile(`^[a-zA-Z0-9-_.]*$`)

func ResourceBackupPolicyKubernetes() *schema.Resource {
	return &schema.Resource{
		CreateContext: resourceBackupPolicyKubernetesCreate,
		ReadContext:   resourceBackupPolicyKubernetesRead,
		UpdateContext: resourceBackupPolicyKubernetesUpdate,
		DeleteContext: resourceBackupPolicyKubernetesDelete,
		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},
		Timeouts: &schema.ResourceTimeout{
			Create: schema.DefaultTimeout(10 * time.Minute),
			Delete: schema.DefaultTimeout(10 * time.Minute),
		},
		CustomizeDiff: func(ctx context.Context, d *schema.ResourceDiff, meta interface{}) error {
			if d.Get("is_auto").(bool) {
				if err := d.SetNew("start_hour", 0); err != nil {
					return err
				}
			}
			return validateMonthlyDayType(d)
		},
		Schema: map[string]*schema.Schema{
			"name": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
				ValidateFunc: validation.All(
					validation.StringLenBetween(3, 50),
					validation.StringMatch(nameRegex, "must contain only letters, digits, hyphens, underscores and dots (no spaces)"),
				),
			},
			"description": {
				Type:         schema.TypeString,
				Optional:     true,
				ValidateFunc: validation.StringLenBetween(0, 512),
			},
			"auto_apply_new_volume": {
				Type:     schema.TypeBool,
				Optional: true,
				Default:  false,
			},
			"is_auto": {
				Type:     schema.TypeBool,
				Optional: true,
				Default:  false,
			},
			"start_hour": {
				Type:         schema.TypeInt,
				Optional:     true,
				Computed:     true,
				ValidateFunc: validation.IntBetween(0, 23),
			},
			"run_priority": {
				Type:         schema.TypeInt,
				Optional:     true,
				Computed:     true,
				ValidateFunc: validation.IntAtLeast(0),
			},
			"purpose": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"backup_vault_ids": {
				Type:     schema.TypeList,
				Required: true,
				ForceNew: true,
				MinItems: 1,
				MaxItems: 1,
				Elem:     &schema.Schema{Type: schema.TypeString},
			},

			"daily": {
				Type:     schema.TypeList,
				Optional: true,
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
							ValidateFunc: validation.IntBetween(0, 6),
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
							ValidateFunc: validation.IntBetween(0, 6),
						},
					},
				},
			},

			"status":     {Type: schema.TypeString, Computed: true},
			"created_at": {Type: schema.TypeString, Computed: true},
		},
	}
}

func resourceBackupPolicyKubernetesCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	cfg := meta.(*config.Config)

	createOpts := dto.CreateBackupPolicyKubernetesRequest{
		Name:               d.Get("name").(string),
		Description:        d.Get("description").(string),
		AutoApplyNewVolume: d.Get("auto_apply_new_volume").(bool),
		IsAuto:             d.Get("is_auto").(bool),
		StartHour:          d.Get("start_hour").(int),
		RunPriority:        d.Get("run_priority").(int),
		BackupVaultIDs:     expandStringList(d.Get("backup_vault_ids")),
		Daily:              expandDaily(d.Get("daily")),
		Weekly:             expandWeekly(d.Get("weekly")),
		Monthly:            expandMonthly(d.Get("monthly")),
	}

	tflog.Debug(ctx, "vnpaycloud_backup_policy_kubernetes create options", map[string]interface{}{
		"name":       createOpts.Name,
		"start_hour": createOpts.StartHour,
	})

	createResp := &dto.BackupPolicyKubernetesResponse{}
	if _, err := cfg.Client.Post(ctx, client.ApiPath.BackupPolicyKuberneteses(cfg.ProjectID), createOpts, createResp, nil); err != nil {
		return diag.Errorf("Error creating vnpaycloud_backup_policy_kubernetes: %s", err)
	}

	d.SetId(createResp.BackupPolicyKubernetes.ID)

	return resourceBackupPolicyKubernetesRead(ctx, d, meta)
}

func resourceBackupPolicyKubernetesRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	cfg := meta.(*config.Config)

	resp := &dto.BackupPolicyKubernetesResponse{}
	if _, err := cfg.Client.Get(ctx, client.ApiPath.BackupPolicyKubernetesWithID(cfg.ProjectID, d.Id()), resp, nil); err != nil {
		return diag.FromErr(util.CheckNotFound(d, err, "Error retrieving vnpaycloud_backup_policy_kubernetes"))
	}

	setBackupPolicyKubernetesData(d, &resp.BackupPolicyKubernetes)

	return nil
}

func resourceBackupPolicyKubernetesUpdate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	cfg := meta.(*config.Config)

	updateOpts := dto.UpdateBackupPolicyKubernetesRequest{
		Description:        d.Get("description").(string),
		AutoApplyNewVolume: d.Get("auto_apply_new_volume").(bool),
		IsAuto:             d.Get("is_auto").(bool),
		StartHour:          d.Get("start_hour").(int),
		RunPriority:        d.Get("run_priority").(int),
		Daily:              expandDaily(d.Get("daily")),
		Weekly:             expandWeekly(d.Get("weekly")),
		Monthly:            expandMonthly(d.Get("monthly")),
	}

	if _, err := cfg.Client.Put(ctx, client.ApiPath.BackupPolicyKubernetesWithID(cfg.ProjectID, d.Id()), updateOpts, nil, nil); err != nil {
		return diag.Errorf("Error updating vnpaycloud_backup_policy_kubernetes %s: %s", d.Id(), err)
	}

	return resourceBackupPolicyKubernetesRead(ctx, d, meta)
}

func resourceBackupPolicyKubernetesDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	cfg := meta.(*config.Config)

	if _, err := cfg.Client.Delete(ctx, client.ApiPath.BackupPolicyKubernetesWithID(cfg.ProjectID, d.Id()), nil); err != nil {
		return diag.FromErr(util.CheckDeleted(d, err, "Error deleting vnpaycloud_backup_policy_kubernetes"))
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
