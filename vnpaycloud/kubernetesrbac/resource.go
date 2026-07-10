package kubernetesrbac

import (
	"context"
	"fmt"
	"strings"

	"github.com/hashicorp/terraform-plugin-log/tflog"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"

	"terraform-provider-vnpaycloud/vnpaycloud/config"
	"terraform-provider-vnpaycloud/vnpaycloud/dto"
	"terraform-provider-vnpaycloud/vnpaycloud/helper/client"
	"terraform-provider-vnpaycloud/vnpaycloud/util"
)

const (
	bindingTypeClusterRoleBinding = "cluster_role_binding"
	bindingTypeRoleBinding        = "role_binding"
)

func ResourceKubernetesRbac() *schema.Resource {
	return &schema.Resource{
		CreateContext: resourceKubernetesRbacCreate,
		ReadContext:   resourceKubernetesRbacRead,
		DeleteContext: resourceKubernetesRbacDelete,
		Importer: &schema.ResourceImporter{
			StateContext: func(ctx context.Context, d *schema.ResourceData, meta interface{}) ([]*schema.ResourceData, error) {
				parts := strings.SplitN(d.Id(), "/", 2)
				if len(parts) != 2 || parts[0] == "" || parts[1] == "" {
					return nil, fmt.Errorf("invalid import ID %q, expected \"<cluster_id>/<rbac_id>\"", d.Id())
				}
				d.Set("cluster_id", parts[0])
				d.SetId(parts[1])
				return []*schema.ResourceData{d}, nil
			},
		},
		Schema: map[string]*schema.Schema{
			"cluster_id": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},
			"user_id": {
				Type:        schema.TypeString,
				Required:    true,
				ForceNew:    true,
				Description: "The portal user id to bind the role to, e.g. iaas.portal.usr.xxxxxxxx.",
			},
			"role": {
				Type:         schema.TypeString,
				Required:     true,
				ForceNew:     true,
				ValidateFunc: validation.StringIsNotEmpty,
				Description:  "The role name to bind, e.g. cluster-admin, admin, edit, view.",
			},
			"binding_type": {
				Type:         schema.TypeString,
				Optional:     true,
				ForceNew:     true,
				Default:      bindingTypeClusterRoleBinding,
				ValidateFunc: validation.StringInSlice([]string{bindingTypeClusterRoleBinding, bindingTypeRoleBinding}, false),
				Description:  "The binding type: cluster_role_binding (default) or role_binding.",
			},
			"namespace": {
				Type:        schema.TypeString,
				Optional:    true,
				ForceNew:    true,
				Description: "The namespace, required when binding_type is role_binding.",
			},

			// Computed attributes
			"kubernetes_role_id": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"is_cluster_role_binding": {
				Type:     schema.TypeBool,
				Computed: true,
			},
			"name": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"email": {
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

func resourceKubernetesRbacCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	cfg := meta.(*config.Config)
	clusterID := d.Get("cluster_id").(string)

	createOpts := dto.CreateKubernetesRbacRequest{
		UserID:      d.Get("user_id").(string),
		Role:        d.Get("role").(string),
		BindingType: d.Get("binding_type").(string),
		Namespace:   d.Get("namespace").(string),
	}

	tflog.Debug(ctx, "vnpaycloud_kubernetes_rbac create options", map[string]interface{}{"create_opts": createOpts})

	createResp := &dto.KubernetesRbacResponse{}
	_, err := cfg.Client.Post(ctx, client.ApiPath.KubernetesRbacs(cfg.ProjectID, clusterID), createOpts, createResp, nil)
	if err != nil {
		return diag.Errorf("Error creating vnpaycloud_kubernetes_rbac: %s", err)
	}

	d.SetId(createResp.Rbac.ID)

	return resourceKubernetesRbacRead(ctx, d, meta)
}

func resourceKubernetesRbacRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	cfg := meta.(*config.Config)
	clusterID := d.Get("cluster_id").(string)

	resp := &dto.KubernetesRbacResponse{}
	_, err := cfg.Client.Get(ctx, client.ApiPath.KubernetesRbacWithID(cfg.ProjectID, clusterID, d.Id()), resp, nil)
	if err != nil {
		return diag.FromErr(util.CheckNotFound(d, err, "Error retrieving vnpaycloud_kubernetes_rbac"))
	}

	tflog.Debug(ctx, "Retrieved vnpaycloud_kubernetes_rbac "+d.Id(), map[string]interface{}{"rbac": resp.Rbac})

	rbac := resp.Rbac
	d.Set("cluster_id", rbac.ClusterID)
	d.Set("user_id", rbac.UserID)
	d.Set("role", rbac.Role)
	d.Set("binding_type", rbac.BindingType)
	d.Set("namespace", rbac.Namespace)
	d.Set("kubernetes_role_id", rbac.KubernetesRoleID)
	d.Set("is_cluster_role_binding", rbac.IsClusterRoleBinding)
	d.Set("name", rbac.Name)
	d.Set("email", rbac.Email)
	d.Set("status", rbac.Status)
	d.Set("created_at", rbac.CreatedAt)

	return nil
}

func resourceKubernetesRbacDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	cfg := meta.(*config.Config)
	clusterID := d.Get("cluster_id").(string)

	if _, err := cfg.Client.Delete(ctx, client.ApiPath.KubernetesRbacWithID(cfg.ProjectID, clusterID, d.Id()), nil); err != nil {
		return diag.FromErr(util.CheckDeleted(d, err, "Error deleting vnpaycloud_kubernetes_rbac"))
	}

	return nil
}
