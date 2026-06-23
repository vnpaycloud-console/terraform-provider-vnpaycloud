package databasepostgresdatabase

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

var pgIdentifierRegex = regexp.MustCompile(`^[a-z_][a-z0-9_]*$`)

func ResourceDatabasePostgresDatabase() *schema.Resource {
	return &schema.Resource{
		CreateContext: resourceDatabasePostgresDatabaseCreate,
		ReadContext:   resourceDatabasePostgresDatabaseRead,
		UpdateContext: resourceDatabasePostgresDatabaseUpdate,
		DeleteContext: resourceDatabasePostgresDatabaseDelete,
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
					validation.StringMatch(pgIdentifierRegex, "must be a valid PostgreSQL identifier (lowercase letters, digits, underscores; not starting with a digit)"),
				),
			},
			"postgres_instance_id": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},
			"owner": {
				Type:     schema.TypeString,
				Required: true,
				ValidateFunc: validation.All(
					validation.StringLenBetween(1, 63),
					validation.StringMatch(pgIdentifierRegex, "must be a valid PostgreSQL role name (lowercase letters, digits, underscores; not starting with a digit)"),
				),
			},
			"force_delete": {
				Type:        schema.TypeBool,
				Optional:    true,
				Default:     false,
				Description: "Terminate active connections and force-drop the database on delete.",
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

func resourceDatabasePostgresDatabaseCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	cfg := meta.(*config.Config)

	createOpts := dto.CreatePostgresDatabaseRequest{
		Name:               d.Get("name").(string),
		PostgresInstanceID: d.Get("postgres_instance_id").(string),
		Owner:              d.Get("owner").(string),
	}

	tflog.Debug(ctx, "vnpaycloud_database_postgres_database create options", map[string]interface{}{"create_opts": createOpts})

	createResp := &dto.PostgresDatabaseResponse{}
	if _, err := cfg.Client.Post(ctx, client.ApiPath.DatabasePostgresDatabases(cfg.ProjectID), createOpts, createResp, nil); err != nil {
		return diag.Errorf("Error creating vnpaycloud_database_postgres_database: %s", err)
	}

	d.SetId(createResp.PostgresDatabase.ID)

	stateConf := &retry.StateChangeConf{
		Pending:    []string{"creating", "unknown"},
		Target:     []string{"active"},
		Refresh:    postgresDatabaseStateRefreshFunc(ctx, cfg.Client, cfg.ProjectID, d.Id()),
		Timeout:    d.Timeout(schema.TimeoutCreate),
		Delay:      5 * time.Second,
		MinTimeout: 3 * time.Second,
	}
	if _, err := stateConf.WaitForStateContext(ctx); err != nil {
		return diag.Errorf("Error waiting for vnpaycloud_database_postgres_database %s to become ready: %s", d.Id(), err)
	}

	return resourceDatabasePostgresDatabaseRead(ctx, d, meta)
}

func resourceDatabasePostgresDatabaseRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	cfg := meta.(*config.Config)

	resp := &dto.PostgresDatabaseResponse{}
	if _, err := cfg.Client.Get(ctx, client.ApiPath.DatabasePostgresDatabaseWithID(cfg.ProjectID, d.Id()), resp, nil); err != nil {
		return diag.FromErr(util.CheckNotFound(d, err, "Error retrieving vnpaycloud_database_postgres_database"))
	}

	db := resp.PostgresDatabase
	d.Set("name", db.Name)
	d.Set("postgres_instance_id", db.PostgresInstanceID)
	d.Set("owner", db.Owner)
	d.Set("status", db.Status)
	d.Set("created_at", db.CreatedAt)

	return nil
}

func resourceDatabasePostgresDatabaseUpdate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	cfg := meta.(*config.Config)

	if d.HasChange("owner") {
		changeOpts := dto.ChangeOwnershipPostgresDatabaseRequest{NewOwner: d.Get("owner").(string)}
		if _, err := cfg.Client.Post(ctx, client.ApiPath.DatabasePostgresDatabaseChangeOwnership(cfg.ProjectID, d.Id()), changeOpts, nil, nil); err != nil {
			return diag.Errorf("Error changing ownership of vnpaycloud_database_postgres_database %s: %s", d.Id(), err)
		}
	}

	return resourceDatabasePostgresDatabaseRead(ctx, d, meta)
}

func resourceDatabasePostgresDatabaseDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	cfg := meta.(*config.Config)

	path := client.ApiPath.DatabasePostgresDatabaseWithID(cfg.ProjectID, d.Id())
	if d.Get("force_delete").(bool) {
		path += "?isForceDelete=true"
	}

	if _, err := cfg.Client.Delete(ctx, path, nil); err != nil {
		return diag.FromErr(util.CheckDeleted(d, err, "Error deleting vnpaycloud_database_postgres_database"))
	}

	stateConf := &retry.StateChangeConf{
		Pending:    []string{"deleting", "active", "unknown"},
		Target:     []string{"deleted"},
		Refresh:    postgresDatabaseStateRefreshFunc(ctx, cfg.Client, cfg.ProjectID, d.Id()),
		Timeout:    d.Timeout(schema.TimeoutDelete),
		Delay:      5 * time.Second,
		MinTimeout: 3 * time.Second,
	}
	if _, err := stateConf.WaitForStateContext(ctx); err != nil {
		return diag.Errorf("Error waiting for vnpaycloud_database_postgres_database %s to delete: %s", d.Id(), err)
	}

	return nil
}
