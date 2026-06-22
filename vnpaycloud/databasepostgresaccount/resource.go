package databasepostgresaccount

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

func ResourceDatabasePostgresAccount() *schema.Resource {
	return &schema.Resource{
		CreateContext: resourceDatabasePostgresAccountCreate,
		ReadContext:   resourceDatabasePostgresAccountRead,
		UpdateContext: resourceDatabasePostgresAccountUpdate,
		DeleteContext: resourceDatabasePostgresAccountDelete,
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
					validation.StringMatch(
						regexp.MustCompile(`^[a-z_][a-z0-9_]*$`),
						"must start with a lowercase letter or underscore and contain only lowercase letters, digits and underscores",
					),
				),
			},
			"postgres_instance_id": {
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
			"grant": {
				Type:     schema.TypeSet,
				Optional: true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"db_name": {
							Type:     schema.TypeString,
							Required: true,
						},
						"db_schema": {
							Type:     schema.TypeString,
							Required: true,
						},
						"privilege": {
							Type:         schema.TypeString,
							Required:     true,
							ValidateFunc: validation.StringInSlice([]string{"readonly", "readwrite"}, false),
						},
					},
				},
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

func resourceDatabasePostgresAccountCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	cfg := meta.(*config.Config)

	createOpts := dto.CreatePostgresAccountRequest{
		Name:               d.Get("name").(string),
		PostgresInstanceID: d.Get("postgres_instance_id").(string),
		Password:           d.Get("password").(string),
	}

	tflog.Debug(ctx, "vnpaycloud_database_postgres_account create options", map[string]interface{}{
		"name":                 createOpts.Name,
		"postgres_instance_id": createOpts.PostgresInstanceID,
	})

	createResp := &dto.PostgresAccountResponse{}
	if _, err := cfg.Client.Post(ctx, client.ApiPath.DatabasePostgresAccounts(cfg.ProjectID), createOpts, createResp, nil); err != nil {
		return diag.Errorf("Error creating vnpaycloud_database_postgres_account: %s", err)
	}

	d.SetId(createResp.PostgresAccount.ID)

	stateConf := &retry.StateChangeConf{
		Pending:    []string{"creating", "unknown"},
		Target:     []string{"active"},
		Refresh:    postgresAccountStateRefreshFunc(ctx, cfg.Client, cfg.ProjectID, d.Id()),
		Timeout:    d.Timeout(schema.TimeoutCreate),
		Delay:      5 * time.Second,
		MinTimeout: 3 * time.Second,
	}
	if _, err := stateConf.WaitForStateContext(ctx); err != nil {
		return diag.Errorf("Error waiting for vnpaycloud_database_postgres_account %s to become ready: %s", d.Id(), err)
	}

	// The create API does not accept grants; apply each configured grant after the account is ready.
	for _, g := range d.Get("grant").(*schema.Set).List() {
		if err := grantPostgresAccountPrivilege(ctx, cfg, d.Id(), g.(map[string]interface{})); err != nil {
			return diag.FromErr(err)
		}
	}

	return resourceDatabasePostgresAccountRead(ctx, d, meta)
}

func resourceDatabasePostgresAccountRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	cfg := meta.(*config.Config)

	resp := &dto.PostgresAccountResponse{}
	if _, err := cfg.Client.Get(ctx, client.ApiPath.DatabasePostgresAccountWithID(cfg.ProjectID, d.Id()), resp, nil); err != nil {
		return diag.FromErr(util.CheckNotFound(d, err, "Error retrieving vnpaycloud_database_postgres_account"))
	}

	acc := resp.PostgresAccount
	d.Set("name", acc.Name)
	d.Set("postgres_instance_id", acc.PostgresInstanceID)
	d.Set("grant", flattenPostgresAccountGrants(acc.Grants))
	d.Set("status", acc.Status)
	d.Set("created_at", acc.CreatedAt)
	// password is a write-only input — the API never returns it, so the state value is preserved.

	return nil
}

func resourceDatabasePostgresAccountUpdate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	cfg := meta.(*config.Config)

	if !d.HasChanges("password", "grant") {
		return resourceDatabasePostgresAccountRead(ctx, d, meta)
	}

	if d.HasChange("password") {
		changeOpts := dto.ChangePasswordPostgresAccountRequest{NewPassword: d.Get("password").(string)}
		if _, err := cfg.Client.Post(ctx, client.ApiPath.DatabasePostgresAccountChangePassword(cfg.ProjectID, d.Id()), changeOpts, nil, nil); err != nil {
			// password is write-only (never returned by the API); restore the prior
			// value so a failed change isn't silently masked as already applied.
			old, _ := d.GetChange("password")
			d.Set("password", old)
			return diag.Errorf("Error changing password for vnpaycloud_database_postgres_account %s: %s", d.Id(), err)
		}
	}

	if d.HasChange("grant") {
		o, n := d.GetChange("grant")
		oldSet := o.(*schema.Set)
		newSet := n.(*schema.Set)

		for _, g := range oldSet.Difference(newSet).List() {
			if err := revokePostgresAccountPrivilege(ctx, cfg, d.Id(), g.(map[string]interface{})); err != nil {
				return diag.FromErr(err)
			}
		}
		for _, g := range newSet.Difference(oldSet).List() {
			if err := grantPostgresAccountPrivilege(ctx, cfg, d.Id(), g.(map[string]interface{})); err != nil {
				return diag.FromErr(err)
			}
		}
	}

	return resourceDatabasePostgresAccountRead(ctx, d, meta)
}

func resourceDatabasePostgresAccountDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	cfg := meta.(*config.Config)

	if _, err := cfg.Client.Delete(ctx, client.ApiPath.DatabasePostgresAccountWithID(cfg.ProjectID, d.Id()), nil); err != nil {
		return diag.FromErr(util.CheckDeleted(d, err, "Error deleting vnpaycloud_database_postgres_account"))
	}

	stateConf := &retry.StateChangeConf{
		Pending:    []string{"deleting", "active", "unknown"},
		Target:     []string{"deleted"},
		Refresh:    postgresAccountStateRefreshFunc(ctx, cfg.Client, cfg.ProjectID, d.Id()),
		Timeout:    d.Timeout(schema.TimeoutDelete),
		Delay:      5 * time.Second,
		MinTimeout: 3 * time.Second,
	}
	if _, err := stateConf.WaitForStateContext(ctx); err != nil {
		return diag.Errorf("Error waiting for vnpaycloud_database_postgres_account %s to delete: %s", d.Id(), err)
	}

	return nil
}
