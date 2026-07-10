package kubernetesrbac

import (
	"context"
	"fmt"

	"terraform-provider-vnpaycloud/vnpaycloud/config"
	"terraform-provider-vnpaycloud/vnpaycloud/dto"
	"terraform-provider-vnpaycloud/vnpaycloud/helper/client"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

// DataSourceKubernetesRoles lists the roles available in a cluster
// (cluster-admin/admin/edit/view are seeded per cluster).
func DataSourceKubernetesRoles() *schema.Resource {
	return &schema.Resource{
		ReadContext: dataSourceKubernetesRolesRead,
		Schema: map[string]*schema.Schema{
			"cluster_id": {
				Type:     schema.TypeString,
				Required: true,
			},
			"roles": {
				Type:     schema.TypeList,
				Computed: true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"id":              {Type: schema.TypeString, Computed: true},
						"name":            {Type: schema.TypeString, Computed: true},
						"namespace":       {Type: schema.TypeString, Computed: true},
						"is_cluster_role": {Type: schema.TypeBool, Computed: true},
						"status":          {Type: schema.TypeString, Computed: true},
						"created_at":      {Type: schema.TypeString, Computed: true},
					},
				},
			},
		},
	}
}

func dataSourceKubernetesRolesRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	cfg := meta.(*config.Config)
	clusterID := d.Get("cluster_id").(string)

	listResp := &dto.ListKubernetesRolesResponse{}
	_, err := cfg.Client.Get(ctx, client.ApiPath.KubernetesRoles(cfg.ProjectID, clusterID), listResp, nil)
	if err != nil {
		return diag.Errorf("Error listing vnpaycloud_kubernetes_roles: %s", err)
	}

	var roles []map[string]interface{}
	for _, r := range listResp.KubernetesRoles {
		roles = append(roles, map[string]interface{}{
			"id":              r.ID,
			"name":            r.Name,
			"namespace":       r.Namespace,
			"is_cluster_role": r.IsClusterRole,
			"status":          r.Status,
			"created_at":      r.CreatedAt,
		})
	}

	d.SetId(fmt.Sprintf("kubernetes-roles-%s-%s", cfg.ProjectID, clusterID))
	d.Set("roles", roles)

	return nil
}

// DataSourceKubernetesRbacs lists the RBAC role bindings of a cluster.
func DataSourceKubernetesRbacs() *schema.Resource {
	return &schema.Resource{
		ReadContext: dataSourceKubernetesRbacsRead,
		Schema: map[string]*schema.Schema{
			"cluster_id": {
				Type:     schema.TypeString,
				Required: true,
			},
			"rbacs": {
				Type:     schema.TypeList,
				Computed: true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"id":                      {Type: schema.TypeString, Computed: true},
						"user_id":                 {Type: schema.TypeString, Computed: true},
						"email":                   {Type: schema.TypeString, Computed: true},
						"role":                    {Type: schema.TypeString, Computed: true},
						"kubernetes_role_id":      {Type: schema.TypeString, Computed: true},
						"binding_type":            {Type: schema.TypeString, Computed: true},
						"is_cluster_role_binding": {Type: schema.TypeBool, Computed: true},
						"namespace":               {Type: schema.TypeString, Computed: true},
						"name":                    {Type: schema.TypeString, Computed: true},
						"status":                  {Type: schema.TypeString, Computed: true},
						"created_at":              {Type: schema.TypeString, Computed: true},
					},
				},
			},
		},
	}
}

func dataSourceKubernetesRbacsRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	cfg := meta.(*config.Config)
	clusterID := d.Get("cluster_id").(string)

	listResp := &dto.ListKubernetesRbacsResponse{}
	_, err := cfg.Client.Get(ctx, client.ApiPath.KubernetesRbacs(cfg.ProjectID, clusterID), listResp, nil)
	if err != nil {
		return diag.Errorf("Error listing vnpaycloud_kubernetes_rbacs: %s", err)
	}

	var rbacs []map[string]interface{}
	for _, rb := range listResp.Rbacs {
		rbacs = append(rbacs, map[string]interface{}{
			"id":                      rb.ID,
			"user_id":                 rb.UserID,
			"email":                   rb.Email,
			"role":                    rb.Role,
			"kubernetes_role_id":      rb.KubernetesRoleID,
			"binding_type":            rb.BindingType,
			"is_cluster_role_binding": rb.IsClusterRoleBinding,
			"namespace":               rb.Namespace,
			"name":                    rb.Name,
			"status":                  rb.Status,
			"created_at":              rb.CreatedAt,
		})
	}

	d.SetId(fmt.Sprintf("kubernetes-rbacs-%s-%s", cfg.ProjectID, clusterID))
	d.Set("rbacs", rbacs)

	return nil
}
