package user

import (
	"context"
	"fmt"
	"net/url"

	"terraform-provider-vnpaycloud/vnpaycloud/config"
	"terraform-provider-vnpaycloud/vnpaycloud/dto"
	"terraform-provider-vnpaycloud/vnpaycloud/helper/client"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

// DataSourceUsers lists the org users. Each user's id is the portal user id, usable as
// user_id in a vnpaycloud_kubernetes_rbac resource.
func DataSourceUsers() *schema.Resource {
	return &schema.Resource{
		ReadContext: dataSourceUsersRead,
		Schema: map[string]*schema.Schema{
			"search": {
				Type:     schema.TypeString,
				Optional: true,
			},
			"users": {
				Type:     schema.TypeList,
				Computed: true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"id":           {Type: schema.TypeString, Computed: true},
						"email":        {Type: schema.TypeString, Computed: true},
						"display_name": {Type: schema.TypeString, Computed: true},
						"username":     {Type: schema.TypeString, Computed: true},
					},
				},
			},
		},
	}
}

func dataSourceUsersRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	cfg := meta.(*config.Config)
	search := d.Get("search").(string)

	path := client.ApiPath.Users(cfg.ProjectID)
	if search != "" {
		path = fmt.Sprintf("%s?search=%s", path, url.QueryEscape(search))
	}

	listResp := &dto.ListUsersResponse{}
	_, err := cfg.Client.Get(ctx, path, listResp, nil)
	if err != nil {
		return diag.Errorf("Error listing vnpaycloud_users: %s", err)
	}

	var users []map[string]interface{}
	for _, u := range listResp.Users {
		users = append(users, map[string]interface{}{
			"id":           u.ID,
			"email":        u.Email,
			"display_name": u.DisplayName,
			"username":     u.Username,
		})
	}

	d.SetId(fmt.Sprintf("users-%s-%s", cfg.ProjectID, search))
	d.Set("users", users)

	return nil
}
