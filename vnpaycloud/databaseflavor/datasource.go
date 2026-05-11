package databaseflavor

import (
	"context"
	"terraform-provider-vnpaycloud/vnpaycloud/config"
	"terraform-provider-vnpaycloud/vnpaycloud/dto"
	"terraform-provider-vnpaycloud/vnpaycloud/helper/client"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

func DataSourceDatabaseFlavor() *schema.Resource {
	return &schema.Resource{
		ReadContext: dataSourceDatabaseFlavorRead,
		Schema: map[string]*schema.Schema{
			"id": {
				Type:     schema.TypeString,
				Required: true,
			},
			"name": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"class": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"ratio": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"cpu_req": {
				Type:     schema.TypeInt,
				Computed: true,
			},
			"mem_req": {
				Type:     schema.TypeInt,
				Computed: true,
			},
			"cpu_limit": {
				Type:     schema.TypeInt,
				Computed: true,
			},
			"mem_limit": {
				Type:     schema.TypeInt,
				Computed: true,
			},
		},
	}
}

func dataSourceDatabaseFlavorRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	cfg := meta.(*config.Config)
	id := d.Get("id").(string)

	// List all flavors and find by ID
	resp := &dto.ListFlavorDatabasesResponse{}
	_, err := cfg.Client.Get(ctx, client.ApiPath.DatabaseFlavors(cfg.ProjectID), resp, nil)
	if err != nil {
		return diag.Errorf("Error listing database flavors: %s", err)
	}

	for _, f := range resp.FlavorDatabases {
		if f.ID == id {
			d.SetId(f.ID)
			d.Set("name", f.Name)
			d.Set("class", f.Class)
			d.Set("ratio", f.Ratio)
			d.Set("cpu_req", f.CpuReq)
			d.Set("mem_req", f.MemReq)
			d.Set("cpu_limit", f.CpuLimit)
			d.Set("mem_limit", f.MemLimit)
			return nil
		}
	}

	return diag.Errorf("Database flavor %s not found", id)
}

func DataSourceDatabaseFlavors() *schema.Resource {
	return &schema.Resource{
		ReadContext: dataSourceDatabaseFlavorsRead,
		Schema: map[string]*schema.Schema{
			"flavors": {
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
						"class": {
							Type:     schema.TypeString,
							Computed: true,
						},
						"ratio": {
							Type:     schema.TypeString,
							Computed: true,
						},
						"cpu_req": {
							Type:     schema.TypeInt,
							Computed: true,
						},
						"mem_req": {
							Type:     schema.TypeInt,
							Computed: true,
						},
						"cpu_limit": {
							Type:     schema.TypeInt,
							Computed: true,
						},
						"mem_limit": {
							Type:     schema.TypeInt,
							Computed: true,
						},
					},
				},
			},
		},
	}
}

func dataSourceDatabaseFlavorsRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	cfg := meta.(*config.Config)

	resp := &dto.ListFlavorDatabasesResponse{}
	_, err := cfg.Client.Get(ctx, client.ApiPath.DatabaseFlavors(cfg.ProjectID), resp, nil)
	if err != nil {
		return diag.Errorf("Error listing database flavors: %s", err)
	}

	var flavors []map[string]interface{}
	for _, f := range resp.FlavorDatabases {
		flavors = append(flavors, map[string]interface{}{
			"id":        f.ID,
			"name":      f.Name,
			"class":     f.Class,
			"ratio":     f.Ratio,
			"cpu_req":   f.CpuReq,
			"mem_req":   f.MemReq,
			"cpu_limit": f.CpuLimit,
			"mem_limit": f.MemLimit,
		})
	}

	d.SetId("database_flavors")
	d.Set("flavors", flavors)

	return nil
}
