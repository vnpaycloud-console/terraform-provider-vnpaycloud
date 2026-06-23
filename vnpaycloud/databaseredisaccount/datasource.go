package databaseredisaccount

import (
	"context"
	"terraform-provider-vnpaycloud/vnpaycloud/config"
	"terraform-provider-vnpaycloud/vnpaycloud/dto"
	"terraform-provider-vnpaycloud/vnpaycloud/helper/client"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

func DataSourceDatabaseRedisAccount() *schema.Resource {
	return &schema.Resource{
		ReadContext: dataSourceRedisAccountRead,
		Schema: map[string]*schema.Schema{
			"id": {
				Type:     schema.TypeString,
				Required: true,
			},
			"name": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"redis_instance_id": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"privilege_template": {
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

func dataSourceRedisAccountRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	cfg := meta.(*config.Config)
	id := d.Get("id").(string)

	resp := &dto.RedisAccountResponse{}
	if _, err := cfg.Client.Get(ctx, client.ApiPath.DatabaseRedisAccountWithID(cfg.ProjectID, id), resp, nil); err != nil {
		return diag.Errorf("Error retrieving vnpaycloud_database_redis_account %s: %s", id, err)
	}

	acc := resp.RedisAccount
	d.SetId(acc.ID)
	d.Set("name", acc.Name)
	d.Set("redis_instance_id", acc.RedisInstanceID)
	d.Set("privilege_template", acc.PrivilegeTemplate)
	d.Set("status", acc.Status)
	d.Set("created_at", acc.CreatedAt)

	return nil
}

func DataSourceDatabaseRedisAccounts() *schema.Resource {
	return &schema.Resource{
		ReadContext: dataSourceRedisAccountsRead,
		Schema: map[string]*schema.Schema{
			"redis_accounts": {
				Type:     schema.TypeList,
				Computed: true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"id":                 {Type: schema.TypeString, Computed: true},
						"name":               {Type: schema.TypeString, Computed: true},
						"redis_instance_id":  {Type: schema.TypeString, Computed: true},
						"privilege_template": {Type: schema.TypeString, Computed: true},
						"status":             {Type: schema.TypeString, Computed: true},
						"created_at":         {Type: schema.TypeString, Computed: true},
					},
				},
			},
		},
	}
}

func dataSourceRedisAccountsRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	cfg := meta.(*config.Config)

	resp := &dto.ListRedisAccountsResponse{}
	if _, err := cfg.Client.Get(ctx, client.ApiPath.DatabaseRedisAccounts(cfg.ProjectID), resp, nil); err != nil {
		return diag.Errorf("Error listing vnpaycloud_database_redis_accounts: %s", err)
	}

	var accounts []map[string]interface{}
	for _, acc := range resp.RedisAccounts {
		accounts = append(accounts, map[string]interface{}{
			"id":                 acc.ID,
			"name":               acc.Name,
			"redis_instance_id":  acc.RedisInstanceID,
			"privilege_template": acc.PrivilegeTemplate,
			"status":             acc.Status,
			"created_at":         acc.CreatedAt,
		})
	}

	d.SetId("database_redis_accounts")
	d.Set("redis_accounts", accounts)

	return nil
}
