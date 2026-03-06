package securitygroup

import (
	"context"
	"terraform-provider-vnpaycloud/vnpaycloud/config"
	"terraform-provider-vnpaycloud/vnpaycloud/dto"
	"terraform-provider-vnpaycloud/vnpaycloud/helper/client"
	"terraform-provider-vnpaycloud/vnpaycloud/util"

	"github.com/hashicorp/terraform-plugin-log/tflog"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

func ResourceSecurityGroup() *schema.Resource {
	return &schema.Resource{
		CreateContext: resourceSecurityGroupCreate,
		ReadContext:   resourceSecurityGroupRead,
		UpdateContext: resourceSecurityGroupUpdate,
		DeleteContext: resourceSecurityGroupDelete,
		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},
		Schema: map[string]*schema.Schema{
			"name": {
				Type:     schema.TypeString,
				Required: true,
			},
			"description": {
				Type:     schema.TypeString,
				Optional: true,
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

func resourceSecurityGroupCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	cfg := meta.(*config.Config)

	createOpts := dto.CreateSecurityGroupRequest{
		Name:        d.Get("name").(string),
		Description: d.Get("description").(string),
	}

	tflog.Debug(ctx, "vnpaycloud_security_group create options", map[string]interface{}{"create_opts": createOpts})

	createResp := &dto.SecurityGroupResponse{}
	_, err := cfg.Client.Post(ctx, client.ApiPath.SecurityGroups(cfg.ProjectID), createOpts, createResp, nil)
	if err != nil {
		return diag.Errorf("Error creating vnpaycloud_security_group: %s", err)
	}

	d.SetId(createResp.SecurityGroup.ID)

	return resourceSecurityGroupRead(ctx, d, meta)
}

func resourceSecurityGroupRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	cfg := meta.(*config.Config)

	sgResp := &dto.SecurityGroupResponse{}
	_, err := cfg.Client.Get(ctx, client.ApiPath.SecurityGroupWithID(cfg.ProjectID, d.Id()), sgResp, nil)
	if err != nil {
		return diag.FromErr(util.CheckNotFound(d, err, "Error retrieving vnpaycloud_security_group"))
	}

	tflog.Debug(ctx, "Retrieved vnpaycloud_security_group "+d.Id(), map[string]interface{}{"security_group": sgResp.SecurityGroup})

	d.Set("name", sgResp.SecurityGroup.Name)
	d.Set("description", sgResp.SecurityGroup.Description)
	d.Set("created_at", sgResp.SecurityGroup.CreatedAt)

	rules := make([]map[string]interface{}, len(sgResp.SecurityGroup.Rules))
	for i, r := range sgResp.SecurityGroup.Rules {
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

func resourceSecurityGroupUpdate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	cfg := meta.(*config.Config)

	if d.HasChanges("name", "description") {
		updateOpts := dto.UpdateSecurityGroupRequest{
			Name:        d.Get("name").(string),
			Description: d.Get("description").(string),
		}

		tflog.Debug(ctx, "vnpaycloud_security_group update options", map[string]interface{}{"update_opts": updateOpts})

		_, err := cfg.Client.Put(ctx, client.ApiPath.SecurityGroupWithID(cfg.ProjectID, d.Id()), updateOpts, nil, nil)
		if err != nil {
			return diag.Errorf("Error updating vnpaycloud_security_group %s: %s", d.Id(), err)
		}
	}

	return resourceSecurityGroupRead(ctx, d, meta)
}

func resourceSecurityGroupDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	cfg := meta.(*config.Config)

	if _, err := cfg.Client.Delete(ctx, client.ApiPath.SecurityGroupWithID(cfg.ProjectID, d.Id()), nil); err != nil {
		return diag.FromErr(util.CheckDeleted(d, err, "Error deleting vnpaycloud_security_group"))
	}

	return nil
}
