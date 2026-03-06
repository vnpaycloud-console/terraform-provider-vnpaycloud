package securitygroup

import (
	"context"
	"terraform-provider-vnpaycloud/vnpaycloud/config"
	"terraform-provider-vnpaycloud/vnpaycloud/dto"
	"terraform-provider-vnpaycloud/vnpaycloud/helper/client"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

func DataSourceSecurityGroup() *schema.Resource {
	return &schema.Resource{
		ReadContext: dataSourceSecurityGroupRead,
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
			"description": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"status": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"rules": {
				Type:     schema.TypeList,
				Computed: true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"id":                {Type: schema.TypeString, Computed: true},
						"security_group_id": {Type: schema.TypeString, Computed: true},
						"direction":         {Type: schema.TypeString, Computed: true},
						"protocol":          {Type: schema.TypeString, Computed: true},
						"ethertype":         {Type: schema.TypeString, Computed: true},
						"port_range_min":    {Type: schema.TypeInt, Computed: true},
						"port_range_max":    {Type: schema.TypeInt, Computed: true},
						"remote_ip_prefix":  {Type: schema.TypeString, Computed: true},
						"remote_group_id":   {Type: schema.TypeString, Computed: true},
					},
				},
			},
			"created_at": {
				Type:     schema.TypeString,
				Computed: true,
			},
		},
	}
}

func dataSourceSecurityGroupRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	cfg := meta.(*config.Config)

	if id, ok := d.GetOk("id"); ok {
		sgResp := &dto.SecurityGroupResponse{}
		_, err := cfg.Client.Get(ctx, client.ApiPath.SecurityGroupWithID(cfg.ProjectID, id.(string)), sgResp, nil)
		if err != nil {
			return diag.Errorf("Error fetching vnpaycloud_security_group %s: %s", id, err)
		}
		return setSecurityGroupData(d, &sgResp.SecurityGroup)
	}

	// List and filter client-side
	listResp := &dto.ListSecurityGroupsResponse{}
	_, err := cfg.Client.Get(ctx, client.ApiPath.SecurityGroups(cfg.ProjectID), listResp, nil)
	if err != nil {
		return diag.Errorf("Error listing vnpaycloud_security_group: %s", err)
	}

	nameFilter, nameOk := d.GetOk("name")

	for _, sg := range listResp.SecurityGroups {
		if nameOk && sg.Name != nameFilter.(string) {
			continue
		}
		return setSecurityGroupData(d, &sg)
	}

	return diag.Errorf("No vnpaycloud_security_group found matching the criteria")
}

func setSecurityGroupData(d *schema.ResourceData, sg *dto.SecurityGroup) diag.Diagnostics {
	d.SetId(sg.ID)
	d.Set("name", sg.Name)
	d.Set("description", sg.Description)
	d.Set("status", sg.Status)
	d.Set("created_at", sg.CreatedAt)

	rules := make([]map[string]interface{}, len(sg.Rules))
	for i, r := range sg.Rules {
		rules[i] = map[string]interface{}{
			"id":                r.ID,
			"security_group_id": r.SecurityGroupID,
			"direction":         r.Direction,
			"protocol":          r.Protocol,
			"ethertype":         r.EtherType,
			"port_range_min":    r.PortRangeMin,
			"port_range_max":    r.PortRangeMax,
			"remote_ip_prefix":  r.RemoteIPPrefix,
			"remote_group_id":   r.RemoteGroupID,
		}
	}
	d.Set("rules", rules)

	return nil
}
