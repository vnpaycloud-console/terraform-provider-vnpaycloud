package servergroup

import (
	"context"
	"fmt"
	"terraform-provider-vnpaycloud/vnpaycloud/config"
	"terraform-provider-vnpaycloud/vnpaycloud/dto"
	"terraform-provider-vnpaycloud/vnpaycloud/helper/client"

	"github.com/hashicorp/terraform-plugin-log/tflog"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

func DataSourceServerGroup() *schema.Resource {
	return &schema.Resource{
		ReadContext: dataSourceServerGroupRead,
		Schema: map[string]*schema.Schema{
			"id": {
				Type:     schema.TypeString,
				Optional: true,
				Computed: true,
			},
			"name": {
				Type:     schema.TypeString,
				Optional: true,
				Computed: true,
			},
			"policy": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"member_ids": {
				Type:     schema.TypeList,
				Computed: true,
				Elem:     &schema.Schema{Type: schema.TypeString},
			},
			"created_at": {
				Type:     schema.TypeString,
				Computed: true,
			},
		},
	}
}

func dataSourceServerGroupRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	cfg := meta.(*config.Config)

	if id, ok := d.GetOk("id"); ok && id.(string) != "" {
		sgResp := &dto.ServerGroupResponse{}
		_, err := cfg.Client.Get(ctx, client.ApiPath.ServerGroupWithID(cfg.ProjectID, id.(string)), sgResp, nil)
		if err != nil {
			return diag.Errorf("Error retrieving vnpaycloud_server_group %s: %s", id, err)
		}
		setServerGroupDataSourceAttributes(d, sgResp.ServerGroup)
		return nil
	}

	listResp := &dto.ListServerGroupsResponse{}
	_, err := cfg.Client.Get(ctx, client.ApiPath.ServerGroups(cfg.ProjectID), listResp, nil)
	if err != nil {
		return diag.Errorf("Unable to query vnpaycloud_server_group: %s", err)
	}

	name := d.Get("name").(string)
	var matched []dto.ServerGroup
	for _, sg := range listResp.ServerGroups {
		if name != "" && sg.Name != name {
			continue
		}
		matched = append(matched, sg)
	}

	if len(matched) < 1 {
		return diag.Errorf("Your vnpaycloud_server_group query returned no results")
	}

	if len(matched) > 1 {
		return diag.Errorf("Your vnpaycloud_server_group query returned multiple results")
	}

	tflog.Debug(ctx, "Retrieved vnpaycloud_server_group datasource", map[string]interface{}{"server_group": matched[0]})
	setServerGroupDataSourceAttributes(d, matched[0])

	return nil
}

func setServerGroupDataSourceAttributes(d *schema.ResourceData, sg dto.ServerGroup) {
	d.SetId(sg.ID)
	d.Set("name", sg.Name)
	d.Set("policy", sg.Policy)
	d.Set("member_ids", sg.MemberIDs)
	d.Set("created_at", sg.CreatedAt)
}

func DataSourceServerGroups() *schema.Resource {
	return &schema.Resource{
		ReadContext: dataSourceServerGroupsRead,
		Schema: map[string]*schema.Schema{
			"server_groups": {
				Type:     schema.TypeList,
				Computed: true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"id":     {Type: schema.TypeString, Computed: true},
						"name":   {Type: schema.TypeString, Computed: true},
						"policy": {Type: schema.TypeString, Computed: true},
						"member_ids": {
							Type:     schema.TypeList,
							Computed: true,
							Elem:     &schema.Schema{Type: schema.TypeString},
						},
						"created_at": {Type: schema.TypeString, Computed: true},
					},
				},
			},
		},
	}
}

func dataSourceServerGroupsRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	cfg := meta.(*config.Config)

	listResp := &dto.ListServerGroupsResponse{}
	_, err := cfg.Client.Get(ctx, client.ApiPath.ServerGroups(cfg.ProjectID), listResp, nil)
	if err != nil {
		return diag.Errorf("Error listing vnpaycloud_server_groups: %s", err)
	}

	var serverGroups []map[string]interface{}
	for _, sg := range listResp.ServerGroups {
		serverGroups = append(serverGroups, map[string]interface{}{
			"id":         sg.ID,
			"name":       sg.Name,
			"policy":     sg.Policy,
			"member_ids": sg.MemberIDs,
			"created_at": sg.CreatedAt,
		})
	}

	d.SetId(fmt.Sprintf("server-groups-%s", cfg.ProjectID))
	d.Set("server_groups", serverGroups)

	return nil
}
