package registryproject

import (
	"context"
	"terraform-provider-vnpaycloud/vnpaycloud/config"
	"terraform-provider-vnpaycloud/vnpaycloud/dto"
	"terraform-provider-vnpaycloud/vnpaycloud/helper/client"
	"terraform-provider-vnpaycloud/vnpaycloud/util"
	"time"

	"github.com/hashicorp/terraform-plugin-log/tflog"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/retry"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

func ResourceRegistryProject() *schema.Resource {
	return &schema.Resource{
		CreateContext: resourceRegistryProjectCreate,
		ReadContext:   resourceRegistryProjectRead,
		DeleteContext: resourceRegistryProjectDelete,
		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},
		Timeouts: &schema.ResourceTimeout{
			Create: schema.DefaultTimeout(10 * time.Minute),
			Delete: schema.DefaultTimeout(10 * time.Minute),
		},
		Schema: map[string]*schema.Schema{
			"name": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},
			"is_public": {
				Type:     schema.TypeBool,
				Optional: true,
				ForceNew: true,
				Default:  false,
			},
			"storage_limit": {
				Type:     schema.TypeInt,
				Optional: true,
				ForceNew: true,
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

func resourceRegistryProjectCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	cfg := meta.(*config.Config)

	createOpts := dto.CreateRegistryProjectRequest{
		Name:     d.Get("name").(string),
		IsPublic: d.Get("is_public").(bool),
	}

	if v, ok := d.GetOk("storage_limit"); ok {
		createOpts.StorageLimit = int64(v.(int))
	}

	tflog.Debug(ctx, "vnpaycloud_registry_project create options", map[string]interface{}{"create_opts": createOpts})

	createResp := &dto.RegistryProjectResponse{}
	_, err := cfg.Client.Post(ctx, client.ApiPath.RegistryProjects(cfg.ProjectID), createOpts, createResp, nil)
	if err != nil {
		return diag.Errorf("Error creating vnpaycloud_registry_project: %s", err)
	}

	d.SetId(createResp.Registry.ID)

	stateConf := &retry.StateChangeConf{
		Pending:    []string{"initiating", "creating", "pending_create", "unknown"},
		Target:     []string{"active", "created"},
		Refresh:    registryProjectStateRefreshFunc(ctx, cfg.Client, cfg.ProjectID, createResp.Registry.ID),
		Timeout:    d.Timeout(schema.TimeoutCreate),
		Delay:      5 * time.Second,
		MinTimeout: 3 * time.Second,
	}

	_, err = stateConf.WaitForStateContext(ctx)
	if err != nil {
		return diag.Errorf("Error waiting for vnpaycloud_registry_project %s to become ready: %s", createResp.Registry.ID, err)
	}

	return resourceRegistryProjectRead(ctx, d, meta)
}

func resourceRegistryProjectRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	cfg := meta.(*config.Config)

	resp := &dto.RegistryProjectResponse{}
	_, err := cfg.Client.Get(ctx, client.ApiPath.RegistryProjectWithID(cfg.ProjectID, d.Id()), resp, nil)
	if err != nil {
		return diag.FromErr(util.CheckNotFound(d, err, "Error retrieving vnpaycloud_registry_project"))
	}

	tflog.Debug(ctx, "Retrieved vnpaycloud_registry_project "+d.Id(), map[string]interface{}{"registry": resp.Registry})

	d.Set("name", resp.Registry.Name)
	d.Set("is_public", resp.Registry.IsPublic)
	d.Set("storage_limit", resp.Registry.StorageLimit)
	d.Set("storage_used", resp.Registry.StorageUsed)
	d.Set("repo_count", resp.Registry.RepoCount)
	d.Set("status", resp.Registry.Status)
	d.Set("created_at", resp.Registry.CreatedAt)

	return nil
}

func resourceRegistryProjectDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	cfg := meta.(*config.Config)

	if _, err := cfg.Client.Delete(ctx, client.ApiPath.RegistryProjectWithID(cfg.ProjectID, d.Id()), nil); err != nil {
		return diag.FromErr(util.CheckDeleted(d, err, "Error deleting vnpaycloud_registry_project"))
	}

	stateConf := &retry.StateChangeConf{
		Pending:    []string{"deleting", "active", "created", "pending_delete", "unknown"},
		Target:     []string{"deleted"},
		Refresh:    registryProjectStateRefreshFunc(ctx, cfg.Client, cfg.ProjectID, d.Id()),
		Timeout:    d.Timeout(schema.TimeoutDelete),
		Delay:      5 * time.Second,
		MinTimeout: 3 * time.Second,
	}

	_, err := stateConf.WaitForStateContext(ctx)
	if err != nil {
		return diag.Errorf("Error waiting for vnpaycloud_registry_project %s to delete: %s", d.Id(), err)
	}

	return nil
}
