package maasrolebinding

import (
	"context"
	"terraform-provider-vnpaycloud/vnpaycloud/config"
	"terraform-provider-vnpaycloud/vnpaycloud/dto"
	"terraform-provider-vnpaycloud/vnpaycloud/helper/client"
	"terraform-provider-vnpaycloud/vnpaycloud/util"

	"github.com/hashicorp/terraform-plugin-log/tflog"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
)

// The service assigns only these two; the organization owner holds admin implicitly.
var assignableRoles = []string{"editor", "viewer"}

// Filters may still match the owner row.
var roles = []string{"admin", "editor", "viewer"}

func ResourceMaasRoleBinding() *schema.Resource {
	return &schema.Resource{
		CreateContext: resourceMaasRoleBindingCreate,
		ReadContext:   resourceMaasRoleBindingRead,
		UpdateContext: resourceMaasRoleBindingUpdate,
		DeleteContext: resourceMaasRoleBindingDelete,
		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},
		Schema: map[string]*schema.Schema{
			"portal_user_id": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},
			"role": {
				Type:         schema.TypeString,
				Required:     true,
				ValidateFunc: validation.StringInSlice(assignableRoles, false),
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

func resourceMaasRoleBindingCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	cfg := meta.(*config.Config)

	createOpts := dto.CreateMaasRoleBindingRequest{
		PortalUserID: d.Get("portal_user_id").(string),
		Role:         d.Get("role").(string),
	}

	tflog.Debug(ctx, "vnpaycloud_maas_role_binding create options", map[string]interface{}{
		"portal_user_id": createOpts.PortalUserID,
		"role":           createOpts.Role,
	})

	createResp := &dto.MaasRoleBindingResponse{}
	if _, err := cfg.Client.Post(ctx, client.ApiPath.MaasRoleBindings(cfg.ProjectID), createOpts, createResp, nil); err != nil {
		return diag.Errorf("Error creating vnpaycloud_maas_role_binding: %s", err)
	}

	d.SetId(createResp.RoleBinding.ID)

	return resourceMaasRoleBindingRead(ctx, d, meta)
}

func resourceMaasRoleBindingRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	cfg := meta.(*config.Config)

	resp := &dto.MaasRoleBindingResponse{}
	if _, err := cfg.Client.Get(ctx, client.ApiPath.MaasRoleBindingWithID(cfg.ProjectID, d.Id()), resp, nil); err != nil {
		return diag.FromErr(util.CheckNotFound(d, err, "Error retrieving vnpaycloud_maas_role_binding"))
	}

	b := resp.RoleBinding
	d.Set("portal_user_id", b.PortalUserID)
	d.Set("role", b.Role)
	d.Set("status", b.Status)
	d.Set("created_at", b.CreatedAt)

	return nil
}

func resourceMaasRoleBindingUpdate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	cfg := meta.(*config.Config)

	if d.HasChange("role") {
		updateOpts := dto.UpdateMaasRoleBindingRequest{
			Role: d.Get("role").(string),
		}
		if _, err := cfg.Client.Put(ctx, client.ApiPath.MaasRoleBindingWithID(cfg.ProjectID, d.Id()), updateOpts, nil, nil); err != nil {
			return diag.Errorf("Error updating vnpaycloud_maas_role_binding %s: %s", d.Id(), err)
		}
	}

	return resourceMaasRoleBindingRead(ctx, d, meta)
}

func resourceMaasRoleBindingDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	cfg := meta.(*config.Config)

	if _, err := cfg.Client.Delete(ctx, client.ApiPath.MaasRoleBindingWithID(cfg.ProjectID, d.Id()), nil); err != nil {
		return diag.FromErr(util.CheckDeleted(d, err, "Error deleting vnpaycloud_maas_role_binding"))
	}

	return nil
}
