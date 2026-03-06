package robotaccount

import (
	"context"
	"terraform-provider-vnpaycloud/vnpaycloud/config"
	"terraform-provider-vnpaycloud/vnpaycloud/dto"
	"terraform-provider-vnpaycloud/vnpaycloud/helper/client"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

func DataSourceRobotAccount() *schema.Resource {
	return &schema.Resource{
		ReadContext: dataSourceRobotAccountRead,
		Schema: map[string]*schema.Schema{
			"id": {
				Type:     schema.TypeString,
				Required: true,
			},
			"registry_id": {
				Type:     schema.TypeString,
				Required: true,
			},
			"name": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"permissions": {
				Type:     schema.TypeList,
				Computed: true,
				Elem:     &schema.Schema{Type: schema.TypeString},
			},
			"expires_at": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"enabled": {
				Type:     schema.TypeBool,
				Computed: true,
			},
			"created_at": {
				Type:     schema.TypeString,
				Computed: true,
			},
		},
	}
}

func dataSourceRobotAccountRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	cfg := meta.(*config.Config)

	id := d.Get("id").(string)
	registryID := d.Get("registry_id").(string)

	resp := &dto.RobotAccountResponse{}
	_, err := cfg.Client.Get(ctx, client.ApiPath.RobotAccountWithID(cfg.ProjectID, registryID, id), resp, nil)
	if err != nil {
		return diag.Errorf("Error retrieving vnpaycloud_registry_robot_account %s: %s", id, err)
	}

	d.SetId(resp.RobotAccount.ID)
	d.Set("registry_id", resp.RobotAccount.RegistryID)
	d.Set("name", resp.RobotAccount.Name)
	d.Set("permissions", resp.RobotAccount.Permissions)
	d.Set("expires_at", resp.RobotAccount.ExpiresAt)
	d.Set("enabled", resp.RobotAccount.Enabled)
	d.Set("created_at", resp.RobotAccount.CreatedAt)

	return nil
}
