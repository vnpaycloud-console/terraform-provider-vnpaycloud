package backupkubernetesrestorepoint

import (
	"context"
	"regexp"
	"time"

	"terraform-provider-vnpaycloud/vnpaycloud/config"
	"terraform-provider-vnpaycloud/vnpaycloud/dto"
	"terraform-provider-vnpaycloud/vnpaycloud/helper/client"
	"terraform-provider-vnpaycloud/vnpaycloud/util"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/retry"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
)

var idRegex = regexp.MustCompile(`^[a-zA-Z0-9-_.]*$`)

func ResourceBackupKubernetesRestorePoint() *schema.Resource {
	return &schema.Resource{
		CreateContext: resourceRestorePointCreate,
		ReadContext:   resourceRestorePointRead,
		DeleteContext: resourceRestorePointDelete,
		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},
		Timeouts: &schema.ResourceTimeout{
			Delete: schema.DefaultTimeout(15 * time.Minute),
		},
		Schema: map[string]*schema.Schema{
			"restore_point_id": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
				ValidateFunc: validation.All(
					validation.StringLenBetween(3, 100),
					validation.StringMatch(idRegex, "must contain only letters, digits, hyphens, underscores and dots"),
				),
			},

			"name":                 {Type: schema.TypeString, Computed: true},
			"cluster_id":           {Type: schema.TypeString, Computed: true},
			"cluster_name":         {Type: schema.TypeString, Computed: true},
			"cluster_backup_id":    {Type: schema.TypeString, Computed: true},
			"pvc_backup_policy_id": {Type: schema.TypeString, Computed: true},
			"project_id":           {Type: schema.TypeString, Computed: true},
			"zone_id":              {Type: schema.TypeString, Computed: true},
			"backup_vault_ids":     {Type: schema.TypeList, Computed: true, Elem: &schema.Schema{Type: schema.TypeString}},
			"velero_backup_name":   {Type: schema.TypeString, Computed: true},
			"backup_point":         {Type: schema.TypeString, Computed: true},
			"kubernetes_version":   {Type: schema.TypeString, Computed: true},
			"status":               {Type: schema.TypeString, Computed: true},
		},
	}
}

func resourceRestorePointCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	cfg := meta.(*config.Config)

	id := d.Get("restore_point_id").(string)
	resp := &dto.BackupKubernetesRestorePointResponse{}
	if _, err := cfg.Client.Get(ctx, client.ApiPath.BackupKubernetesRestorePointWithID(cfg.ProjectID, id), resp, nil); err != nil {
		return diag.Errorf("Error reading vnpaycloud_backup_kubernetes_restore_point %s: %s", id, err)
	}

	d.SetId(resp.RestorePoint.ID)
	setBackupKubernetesRestorePointData(d, &resp.RestorePoint)

	return nil
}

func resourceRestorePointRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	cfg := meta.(*config.Config)

	resp := &dto.BackupKubernetesRestorePointResponse{}
	if _, err := cfg.Client.Get(ctx, client.ApiPath.BackupKubernetesRestorePointWithID(cfg.ProjectID, d.Id()), resp, nil); err != nil {
		return diag.FromErr(util.CheckNotFound(d, err, "Error retrieving vnpaycloud_backup_kubernetes_restore_point"))
	}

	setBackupKubernetesRestorePointData(d, &resp.RestorePoint)

	return nil
}

func resourceRestorePointDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	cfg := meta.(*config.Config)

	if _, err := cfg.Client.Delete(ctx, client.ApiPath.BackupKubernetesRestorePointWithID(cfg.ProjectID, d.Id()), nil); err != nil {
		return diag.FromErr(util.CheckDeleted(d, err, "Error deleting vnpaycloud_backup_kubernetes_restore_point"))
	}

	stateConf := &retry.StateChangeConf{
		Pending:    []string{"deleting", "active", "unknown"},
		Target:     []string{"deleted"},
		Refresh:    restorePointDeleteStateRefreshFunc(ctx, cfg.Client, cfg.ProjectID, d.Id()),
		Timeout:    d.Timeout(schema.TimeoutDelete),
		Delay:      10 * time.Second,
		MinTimeout: 15 * time.Second,
	}
	if _, err := stateConf.WaitForStateContext(ctx); err != nil {
		return diag.Errorf("Error waiting for vnpaycloud_backup_kubernetes_restore_point %s to be deleted: %s", d.Id(), err)
	}

	return nil
}

func setBackupKubernetesRestorePointData(d *schema.ResourceData, p *dto.BackupKubernetesRestorePoint) {
	d.Set("restore_point_id", p.ID)
	d.Set("name", p.Name)
	d.Set("cluster_id", p.ClusterID)
	d.Set("cluster_name", p.ClusterName)
	d.Set("cluster_backup_id", p.ClusterBackupID)
	d.Set("pvc_backup_policy_id", p.PVCBackupPolicyID)
	d.Set("project_id", p.ProjectID)
	d.Set("zone_id", p.ZoneID)
	d.Set("backup_vault_ids", p.BackupVaultIDs)
	d.Set("velero_backup_name", p.VeleroBackupName)
	d.Set("backup_point", p.BackupPoint)
	d.Set("kubernetes_version", p.KubernetesVersion)
	d.Set("status", p.Status)
}
