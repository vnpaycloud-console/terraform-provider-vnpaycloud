package backupvault

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

var backupVaultNameRegex = regexp.MustCompile(`^[a-zA-Z0-9-_.]*$`)

func ResourceBackupVault() *schema.Resource {
	return &schema.Resource{
		CreateContext: resourceBackupVaultCreate,
		ReadContext:   resourceBackupVaultRead,
		UpdateContext: resourceBackupVaultUpdate,
		DeleteContext: resourceBackupVaultDelete,
		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},
		Timeouts: &schema.ResourceTimeout{
			Create: schema.DefaultTimeout(10 * time.Minute),
			Delete: schema.DefaultTimeout(10 * time.Minute),
		},
		Schema: map[string]*schema.Schema{
			"name": {
				Type:     schema.TypeString,
				Required: true,
				ValidateFunc: validation.All(
					validation.StringLenBetween(3, 100),
					validation.StringMatch(backupVaultNameRegex, "must contain only letters, digits, hyphens, underscores and dots"),
				),
			},
			"purpose": {
				Type:         schema.TypeString,
				Required:     true,
				ForceNew:     true,
				ValidateFunc: validation.StringInSlice([]string{"backup_server", "workload_cluster"}, false),
			},
			"type": {
				Type:         schema.TypeString,
				Required:     true,
				ForceNew:     true,
				ValidateFunc: validation.StringInSlice([]string{"ceph", "s3"}, false),
			},
			"description": {
				Type:     schema.TypeString,
				Optional: true,
			},
			"object_lock": {
				Type:     schema.TypeBool,
				Computed: true,
			},
			"lock_time_unit": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"lock_time_number": {
				Type:     schema.TypeInt,
				Computed: true,
			},
			"zone_id": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"storage_location_id": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"disk_used": {
				Type:     schema.TypeInt,
				Computed: true,
			},
			"quota": {
				Type:     schema.TypeInt,
				Computed: true,
			},
			"committed_quota": {
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

func resourceBackupVaultCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	cfg := meta.(*config.Config)

	createOpts := dto.CreateBackupVaultRequest{
		Name:        d.Get("name").(string),
		Purpose:     d.Get("purpose").(string),
		Type:        d.Get("type").(string),
		Description: d.Get("description").(string),
	}

	tflog.Debug(ctx, "vnpaycloud_backup_vault create options", map[string]interface{}{
		"name":    createOpts.Name,
		"purpose": createOpts.Purpose,
		"type":    createOpts.Type,
	})

	createResp := &dto.BackupVaultResponse{}
	if _, err := cfg.Client.Post(ctx, client.ApiPath.BackupVaults(cfg.ProjectID), createOpts, createResp, nil); err != nil {
		return diag.Errorf("Error creating vnpaycloud_backup_vault: %s", err)
	}

	d.SetId(createResp.BackupVault.ID)

	stateConf := &retry.StateChangeConf{
		Pending:    []string{"creating", "unknown"},
		Target:     []string{"active"},
		Refresh:    backupVaultStateRefreshFunc(ctx, cfg.Client, cfg.ProjectID, d.Id()),
		Timeout:    d.Timeout(schema.TimeoutCreate),
		Delay:      5 * time.Second,
		MinTimeout: 3 * time.Second,
	}
	if _, err := stateConf.WaitForStateContext(ctx); err != nil {
		return diag.Errorf("Error waiting for vnpaycloud_backup_vault %s to become ready: %s", d.Id(), err)
	}

	return resourceBackupVaultRead(ctx, d, meta)
}

func resourceBackupVaultRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	cfg := meta.(*config.Config)

	resp := &dto.BackupVaultResponse{}
	if _, err := cfg.Client.Get(ctx, client.ApiPath.BackupVaultWithID(cfg.ProjectID, d.Id()), resp, nil); err != nil {
		return diag.FromErr(util.CheckNotFound(d, err, "Error retrieving vnpaycloud_backup_vault"))
	}

	v := resp.BackupVault
	d.Set("name", v.Name)
	d.Set("purpose", v.Purpose)
	d.Set("type", v.Type)
	d.Set("description", v.Description)
	d.Set("object_lock", v.ObjectLock)
	d.Set("lock_time_unit", v.LockTimeUnit)
	d.Set("lock_time_number", v.LockTimeNumber)
	d.Set("zone_id", v.ZoneID)
	d.Set("storage_location_id", v.StorageLocationID)
	d.Set("disk_used", v.DiskUsed)
	d.Set("quota", v.Quota)
	d.Set("committed_quota", committedQuotaForState(v.CommittedQuota))
	d.Set("status", v.Status)
	d.Set("created_at", v.CreatedAt)

	return nil
}

func resourceBackupVaultUpdate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	if diags := resourceBackupVaultUpdateInner(ctx, d, meta); diags.HasError() {
		return append(resourceBackupVaultRead(ctx, d, meta), diags...)
	} else {
		return diags
	}
}

func resourceBackupVaultUpdateInner(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	cfg := meta.(*config.Config)

	if d.HasChanges("name", "description") {
		updateOpts := dto.UpdateBackupVaultRequest{
			Name:        d.Get("name").(string),
			Description: d.Get("description").(string),
		}
		if _, err := cfg.Client.Put(ctx, client.ApiPath.BackupVaultWithID(cfg.ProjectID, d.Id()), updateOpts, nil, nil); err != nil {
			return diag.Errorf("Error updating vnpaycloud_backup_vault %s: %s", d.Id(), err)
		}
	}

	return resourceBackupVaultRead(ctx, d, meta)
}

func resourceBackupVaultDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	cfg := meta.(*config.Config)

	if _, err := cfg.Client.Delete(ctx, client.ApiPath.BackupVaultWithID(cfg.ProjectID, d.Id()), nil); err != nil {
		return diag.FromErr(util.CheckDeleted(d, err, "Error deleting vnpaycloud_backup_vault"))
	}

	stateConf := &retry.StateChangeConf{
		Pending:    []string{"deleting", "active", "unknown"},
		Target:     []string{"deleted"},
		Refresh:    backupVaultStateRefreshFunc(ctx, cfg.Client, cfg.ProjectID, d.Id()),
		Timeout:    d.Timeout(schema.TimeoutDelete),
		Delay:      5 * time.Second,
		MinTimeout: 3 * time.Second,
	}
	if _, err := stateConf.WaitForStateContext(ctx); err != nil {
		return diag.Errorf("Error waiting for vnpaycloud_backup_vault %s to delete: %s", d.Id(), err)
	}

	return nil
}
