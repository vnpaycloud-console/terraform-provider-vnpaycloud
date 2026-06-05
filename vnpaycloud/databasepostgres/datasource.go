package databasepostgres

import (
	"context"
	"terraform-provider-vnpaycloud/vnpaycloud/config"
	"terraform-provider-vnpaycloud/vnpaycloud/dto"
	"terraform-provider-vnpaycloud/vnpaycloud/helper/client"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

func DataSourceDatabasePostgresInstance() *schema.Resource {
	return &schema.Resource{
		ReadContext: dataSourcePostgresInstanceRead,
		Schema: map[string]*schema.Schema{
			"id": {
				Type:     schema.TypeString,
				Required: true,
			},
			"name": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"description": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"flavor_database_id": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"version": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"volume_type": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"volume_size": {
				Type:     schema.TypeInt,
				Computed: true,
			},
			"mode": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"replica": {
				Type:     schema.TypeInt,
				Computed: true,
			},
			"primary_ip": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"primary_port": {
				Type:     schema.TypeInt,
				Computed: true,
			},
			"standby_ip": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"standby_port": {
				Type:     schema.TypeInt,
				Computed: true,
			},
			"enable_tls": {
				Type:     schema.TypeBool,
				Computed: true,
			},
			"tls_mode": {
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

func dataSourcePostgresInstanceRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	cfg := meta.(*config.Config)
	id := d.Get("id").(string)

	resp := &dto.PostgresInstanceResponse{}
	_, err := cfg.Client.Get(ctx, client.ApiPath.DatabasePostgresInstanceWithID(cfg.ProjectID, id), resp, nil)
	if err != nil {
		return diag.Errorf("Error retrieving vnpaycloud_database_postgres_instance %s: %s", id, err)
	}

	inst := resp.PostgresInstance
	d.SetId(inst.ID)
	d.Set("name", inst.Name)
	d.Set("description", inst.Description)
	d.Set("flavor_database_id", inst.FlavorDatabaseID)
	d.Set("version", inst.Version)
	d.Set("volume_type", inst.VolumeType)
	d.Set("volume_size", int(inst.VolumeSize))
	d.Set("mode", inst.Mode)
	d.Set("replica", inst.Replica)
	d.Set("primary_ip", inst.PrimaryIP)
	d.Set("primary_port", inst.PrimaryPort)
	d.Set("standby_ip", inst.StandbyIP)
	d.Set("standby_port", inst.StandbyPort)
	d.Set("enable_tls", inst.EnableTls)
	d.Set("tls_mode", inst.TlsMode)
	d.Set("status", inst.Status)
	d.Set("created_at", inst.CreatedAt)

	return nil
}

func DataSourceDatabasePostgresInstances() *schema.Resource {
	return &schema.Resource{
		ReadContext: dataSourcePostgresInstancesRead,
		Schema: map[string]*schema.Schema{
			"postgres_instances": {
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
						"version": {
							Type:     schema.TypeString,
							Computed: true,
						},
						"mode": {
							Type:     schema.TypeString,
							Computed: true,
						},
						"replica": {
							Type:     schema.TypeInt,
							Computed: true,
						},
						"primary_ip": {
							Type:     schema.TypeString,
							Computed: true,
						},
						"primary_port": {
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

func dataSourcePostgresInstancesRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	cfg := meta.(*config.Config)

	resp := &dto.ListPostgresInstancesResponse{}
	_, err := cfg.Client.Get(ctx, client.ApiPath.DatabasePostgresInstances(cfg.ProjectID), resp, nil)
	if err != nil {
		return diag.Errorf("Error listing vnpaycloud_database_postgres_instances: %s", err)
	}

	var instances []map[string]interface{}
	for _, inst := range resp.PostgresInstances {
		instances = append(instances, map[string]interface{}{
			"id":           inst.ID,
			"name":         inst.Name,
			"version":      inst.Version,
			"mode":         inst.Mode,
			"replica":      inst.Replica,
			"primary_ip":   inst.PrimaryIP,
			"primary_port": inst.PrimaryPort,
			"status":       inst.Status,
			"created_at":   inst.CreatedAt,
		})
	}

	d.SetId("database_postgres_instances")
	d.Set("postgres_instances", instances)

	return nil
}
