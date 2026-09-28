package backupkubernetes

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
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/retry"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
)

var idRegex = regexp.MustCompile(`^[a-zA-Z0-9-_.]*$`)

func ResourceBackupKubernetes() *schema.Resource {
	return &schema.Resource{
		CreateContext: resourceBackupKubernetesCreate,
		ReadContext:   resourceBackupKubernetesRead,
		UpdateContext: resourceBackupKubernetesUpdate,
		DeleteContext: resourceBackupKubernetesDelete,
		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},
		Timeouts: &schema.ResourceTimeout{
			Create: schema.DefaultTimeout(15 * time.Minute),
			Delete: schema.DefaultTimeout(15 * time.Minute),
		},
		Schema: map[string]*schema.Schema{
			"protected_cluster_id": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
				ValidateFunc: validation.All(
					validation.StringLenBetween(3, 100),
					validation.StringMatch(idRegex, "must contain only letters, digits, hyphens, underscores and dots"),
				),
			},
			"pvc_backup_policy_id": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
				ValidateFunc: validation.All(
					validation.StringLenBetween(3, 100),
					validation.StringMatch(idRegex, "must contain only letters, digits, hyphens, underscores and dots"),
				),
			},
			"description": {
				Type:         schema.TypeString,
				Optional:     true,
				ValidateFunc: validation.StringLenBetween(0, 512),
			},

			"name":                {Type: schema.TypeString, Computed: true},
			"cluster_backup_type": {Type: schema.TypeString, Computed: true},
			"zone_id":             {Type: schema.TypeString, Computed: true},
			"project_id":          {Type: schema.TypeString, Computed: true},
			"customer_username":   {Type: schema.TypeString, Computed: true},
			"status":              {Type: schema.TypeString, Computed: true},
			"created_at":          {Type: schema.TypeString, Computed: true},
		},
	}
}

func resourceBackupKubernetesCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	cfg := meta.(*config.Config)

	createOpts := dto.CreateBackupKubernetesRequest{
		ProtectedClusterID: d.Get("protected_cluster_id").(string),
		PVCBackupPolicyID:  d.Get("pvc_backup_policy_id").(string),
		Description:        d.Get("description").(string),
	}

	tflog.Debug(ctx, "vnpaycloud_backup_kubernetes create options", map[string]interface{}{
		"protected_cluster_id": createOpts.ProtectedClusterID,
		"pvc_backup_policy_id": createOpts.PVCBackupPolicyID,
	})

	createResp := &dto.BackupKubernetesResponse{}
	if _, err := cfg.Client.Post(ctx, client.ApiPath.BackupKuberneteses(cfg.ProjectID), createOpts, createResp, nil); err != nil {
		return diag.Errorf("Error creating vnpaycloud_backup_kubernetes: %s", err)
	}

	d.SetId(createResp.BackupKubernetes.ID)

	stateConf := &retry.StateChangeConf{
		Pending:    []string{"creating", "initiating", "pending_create", "unknown"},
		Target:     []string{"active"},
		Refresh:    backupKubernetesStateRefreshFunc(ctx, cfg.Client, cfg.ProjectID, d.Id()),
		Timeout:    d.Timeout(schema.TimeoutCreate),
		Delay:      30 * time.Second,
		MinTimeout: 15 * time.Second,
	}
	if _, err := stateConf.WaitForStateContext(ctx); err != nil {
		return diag.Errorf("Error waiting for vnpaycloud_backup_kubernetes %s to become active: %s", d.Id(), err)
	}

	return resourceBackupKubernetesRead(ctx, d, meta)
}

func resourceBackupKubernetesRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	cfg := meta.(*config.Config)

	resp := &dto.BackupKubernetesResponse{}
	if _, err := cfg.Client.Get(ctx, client.ApiPath.BackupKubernetesWithID(cfg.ProjectID, d.Id()), resp, nil); err != nil {
		return diag.FromErr(util.CheckNotFound(d, err, "Error retrieving vnpaycloud_backup_kubernetes"))
	}

	setBackupKubernetesData(d, &resp.BackupKubernetes)

	return nil
}

func resourceBackupKubernetesUpdate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	cfg := meta.(*config.Config)

	updateOpts := dto.UpdateBackupKubernetesRequest{
		Description: d.Get("description").(string),
	}

	if _, err := cfg.Client.Put(ctx, client.ApiPath.BackupKubernetesWithID(cfg.ProjectID, d.Id()), updateOpts, nil, nil); err != nil {
		return diag.Errorf("Error updating vnpaycloud_backup_kubernetes %s: %s", d.Id(), err)
	}

	return resourceBackupKubernetesRead(ctx, d, meta)
}

func resourceBackupKubernetesDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	cfg := meta.(*config.Config)

	if _, err := cfg.Client.Delete(ctx, client.ApiPath.BackupKubernetesWithID(cfg.ProjectID, d.Id()), nil); err != nil {
		return diag.FromErr(util.CheckDeleted(d, err, "Error deleting vnpaycloud_backup_kubernetes"))
	}

	return nil
}
