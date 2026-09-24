package backupserver

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

var (
	idRegex = regexp.MustCompile(`^[a-zA-Z0-9-_.]*$`)
)

func ResourceBackupServer() *schema.Resource {
	return &schema.Resource{
		CreateContext: resourceBackupServerCreate,
		ReadContext:   resourceBackupServerRead,
		UpdateContext: resourceBackupServerUpdate,
		DeleteContext: resourceBackupServerDelete,
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
				Computed: true,
			},
			"server_id": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
				ValidateFunc: validation.All(
					validation.StringLenBetween(3, 100),
					validation.StringMatch(idRegex, "must contain only letters, digits, hyphens, underscores and dots"),
				),
			},
			"volume_ids": {
				Type:     schema.TypeList,
				Required: true,
				MinItems: 1,
				Elem: &schema.Schema{
					Type: schema.TypeString,
					ValidateFunc: validation.All(
						validation.StringLenBetween(3, 100),
						validation.StringMatch(idRegex, "must contain only letters, digits, hyphens, underscores and dots"),
					),
				},
			},
			"backup_policy_id": {
				Type:     schema.TypeString,
				Required: true,
				ValidateFunc: validation.All(
					validation.StringLenBetween(3, 100),
					validation.StringMatch(idRegex, "must contain only letters, digits, hyphens, underscores and dots"),
				),
			},

			"purpose":        {Type: schema.TypeString, Computed: true},
			"zone_id":        {Type: schema.TypeString, Computed: true},
			"is_compliant":   {Type: schema.TypeString, Computed: true},
			"compliance_msg": {Type: schema.TypeString, Computed: true},
			"status":         {Type: schema.TypeString, Computed: true},
			"created_at":     {Type: schema.TypeString, Computed: true},
		},
	}
}

func resourceBackupServerCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	cfg := meta.(*config.Config)

	createOpts := dto.CreateBackupServerRequest{
		ServerID:       d.Get("server_id").(string),
		VolumeIDs:      expandStringList(d.Get("volume_ids")),
		BackupPolicyID: d.Get("backup_policy_id").(string),
	}

	tflog.Debug(ctx, "vnpaycloud_backup_server create options", map[string]interface{}{
		"server_id":        createOpts.ServerID,
		"backup_policy_id": createOpts.BackupPolicyID,
	})

	createResp := &dto.BackupServerResponse{}
	if _, err := cfg.Client.Post(ctx, client.ApiPath.BackupServers(cfg.ProjectID), createOpts, createResp, nil); err != nil {
		return diag.Errorf("Error creating vnpaycloud_backup_server: %s", err)
	}

	d.SetId(createResp.BackupServer.ID)

	return resourceBackupServerRead(ctx, d, meta)
}

func resourceBackupServerRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	cfg := meta.(*config.Config)

	resp := &dto.BackupServerResponse{}
	if _, err := cfg.Client.Get(ctx, client.ApiPath.BackupServerWithID(cfg.ProjectID, d.Id()), resp, nil); err != nil {
		return diag.FromErr(util.CheckNotFound(d, err, "Error retrieving vnpaycloud_backup_server"))
	}

	setBackupServerData(d, &resp.BackupServer)

	return nil
}

func resourceBackupServerUpdate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	cfg := meta.(*config.Config)

	updateOpts := dto.UpdateBackupServerRequest{
		VolumeIDs:      expandStringList(d.Get("volume_ids")),
		BackupPolicyID: d.Get("backup_policy_id").(string),
	}

	if _, err := cfg.Client.Put(ctx, client.ApiPath.BackupServerWithID(cfg.ProjectID, d.Id()), updateOpts, nil, nil); err != nil {
		return diag.Errorf("Error updating vnpaycloud_backup_server %s: %s", d.Id(), err)
	}

	return resourceBackupServerRead(ctx, d, meta)
}

func resourceBackupServerDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	cfg := meta.(*config.Config)

	if _, err := cfg.Client.Delete(ctx, client.ApiPath.BackupServerWithID(cfg.ProjectID, d.Id()), nil); err != nil {
		return diag.FromErr(util.CheckDeleted(d, err, "Error deleting vnpaycloud_backup_server"))
	}

	return nil
}
