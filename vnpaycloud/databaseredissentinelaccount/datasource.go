package databaseredissentinelaccount

import (
	"context"
	"terraform-provider-vnpaycloud/vnpaycloud/config"
	"terraform-provider-vnpaycloud/vnpaycloud/dto"
	"terraform-provider-vnpaycloud/vnpaycloud/helper/client"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

func DataSourceDatabaseRedisSentinelAccount() *schema.Resource {
	return &schema.Resource{
		ReadContext: dataSourceRedisSentinelAccountRead,
		Schema: map[string]*schema.Schema{
			"id": {
				Type:     schema.TypeString,
				Required: true,
			},
			"name": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"redis_sentinel_instance_id": {
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

func dataSourceRedisSentinelAccountRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	cfg := meta.(*config.Config)
	id := d.Get("id").(string)

	resp := &dto.RedisSentinelAccountResponse{}
	if _, err := cfg.Client.Get(ctx, client.ApiPath.DatabaseRedisSentinelAccountWithID(cfg.ProjectID, id), resp, nil); err != nil {
		return diag.Errorf("Error retrieving vnpaycloud_database_redis_sentinel_account %s: %s", id, err)
	}

	acc := resp.RedisSentinelAccount
	d.SetId(acc.ID)
	d.Set("name", acc.Name)
	d.Set("redis_sentinel_instance_id", acc.RedisSentinelInstanceID)
	d.Set("privilege_template", acc.PrivilegeTemplate)
	d.Set("status", acc.Status)
	d.Set("created_at", acc.CreatedAt)

	return nil
}

func DataSourceDatabaseRedisSentinelAccounts() *schema.Resource {
	return &schema.Resource{
		ReadContext: dataSourceRedisSentinelAccountsRead,
		Schema: map[string]*schema.Schema{
			"redis_sentinel_accounts": {
				Type:     schema.TypeList,
				Computed: true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"id":                         {Type: schema.TypeString, Computed: true},
						"name":                       {Type: schema.TypeString, Computed: true},
						"redis_sentinel_instance_id": {Type: schema.TypeString, Computed: true},
						"privilege_template":         {Type: schema.TypeString, Computed: true},
						"status":                     {Type: schema.TypeString, Computed: true},
						"created_at":                 {Type: schema.TypeString, Computed: true},
					},
				},
			},
		},
	}
}

func dataSourceRedisSentinelAccountsRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	cfg := meta.(*config.Config)

	resp := &dto.ListRedisSentinelAccountsResponse{}
	if _, err := cfg.Client.Get(ctx, client.ApiPath.DatabaseRedisSentinelAccounts(cfg.ProjectID), resp, nil); err != nil {
		return diag.Errorf("Error listing vnpaycloud_database_redis_sentinel_accounts: %s", err)
	}

	var accounts []map[string]interface{}
	for _, acc := range resp.RedisSentinelAccounts {
		accounts = append(accounts, map[string]interface{}{
			"id":                         acc.ID,
			"name":                       acc.Name,
			"redis_sentinel_instance_id": acc.RedisSentinelInstanceID,
			"privilege_template":         acc.PrivilegeTemplate,
			"status":                     acc.Status,
			"created_at":                 acc.CreatedAt,
		})
	}

	d.SetId("database_redis_sentinel_accounts")
	d.Set("redis_sentinel_accounts", accounts)

	return nil
}
