package databasepostgresdatabase

import (
	"context"
	"terraform-provider-vnpaycloud/vnpaycloud/config"
	"terraform-provider-vnpaycloud/vnpaycloud/dto"
	"terraform-provider-vnpaycloud/vnpaycloud/helper/client"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

func DataSourceDatabasePostgresDatabase() *schema.Resource {
	return &schema.Resource{
		ReadContext: dataSourcePostgresDatabaseRead,
		Schema: map[string]*schema.Schema{
			"id": {
				Type:     schema.TypeString,
				Required: true,
			},
			"name": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"postgres_instance_id": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"owner": {
				Type:     schema.TypeString,
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

func dataSourcePostgresDatabaseRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	cfg := meta.(*config.Config)
	id := d.Get("id").(string)

	resp := &dto.PostgresDatabaseResponse{}
	if _, err := cfg.Client.Get(ctx, client.ApiPath.DatabasePostgresDatabaseWithID(cfg.ProjectID, id), resp, nil); err != nil {
		return diag.Errorf("Error retrieving vnpaycloud_database_postgres_database %s: %s", id, err)
	}

	db := resp.PostgresDatabase
	d.SetId(db.ID)
	d.Set("name", db.Name)
	d.Set("postgres_instance_id", db.PostgresInstanceID)
	d.Set("owner", db.Owner)
	d.Set("status", db.Status)
	d.Set("created_at", db.CreatedAt)

	return nil
}

func DataSourceDatabasePostgresDatabases() *schema.Resource {
	return &schema.Resource{
		ReadContext: dataSourcePostgresDatabasesRead,
		Schema: map[string]*schema.Schema{
			"postgres_databases": {
				Type:     schema.TypeList,
				Computed: true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"id":                   {Type: schema.TypeString, Computed: true},
						"name":                 {Type: schema.TypeString, Computed: true},
						"postgres_instance_id": {Type: schema.TypeString, Computed: true},
						"owner":                {Type: schema.TypeString, Computed: true},
						"status":               {Type: schema.TypeString, Computed: true},
						"created_at":           {Type: schema.TypeString, Computed: true},
					},
				},
			},
		},
	}
}

func dataSourcePostgresDatabasesRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	cfg := meta.(*config.Config)

	resp := &dto.ListPostgresDatabasesResponse{}
	if _, err := cfg.Client.Get(ctx, client.ApiPath.DatabasePostgresDatabases(cfg.ProjectID), resp, nil); err != nil {
		return diag.Errorf("Error listing vnpaycloud_database_postgres_databases: %s", err)
	}

	var databases []map[string]interface{}
	for _, db := range resp.PostgresDatabases {
		databases = append(databases, map[string]interface{}{
			"id":                   db.ID,
			"name":                 db.Name,
			"postgres_instance_id": db.PostgresInstanceID,
			"owner":                db.Owner,
			"status":               db.Status,
			"created_at":           db.CreatedAt,
		})
	}

	d.SetId("database_postgres_databases")
	d.Set("postgres_databases", databases)

	return nil
}
