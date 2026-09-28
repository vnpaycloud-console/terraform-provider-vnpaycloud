package backupserverrestorepoint

import (
	"context"
	"fmt"
	"net/url"
	"strings"
	"terraform-provider-vnpaycloud/vnpaycloud/config"
	"terraform-provider-vnpaycloud/vnpaycloud/dto"
	"terraform-provider-vnpaycloud/vnpaycloud/helper/client"
	"terraform-provider-vnpaycloud/vnpaycloud/util"
	"time"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/retry"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

// latestBackupGuardMessage is how the backend refuses to delete a server backup's newest
// restore point while older ones remain.
const latestBackupGuardMessage = "unable to delete tier-1 latest backup"

var (
	notFoundGrace          = 60 * time.Second
	latestBackupGuardGrace = 30 * time.Second
)

func ResourceBackupServerRestorePoint() *schema.Resource {
	return &schema.Resource{
		CreateContext: resourceBackupServerRestorePointCreate,
		ReadContext:   resourceBackupServerRestorePointRead,
		DeleteContext: resourceBackupServerRestorePointDelete,
		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},
		Timeouts: &schema.ResourceTimeout{
			Create: schema.DefaultTimeout(20 * time.Minute),
			Delete: schema.DefaultTimeout(60 * time.Minute),
		},
		Schema: map[string]*schema.Schema{
			"backup_server_id": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},
			"name":                 {Type: schema.TypeString, Computed: true},
			"backup_policy_id":     {Type: schema.TypeString, Computed: true},
			"type":                 {Type: schema.TypeString, Computed: true},
			"purpose":              {Type: schema.TypeString, Computed: true},
			"zone_id":              {Type: schema.TypeString, Computed: true},
			"is_visible":           {Type: schema.TypeBool, Computed: true},
			"create_by_backup_now": {Type: schema.TypeBool, Computed: true},
			"backup_point":         {Type: schema.TypeString, Computed: true},
			"status":               {Type: schema.TypeString, Computed: true},
			"created_at":           {Type: schema.TypeString, Computed: true},
		},
	}
}

func resourceBackupServerRestorePointCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	cfg := meta.(*config.Config)
	backupServerID := d.Get("backup_server_id").(string)

	lockKey := "vnpaycloud_backup_server/" + backupServerID
	cfg.MutexKV.Lock(lockKey)
	defer cfg.MutexKV.Unlock(lockKey)

	createOpts := dto.CreateBackupServerRestorePointRequest{
		BackupServerID: backupServerID,
	}

	resp := &dto.BackupServerRestorePointResponse{}
	if _, err := cfg.Client.Post(ctx, client.ApiPath.BackupServerRestorePoints(cfg.ProjectID), createOpts, resp, nil); err != nil {
		return diag.Errorf("Error creating vnpaycloud_backup_server_restore_point: %s", err)
	}

	d.SetId(resp.RestorePoint.ID)

	stateConf := &retry.StateChangeConf{
		Pending:    []string{"initiating", "creating", "unknown"},
		Target:     []string{"active"},
		Refresh:    backupServerRestorePointStateRefreshFunc(ctx, cfg, backupServerID, d.Id()),
		Timeout:    d.Timeout(schema.TimeoutCreate),
		Delay:      10 * time.Second,
		MinTimeout: 5 * time.Second,
	}
	if _, err := stateConf.WaitForStateContext(ctx); err != nil {
		return diag.Errorf("Error waiting for vnpaycloud_backup_server_restore_point %s to become ready: %s", d.Id(), err)
	}

	return resourceBackupServerRestorePointRead(ctx, d, meta)
}

func resourceBackupServerRestorePointRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	cfg := meta.(*config.Config)

	rp, err := findBackupServerRestorePoint(ctx, cfg, d.Get("backup_server_id").(string), d.Id())
	if err != nil {
		return diag.FromErr(err)
	}
	if rp == nil {
		d.SetId("")
		return nil
	}

	setBackupServerRestorePointData(d, rp)
	return nil
}

func resourceBackupServerRestorePointDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	cfg := meta.(*config.Config)

	if err := deleteBackupServerRestorePoint(ctx, cfg, d.Get("backup_server_id").(string), d.Id(), d.Timeout(schema.TimeoutDelete)); err != nil {
		return diag.FromErr(util.CheckDeleted(d, err, "Error deleting vnpaycloud_backup_server_restore_point"))
	}

	var lastStatus string
	stateConf := &retry.StateChangeConf{
		Pending:    []string{"deleting", "active", "unknown"},
		Target:     []string{"deleted"},
		Refresh:    backupServerRestorePointDeleteStateRefreshFunc(ctx, cfg, d.Get("backup_server_id").(string), d.Id(), &lastStatus),
		Timeout:    d.Timeout(schema.TimeoutDelete),
		Delay:      10 * time.Second,
		MinTimeout: 10 * time.Second,
	}
	if _, err := stateConf.WaitForStateContext(ctx); err != nil {
		if lastStatus != "" {
			return diag.Errorf(
				"Error waiting for vnpaycloud_backup_server_restore_point %s to be deleted: the backend still reports "+
					"status %q. Deletion keeps running server-side; re-run destroy later, or raise the delete timeout. "+
					"Underlying error: %s",
				d.Id(), lastStatus, err)
		}
		return diag.Errorf("Error waiting for vnpaycloud_backup_server_restore_point %s to be deleted: %s", d.Id(), err)
	}

	return nil
}

