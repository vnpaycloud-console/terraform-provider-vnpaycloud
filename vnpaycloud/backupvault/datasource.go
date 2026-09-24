package backupvault

import (
	"context"
	"net/url"
	"terraform-provider-vnpaycloud/vnpaycloud/config"
	"terraform-provider-vnpaycloud/vnpaycloud/dto"
	"terraform-provider-vnpaycloud/vnpaycloud/helper/client"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
)

func DataSourceBackupVault() *schema.Resource {
	return &schema.Resource{
		ReadContext: dataSourceBackupVaultRead,
		Schema: map[string]*schema.Schema{
			"id": {
				Type:         schema.TypeString,
				Optional:     true,
				Computed:     true,
				AtLeastOneOf: []string{"id", "name"},
			},
			"name": {
				Type:         schema.TypeString,
				Optional:     true,
				Computed:     true,
				AtLeastOneOf: []string{"id", "name"},
			},
			"purpose":             {Type: schema.TypeString, Computed: true},
			"type":                {Type: schema.TypeString, Computed: true},
			"description":         {Type: schema.TypeString, Computed: true},
			"object_lock":         {Type: schema.TypeBool, Computed: true},
			"lock_time_unit":      {Type: schema.TypeString, Computed: true},
			"lock_time_number":    {Type: schema.TypeInt, Computed: true},
			"zone_id":             {Type: schema.TypeString, Computed: true},
			"storage_location_id": {Type: schema.TypeString, Computed: true},
			"disk_used":           {Type: schema.TypeInt, Computed: true},
			"quota":               {Type: schema.TypeInt, Computed: true},
			"committed_quota":     {Type: schema.TypeInt, Computed: true},
			"status":              {Type: schema.TypeString, Computed: true},
			"created_at":          {Type: schema.TypeString, Computed: true},
		},
	}
}

func dataSourceBackupVaultRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	cfg := meta.(*config.Config)

	if id, ok := d.GetOk("id"); ok && id.(string) != "" {
		resp := &dto.BackupVaultResponse{}
		if _, err := cfg.Client.Get(ctx, client.ApiPath.BackupVaultWithID(cfg.ProjectID, id.(string)), resp, nil); err != nil {
			return diag.Errorf("Error fetching vnpaycloud_backup_vault %s: %s", id, err)
		}
		if v, ok := d.GetOk("name"); ok && resp.BackupVault.Name != v.(string) {
			return diag.Errorf("vnpaycloud_backup_vault %q does not match name %q", id.(string), v.(string))
		}
		setBackupVaultData(d, &resp.BackupVault)
		return nil
	}

	name := d.Get("name").(string)
	listResp := &dto.ListBackupVaultsResponse{}
	path := client.ApiPath.BackupVaults(cfg.ProjectID) + "?name=" + url.QueryEscape(name)
	if _, err := cfg.Client.Get(ctx, path, listResp, nil); err != nil {
		return diag.Errorf("Error listing vnpaycloud_backup_vault: %s", err)
	}

	var matched []dto.BackupVault
	for _, v := range listResp.BackupVaults {
		if v.Name == name {
			matched = append(matched, v)
		}
	}

	if len(matched) == 0 {
		return diag.Errorf("No vnpaycloud_backup_vault found with name %q", name)
	}
	if len(matched) > 1 {
		return diag.Errorf("Found %d vnpaycloud_backup_vault matching name %q; use id to select one", len(matched), name)
	}

	setBackupVaultData(d, &matched[0])
	return nil
}

func setBackupVaultData(d *schema.ResourceData, v *dto.BackupVault) {
	d.SetId(v.ID)
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
}

func DataSourceBackupVaults() *schema.Resource {
	return &schema.Resource{
		ReadContext: dataSourceBackupVaultsRead,
		Schema: map[string]*schema.Schema{
			"name": {
				Type:     schema.TypeString,
				Optional: true,
			},
			"purpose": {
				Type:         schema.TypeString,
				Optional:     true,
				ValidateFunc: validation.StringInSlice([]string{"backup_server", "workload_cluster"}, false),
			},
			"type": {
				Type:         schema.TypeString,
				Optional:     true,
				ValidateFunc: validation.StringInSlice([]string{"ceph", "s3"}, false),
			},
			"backup_vaults": {
				Type:     schema.TypeList,
				Computed: true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"id":                  {Type: schema.TypeString, Computed: true},
						"name":                {Type: schema.TypeString, Computed: true},
						"purpose":             {Type: schema.TypeString, Computed: true},
						"type":                {Type: schema.TypeString, Computed: true},
						"description":         {Type: schema.TypeString, Computed: true},
						"object_lock":         {Type: schema.TypeBool, Computed: true},
						"lock_time_unit":      {Type: schema.TypeString, Computed: true},
						"lock_time_number":    {Type: schema.TypeInt, Computed: true},
						"zone_id":             {Type: schema.TypeString, Computed: true},
						"storage_location_id": {Type: schema.TypeString, Computed: true},
						"disk_used":           {Type: schema.TypeInt, Computed: true},
						"quota":               {Type: schema.TypeInt, Computed: true},
						"committed_quota":     {Type: schema.TypeInt, Computed: true},
						"status":              {Type: schema.TypeString, Computed: true},
						"created_at":          {Type: schema.TypeString, Computed: true},
					},
				},
			},
		},
	}
}

func dataSourceBackupVaultsRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	cfg := meta.(*config.Config)

	path := client.ApiPath.BackupVaults(cfg.ProjectID)
	q := url.Values{}
	if v, ok := d.GetOk("name"); ok {
		q.Set("name", v.(string))
	}
	if v, ok := d.GetOk("purpose"); ok {
		q.Set("purpose", v.(string))
	}
	if v, ok := d.GetOk("type"); ok {
		q.Set("type", v.(string))
	}
	if len(q) > 0 {
		path += "?" + q.Encode()
	}

	listResp := &dto.ListBackupVaultsResponse{}
	if _, err := cfg.Client.Get(ctx, path, listResp, nil); err != nil {
		return diag.Errorf("Unable to query vnpaycloud_backup_vaults: %s", err)
	}

	vaults := make([]map[string]interface{}, 0, len(listResp.BackupVaults))
	for _, v := range listResp.BackupVaults {
		vaults = append(vaults, map[string]interface{}{
			"id":                  v.ID,
			"name":                v.Name,
			"purpose":             v.Purpose,
			"type":                v.Type,
			"description":         v.Description,
			"object_lock":         v.ObjectLock,
			"lock_time_unit":      v.LockTimeUnit,
			"lock_time_number":    v.LockTimeNumber,
			"zone_id":             v.ZoneID,
			"storage_location_id": v.StorageLocationID,
			"disk_used":           v.DiskUsed,
			"quota":               v.Quota,
			"committed_quota":     committedQuotaForState(v.CommittedQuota),
			"status":              v.Status,
			"created_at":          v.CreatedAt,
		})
	}

	d.SetId(cfg.ProjectID)
	d.Set("backup_vaults", vaults)

	return nil
}
