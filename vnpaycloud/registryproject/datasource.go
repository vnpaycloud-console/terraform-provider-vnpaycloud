package registryproject

import (
	"context"
	"terraform-provider-vnpaycloud/vnpaycloud/config"
	"terraform-provider-vnpaycloud/vnpaycloud/dto"
	"terraform-provider-vnpaycloud/vnpaycloud/helper/client"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

func DataSourceRegistryProject() *schema.Resource {
	return &schema.Resource{
		ReadContext: dataSourceRegistryProjectRead,
		Schema: map[string]*schema.Schema{
			"id": {
				Type:     schema.TypeString,
				Required: true,
			},
			"name": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"is_public": {
				Type:     schema.TypeBool,
				Computed: true,
			},
			"storage_limit": {
				Type:     schema.TypeInt,
				Computed: true,
			},
			"storage_used": {
				Type:     schema.TypeInt,
				Computed: true,
			},
			"repo_count": {
				Type:     schema.TypeInt,
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

func dataSourceRegistryProjectRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	cfg := meta.(*config.Config)

	id := d.Get("id").(string)

	resp := &dto.RegistryProjectResponse{}
	_, err := cfg.Client.Get(ctx, client.ApiPath.RegistryProjectWithID(cfg.ProjectID, id), resp, nil)
	if err != nil {
		return diag.Errorf("Error retrieving vnpaycloud_registry_project %s: %s", id, err)
	}

	d.SetId(resp.Registry.ID)
	d.Set("name", resp.Registry.Name)
	d.Set("is_public", resp.Registry.IsPublic)
	d.Set("storage_limit", resp.Registry.StorageLimit)
	d.Set("storage_used", resp.Registry.StorageUsed)
	d.Set("repo_count", resp.Registry.RepoCount)
	d.Set("status", resp.Registry.Status)
	d.Set("created_at", resp.Registry.CreatedAt)

	return nil
}

func DataSourceRegistryProjects() *schema.Resource {
	return &schema.Resource{
		ReadContext: dataSourceRegistryProjectsRead,
		Schema: map[string]*schema.Schema{
			"registries": {
				Type:     schema.TypeList,
				Computed: true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"id": {
							Type:     schema.TypeString,
							Computed: true,
						},
						"name": {
							Type:     schema.TypeString,
							Computed: true,
						},
						"is_public": {
							Type:     schema.TypeBool,
							Computed: true,
						},
						"storage_limit": {
							Type:     schema.TypeInt,
							Computed: true,
						},
						"storage_used": {
							Type:     schema.TypeInt,
							Computed: true,
						},
						"repo_count": {
							Type:     schema.TypeInt,
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
				},
			},
		},
	}
}

func dataSourceRegistryProjectsRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	cfg := meta.(*config.Config)

	resp := &dto.ListRegistryProjectsResponse{}
	_, err := cfg.Client.Get(ctx, client.ApiPath.RegistryProjects(cfg.ProjectID), resp, nil)
	if err != nil {
		return diag.Errorf("Error listing vnpaycloud_registry_projects: %s", err)
	}

	var registries []map[string]interface{}
	for _, r := range resp.Registries {
		registries = append(registries, map[string]interface{}{
			"id":            r.ID,
			"name":          r.Name,
			"is_public":     r.IsPublic,
			"storage_limit": r.StorageLimit,
			"storage_used":  r.StorageUsed,
			"repo_count":    r.RepoCount,
			"status":        r.Status,
			"created_at":    r.CreatedAt,
		})
	}

	d.SetId(cfg.ProjectID)
	d.Set("registries", registries)

	return nil
}
