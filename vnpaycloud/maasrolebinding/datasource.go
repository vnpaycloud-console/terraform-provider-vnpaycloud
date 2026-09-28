package maasrolebinding

import (
	"context"
	"net/url"
	"terraform-provider-vnpaycloud/vnpaycloud/config"
	"terraform-provider-vnpaycloud/vnpaycloud/dto"
	"terraform-provider-vnpaycloud/vnpaycloud/helper/client"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
)

func DataSourceMaasRoleBinding() *schema.Resource {
	return &schema.Resource{
		ReadContext: dataSourceMaasRoleBindingRead,
		Schema: map[string]*schema.Schema{
			"id": {
				Type:         schema.TypeString,
				Optional:     true,
				Computed:     true,
				AtLeastOneOf: []string{"id", "portal_user_id"},
			},
			"portal_user_id": {
				Type:         schema.TypeString,
				Optional:     true,
				Computed:     true,
				AtLeastOneOf: []string{"id", "portal_user_id"},
			},
			"role":       {Type: schema.TypeString, Computed: true},
			"status":     {Type: schema.TypeString, Computed: true},
			"created_at": {Type: schema.TypeString, Computed: true},
		},
	}
}

func dataSourceMaasRoleBindingRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	cfg := meta.(*config.Config)

	if id, ok := d.GetOk("id"); ok && id.(string) != "" {
		resp := &dto.MaasRoleBindingResponse{}
		if _, err := cfg.Client.Get(ctx, client.ApiPath.MaasRoleBindingWithID(cfg.ProjectID, id.(string)), resp, nil); err != nil {
			return diag.Errorf("Error fetching vnpaycloud_maas_role_binding %s: %s", id, err)
		}
		setRoleBindingData(d, &resp.RoleBinding)
		return nil
	}

	portalUserID := d.Get("portal_user_id").(string)
	listResp := &dto.ListMaasRoleBindingsResponse{}
	path := client.ApiPath.MaasRoleBindings(cfg.ProjectID) + "?portalUserId=" + url.QueryEscape(portalUserID)
	if _, err := cfg.Client.Get(ctx, path, listResp, nil); err != nil {
		return diag.Errorf("Error listing vnpaycloud_maas_role_binding: %s", err)
	}

	var matched []dto.MaasRoleBinding
	for _, b := range listResp.RoleBindings {
		if b.PortalUserID == portalUserID {
			matched = append(matched, b)
		}
	}

	if len(matched) == 0 {
		return diag.Errorf("No vnpaycloud_maas_role_binding found for portal_user_id %q", portalUserID)
	}
	if len(matched) > 1 {
		return diag.Errorf("Found %d vnpaycloud_maas_role_binding for portal_user_id %q; use id to select one", len(matched), portalUserID)
	}

	setRoleBindingData(d, &matched[0])
	return nil
}

func setRoleBindingData(d *schema.ResourceData, b *dto.MaasRoleBinding) {
	d.SetId(b.ID)
	d.Set("portal_user_id", b.PortalUserID)
	d.Set("role", b.Role)
	d.Set("status", b.Status)
	d.Set("created_at", b.CreatedAt)
}

func DataSourceMaasRoleBindings() *schema.Resource {
	return &schema.Resource{
		ReadContext: dataSourceMaasRoleBindingsRead,
		Schema: map[string]*schema.Schema{
			"portal_user_id": {
				Type:     schema.TypeString,
				Optional: true,
			},
			"role": {
				Type:         schema.TypeString,
				Optional:     true,
				ValidateFunc: validation.StringInSlice(roles, false),
			},
			"role_bindings": {
				Type:     schema.TypeList,
				Computed: true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"id":             {Type: schema.TypeString, Computed: true},
						"portal_user_id": {Type: schema.TypeString, Computed: true},
						"role":           {Type: schema.TypeString, Computed: true},
						"status":         {Type: schema.TypeString, Computed: true},
						"created_at":     {Type: schema.TypeString, Computed: true},
					},
				},
			},
		},
	}
}

func dataSourceMaasRoleBindingsRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	cfg := meta.(*config.Config)

	path := client.ApiPath.MaasRoleBindings(cfg.ProjectID)
	q := url.Values{}
	if v, ok := d.GetOk("portal_user_id"); ok {
		q.Set("portalUserId", v.(string))
	}
	if v, ok := d.GetOk("role"); ok {
		q.Set("role", v.(string))
	}
	if len(q) > 0 {
		path += "?" + q.Encode()
	}

	listResp := &dto.ListMaasRoleBindingsResponse{}
	if _, err := cfg.Client.Get(ctx, path, listResp, nil); err != nil {
		return diag.Errorf("Unable to query vnpaycloud_maas_role_bindings: %s", err)
	}

	bindings := make([]map[string]interface{}, 0, len(listResp.RoleBindings))
	for _, b := range listResp.RoleBindings {
		bindings = append(bindings, map[string]interface{}{
			"id":             b.ID,
			"portal_user_id": b.PortalUserID,
			"role":           b.Role,
			"status":         b.Status,
			"created_at":     b.CreatedAt,
		})
	}

	d.SetId(cfg.ProjectID)
	d.Set("role_bindings", bindings)

	return nil
}
