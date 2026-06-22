package databaseredissentinelaccount

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

var redisAccountNameRegex = regexp.MustCompile(`^[a-z0-9]([a-z0-9.]*[a-z0-9])?$`)

func ResourceDatabaseRedisSentinelAccount() *schema.Resource {
	return &schema.Resource{
		CreateContext: resourceDatabaseRedisSentinelAccountCreate,
		ReadContext:   resourceDatabaseRedisSentinelAccountRead,
		UpdateContext: resourceDatabaseRedisSentinelAccountUpdate,
		DeleteContext: resourceDatabaseRedisSentinelAccountDelete,
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
				ForceNew: true,
				ValidateFunc: validation.All(
					validation.StringLenBetween(1, 63),
					validation.StringMatch(redisAccountNameRegex, "must contain only lowercase letters, digits and dots, and start/end with an alphanumeric character"),
				),
			},
			"redis_sentinel_instance_id": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},
			"password": {
				Type:         schema.TypeString,
				Required:     true,
				Sensitive:    true,
				ValidateFunc: validation.StringLenBetween(8, 128),
			},
			"privilege_template": {
				Type:         schema.TypeString,
				Required:     true,
				ValidateFunc: validation.StringInSlice([]string{"readonly", "readwrite"}, false),
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

func resourceDatabaseRedisSentinelAccountCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	cfg := meta.(*config.Config)

	createOpts := dto.CreateRedisSentinelAccountRequest{
		Name:                    d.Get("name").(string),
		RedisSentinelInstanceID: d.Get("redis_sentinel_instance_id").(string),
		Password:                d.Get("password").(string),
		PrivilegeTemplate:       d.Get("privilege_template").(string),
	}

	tflog.Debug(ctx, "vnpaycloud_database_redis_sentinel_account create options", map[string]interface{}{
		"name":                       createOpts.Name,
		"redis_sentinel_instance_id": createOpts.RedisSentinelInstanceID,
		"privilege_template":         createOpts.PrivilegeTemplate,
	})

	createResp := &dto.RedisSentinelAccountResponse{}
	if _, err := cfg.Client.Post(ctx, client.ApiPath.DatabaseRedisSentinelAccounts(cfg.ProjectID), createOpts, createResp, nil); err != nil {
		return diag.Errorf("Error creating vnpaycloud_database_redis_sentinel_account: %s", err)
	}

	d.SetId(createResp.RedisSentinelAccount.ID)

	stateConf := &retry.StateChangeConf{
		Pending:    []string{"creating", "unknown"},
		Target:     []string{"active"},
		Refresh:    redisSentinelAccountStateRefreshFunc(ctx, cfg.Client, cfg.ProjectID, d.Id()),
		Timeout:    d.Timeout(schema.TimeoutCreate),
		Delay:      5 * time.Second,
		MinTimeout: 3 * time.Second,
	}
	if _, err := stateConf.WaitForStateContext(ctx); err != nil {
		return diag.Errorf("Error waiting for vnpaycloud_database_redis_sentinel_account %s to become ready: %s", d.Id(), err)
	}

	return resourceDatabaseRedisSentinelAccountRead(ctx, d, meta)
}

func resourceDatabaseRedisSentinelAccountRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	cfg := meta.(*config.Config)

	resp := &dto.RedisSentinelAccountResponse{}
	if _, err := cfg.Client.Get(ctx, client.ApiPath.DatabaseRedisSentinelAccountWithID(cfg.ProjectID, d.Id()), resp, nil); err != nil {
		return diag.FromErr(util.CheckNotFound(d, err, "Error retrieving vnpaycloud_database_redis_sentinel_account"))
	}

	acc := resp.RedisSentinelAccount
	d.Set("name", acc.Name)
	d.Set("redis_sentinel_instance_id", acc.RedisSentinelInstanceID)
	d.Set("privilege_template", acc.PrivilegeTemplate)
	d.Set("status", acc.Status)
	d.Set("created_at", acc.CreatedAt)
	// password is a write-only input — the API never returns it, so the state value is preserved.

	return nil
}

func resourceDatabaseRedisSentinelAccountUpdate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	cfg := meta.(*config.Config)

	if !d.HasChanges("password", "privilege_template") {
		return resourceDatabaseRedisSentinelAccountRead(ctx, d, meta)
	}

	// Changing the password re-applies the privilege in the same call; a privilege-only
	// change uses grant-privilege.
	if d.HasChange("password") {
		changeOpts := dto.ChangePasswordRedisSentinelAccountRequest{
			NewPassword:       d.Get("password").(string),
			PrivilegeTemplate: d.Get("privilege_template").(string),
		}
		if _, err := cfg.Client.Post(ctx, client.ApiPath.DatabaseRedisSentinelAccountChangePassword(cfg.ProjectID, d.Id()), changeOpts, nil, nil); err != nil {
			// password is write-only (never returned by the API); restore the prior
			// value so a failed change isn't silently masked as already applied.
			old, _ := d.GetChange("password")
			d.Set("password", old)
			return diag.Errorf("Error changing password for vnpaycloud_database_redis_sentinel_account %s: %s", d.Id(), err)
		}
	} else if d.HasChange("privilege_template") {
		grantOpts := dto.GrantPrivilegeRedisSentinelAccountRequest{PrivilegeTemplate: d.Get("privilege_template").(string)}
		if _, err := cfg.Client.Post(ctx, client.ApiPath.DatabaseRedisSentinelAccountGrantPrivilege(cfg.ProjectID, d.Id()), grantOpts, nil, nil); err != nil {
			return diag.Errorf("Error updating privilege for vnpaycloud_database_redis_sentinel_account %s: %s", d.Id(), err)
		}
	}

	return resourceDatabaseRedisSentinelAccountRead(ctx, d, meta)
}

func resourceDatabaseRedisSentinelAccountDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	cfg := meta.(*config.Config)

	if _, err := cfg.Client.Delete(ctx, client.ApiPath.DatabaseRedisSentinelAccountWithID(cfg.ProjectID, d.Id()), nil); err != nil {
		return diag.FromErr(util.CheckDeleted(d, err, "Error deleting vnpaycloud_database_redis_sentinel_account"))
	}

	stateConf := &retry.StateChangeConf{
		Pending:    []string{"deleting", "active", "unknown"},
		Target:     []string{"deleted"},
		Refresh:    redisSentinelAccountStateRefreshFunc(ctx, cfg.Client, cfg.ProjectID, d.Id()),
		Timeout:    d.Timeout(schema.TimeoutDelete),
		Delay:      5 * time.Second,
		MinTimeout: 3 * time.Second,
	}
	if _, err := stateConf.WaitForStateContext(ctx); err != nil {
		return diag.Errorf("Error waiting for vnpaycloud_database_redis_sentinel_account %s to delete: %s", d.Id(), err)
	}

	return nil
}