func setBackupServerRestorePointData(d *schema.ResourceData, rp *dto.BackupServerRestorePoint) {
	d.Set("backup_server_id", rp.BackupServerID)
	d.Set("name", rp.Name)
	d.Set("backup_policy_id", rp.BackupPolicyID)
	d.Set("type", rp.Type)
	d.Set("purpose", rp.Purpose)
	d.Set("zone_id", rp.ZoneID)
	d.Set("is_visible", rp.IsVisible)
	d.Set("create_by_backup_now", rp.CreateByBackupNow)
	d.Set("backup_point", rp.BackupPoint)
	d.Set("status", rp.Status)
	d.Set("created_at", rp.CreatedAt)
}

// findBackupServerRestorePoint returns the restore point with the given id from the backup
// server's list, or nil if it no longer exists.
func findBackupServerRestorePoint(ctx context.Context, cfg *config.Config, backupServerID, id string) (*dto.BackupServerRestorePoint, error) {
	points, err := listBackupServerRestorePoints(ctx, cfg, backupServerID)
	if err != nil {
		return nil, err
	}
	for i := range points {
		if points[i].ID == id {
			return &points[i], nil
		}
	}
	return nil, nil
}

func listBackupServerRestorePoints(ctx context.Context, cfg *config.Config, backupServerID string) ([]dto.BackupServerRestorePoint, error) {
	path := client.ApiPath.BackupServerRestorePoints(cfg.ProjectID)
	if backupServerID != "" {
		path += "?" + url.Values{"backupServerId": []string{backupServerID}}.Encode()
	}

	listResp := &dto.ListBackupServerRestorePointsResponse{}
	if _, err := cfg.Client.Get(ctx, path, listResp, nil); err != nil {
		return nil, err
	}
	return listResp.RestorePoints, nil
}

// deleteBackupServerRestorePoint retries the backend's refusal to delete a server backup's
// newest restore point for as long as an older one is being deleted alongside it, which is
// what happens when a single destroy removes several restore points of the same server.
func deleteBackupServerRestorePoint(ctx context.Context, cfg *config.Config, backupServerID, id string, timeout time.Duration) error {
	path := client.ApiPath.BackupServerRestorePointWithID(cfg.ProjectID, id)
	start := time.Now()

	return retry.RetryContext(ctx, timeout, func() *retry.RetryError {
		_, err := cfg.Client.Delete(ctx, path, nil)
		if err == nil {
			return nil
		}
		if !strings.Contains(err.Error(), latestBackupGuardMessage) {
			return retry.NonRetryableError(err)
		}
		if time.Since(start) < latestBackupGuardGrace {
			return retry.RetryableError(err)
		}

		pending, listErr := otherRestorePointDeleting(ctx, cfg, backupServerID, id)
		if listErr != nil {
			return retry.NonRetryableError(listErr)
		}
		if pending {
			return retry.RetryableError(err)
		}
		return retry.NonRetryableError(fmt.Errorf("%w: the newest restore point of a server backup can only be deleted "+
			"once its older restore points are gone, so delete those first", err))
	})
}

func otherRestorePointDeleting(ctx context.Context, cfg *config.Config, backupServerID, id string) (bool, error) {
	points, err := listBackupServerRestorePoints(ctx, cfg, backupServerID)
	if err != nil {
		return false, err
	}
	for _, rp := range points {
		if rp.ID != id && rp.Status == "deleting" {
			return true, nil
		}
	}
	return false, nil
}

func backupServerRestorePointDeleteStateRefreshFunc(ctx context.Context, cfg *config.Config, backupServerID, id string, lastStatus *string) retry.StateRefreshFunc {
	return func() (interface{}, string, error) {
		rp, err := findBackupServerRestorePoint(ctx, cfg, backupServerID, id)
		if err != nil {
			return nil, "", err
		}
		if rp == nil {
			return &dto.BackupServerRestorePoint{ID: id}, "deleted", nil
		}
		// Recorded here because a timed-out wait hands back no object to report on.
		*lastStatus = rp.Status
		return rp, rp.Status, nil
	}
}

func backupServerRestorePointStateRefreshFunc(ctx context.Context, cfg *config.Config, backupServerID, id string) retry.StateRefreshFunc {
	start := time.Now()

	return func() (interface{}, string, error) {
		rp, err := findBackupServerRestorePoint(ctx, cfg, backupServerID, id)
		if err != nil {
			return nil, "", err
		}
		if rp == nil {
			// A freshly created restore point can take a moment to be listed, so
			// tolerate that briefly; beyond the grace period it is gone for good.
			if time.Since(start) < notFoundGrace {
				return nil, "", nil
			}
			return nil, "", fmt.Errorf("backup job failed on the backend: the restore point was rolled back. Retry, or contact support")
		}
		switch rp.Status {
		case "failed", "error", "deleted":
			return rp, rp.Status, fmt.Errorf("backup job failed on the backend (status %q): the restore point was rolled back. Retry, or contact support", rp.Status)
		}
		return rp, rp.Status, nil
	}
}
