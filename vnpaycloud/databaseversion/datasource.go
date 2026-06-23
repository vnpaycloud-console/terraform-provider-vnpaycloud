package databaseversion

import (
	"context"
	"terraform-provider-vnpaycloud/vnpaycloud/config"
	"terraform-provider-vnpaycloud/vnpaycloud/dto"
	"terraform-provider-vnpaycloud/vnpaycloud/helper/client"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

func DataSourceDatabasePostgresVersions() *schema.Resource {
	return &schema.Resource{
		ReadContext: dataSourcePostgresVersionsRead,
		Schema: map[string]*schema.Schema{
			"versions": {
				Type:     schema.TypeList,
				Computed: true,
				Elem:     &schema.Schema{Type: schema.TypeString},
			},
		},
	}
}

func dataSourcePostgresVersionsRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	cfg := meta.(*config.Config)

	resp := &dto.ListPostgresVersionsResponse{}
	if _, err := cfg.Client.Get(ctx, client.ApiPath.DatabasePostgresVersions(cfg.ProjectID), resp, nil); err != nil {
		return diag.Errorf("Error listing vnpaycloud_database_postgres_versions: %s", err)
	}

	d.SetId("database_postgres_versions")
	d.Set("versions", resp.Versions)

	return nil
}

func DataSourceDatabaseRedisVersions() *schema.Resource {
	return &schema.Resource{
		ReadContext: dataSourceRedisVersionsRead,
		Schema: map[string]*schema.Schema{
			"versions": {
				Type:     schema.TypeList,
				Computed: true,
				Elem:     &schema.Schema{Type: schema.TypeString},
			},
		},
	}
}

func dataSourceRedisVersionsRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	cfg := meta.(*config.Config)

	resp := &dto.ListRedisVersionsResponse{}
	if _, err := cfg.Client.Get(ctx, client.ApiPath.DatabaseRedisVersions(cfg.ProjectID), resp, nil); err != nil {
		return diag.Errorf("Error listing vnpaycloud_database_redis_versions: %s", err)
	}

	d.SetId("database_redis_versions")
	d.Set("versions", resp.Versions)

	return nil
}
