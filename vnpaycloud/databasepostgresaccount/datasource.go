package databasepostgresaccount

import (
	"context"
	"terraform-provider-vnpaycloud/vnpaycloud/config"
	"terraform-provider-vnpaycloud/vnpaycloud/dto"
	"terraform-provider-vnpaycloud/vnpaycloud/helper/client"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

func DataSourceDatabasePostgresAccount() *schema.Resource {
	return &schema.Resource{
		ReadContext: dataSourcePostgresAccountRead,
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
			"grant": {
				Type:     schema.TypeList,
				Computed: true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"db_name":   {Type: schema.TypeString, Computed: true},
						"db_schema": {Type: schema.TypeString, Computed: true},
						"privilege": {Type: schema.TypeString, Computed: true},
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

func dataSourcePostgresAccountRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	cfg := meta.(*config.Config)
	id := d.Get("id").(string)

	resp := &dto.PostgresAccountResponse{}
	if _, err := cfg.Client.Get(ctx, client.ApiPath.DatabasePostgresAccountWithID(cfg.ProjectID, id), resp, nil); err != nil {
		return diag.Errorf("Error retrieving vnpaycloud_database_postgres_account %s: %s", id, err)
	}

	acc := resp.PostgresAccount
	d.SetId(acc.ID)
	d.Set("name", acc.Name)
	d.Set("postgres_instance_id", acc.PostgresInstanceID)
	d.Set("grant", flattenPostgresAccountGrants(acc.Grants))
	d.Set("status", acc.Status)
	d.Set("created_at", acc.CreatedAt)

	return nil
}

func DataSourceDatabasePostgresAccounts() *schema.Resource {
	return &schema.Resource{
		ReadContext: dataSourcePostgresAccountsRead,
		Schema: map[string]*schema.Schema{
			"postgres_accounts": {
				Type:     schema.TypeList,
				Computed: true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"id":                   {Type: schema.TypeString, Computed: true},
						"name":                 {Type: schema.TypeString, Computed: true},
						"postgres_instance_id": {Type: schema.TypeString, Computed: true},
						"status":               {Type: schema.TypeString, Computed: true},
						"created_at":           {Type: schema.TypeString, Computed: true},
					},
				},
			},
		},
	}
}

func dataSourcePostgresAccountsRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	cfg := meta.(*config.Config)

	resp := &dto.ListPostgresAccountsResponse{}
	if _, err := cfg.Client.Get(ctx, client.ApiPath.DatabasePostgresAccounts(cfg.ProjectID), resp, nil); err != nil {
		return diag.Errorf("Error listing vnpaycloud_database_postgres_accounts: %s", err)
	}

	var accounts []map[string]interface{}
	for _, acc := range resp.PostgresAccounts {
		accounts = append(accounts, map[string]interface{}{
			"id":                   acc.ID,
			"name":                 acc.Name,
			"postgres_instance_id": acc.PostgresInstanceID,
			"status":               acc.Status,
			"created_at":           acc.CreatedAt,
		})
	}

	d.SetId("database_postgres_accounts")
	d.Set("postgres_accounts", accounts)

	return nil
}
