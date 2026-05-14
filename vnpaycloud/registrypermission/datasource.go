package registrypermission

import (
	"context"
	"terraform-provider-vnpaycloud/vnpaycloud/config"
	"terraform-provider-vnpaycloud/vnpaycloud/dto"
	"terraform-provider-vnpaycloud/vnpaycloud/helper/client"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

// DataSourceRegistryPermissions returns the catalogue of (resource, action)
// pairs the registry accepts for robot account permissions. Use it to discover
// valid values for `vnpaycloud_registry_robot_account.permission.actions`.
func DataSourceRegistryPermissions() *schema.Resource {
	return &schema.Resource{
		ReadContext: dataSourceRegistryPermissionsRead,
		Schema: map[string]*schema.Schema{
			"permissions": {
				Type:     schema.TypeList,
				Computed: true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"resource": {
							Type:     schema.TypeString,
							Computed: true,
						},
						"action": {
							Type:     schema.TypeString,
							Computed: true,
						},
						"key": {
							Type:        schema.TypeString,
							Computed:    true,
							Description: "Convenience field '<resource>:<action>' (the exact value to use in robot account permissions.actions).",
						},
					},
				},
			},
		},
	}
}

func dataSourceRegistryPermissionsRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	cfg := meta.(*config.Config)

	resp := &dto.ListRegistryPermissionsResponse{}
	_, err := cfg.Client.Get(ctx, client.ApiPath.RegistryPermissions(cfg.ProjectID), resp, nil)
	if err != nil {
		return diag.Errorf("Error listing vnpaycloud_registry_permissions: %s", err)
	}

	permissions := make([]map[string]interface{}, 0, len(resp.Permissions))
	for _, p := range resp.Permissions {
		permissions = append(permissions, map[string]interface{}{
			"resource": p.Resource,
			"action":   p.Action,
			"key":      p.Resource + ":" + p.Action,
		})
	}

	d.SetId("registry-permissions")
	d.Set("permissions", permissions)

	return nil
}
