package instance

import (
	"context"
	"fmt"
	"terraform-provider-vnpaycloud/vnpaycloud/config"
	"terraform-provider-vnpaycloud/vnpaycloud/dto"
	"terraform-provider-vnpaycloud/vnpaycloud/helper/client"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

func DataSourceInstance() *schema.Resource {
	return &schema.Resource{
		ReadContext: dataSourceInstanceRead,
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
			"image_name": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"image_id": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"flavor_name": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"volume_ids": {
				Type:     schema.TypeList,
				Computed: true,
				Elem:     &schema.Schema{Type: schema.TypeString},
			},
			"status": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"power_state": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"network_interface_ids": {
				Type:     schema.TypeList,
				Computed: true,
				Elem:     &schema.Schema{Type: schema.TypeString},
			},
			"key_pair": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"security_groups": {
				Type:     schema.TypeList,
				Computed: true,
				Elem:     &schema.Schema{Type: schema.TypeString},
			},
			"server_group_id": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"zone_id": {
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

func dataSourceInstanceRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	cfg := meta.(*config.Config)

	if id, ok := d.GetOk("id"); ok {
		instResp := &dto.InstanceResponse{}
		_, err := cfg.Client.Get(ctx, client.ApiPath.InstanceWithID(cfg.ProjectID, id.(string)), instResp, nil)
		if err != nil {
			return diag.Errorf("Error fetching vnpaycloud_instance %s: %s", id, err)
		}
		return setInstanceData(d, &instResp.Instance)
	}

	listResp := &dto.ListInstancesResponse{}
	_, err := cfg.Client.Get(ctx, client.ApiPath.Instances(cfg.ProjectID), listResp, nil)
	if err != nil {
		return diag.Errorf("Error listing vnpaycloud_instance: %s", err)
	}

	nameFilter, nameOk := d.GetOk("name")

	for _, inst := range listResp.Instances {
		if nameOk && inst.Name != nameFilter.(string) {
			continue
		}
		return setInstanceData(d, &inst)
	}

	return diag.Errorf("No vnpaycloud_instance found matching the criteria")
}

func setInstanceData(d *schema.ResourceData, inst *dto.Instance) diag.Diagnostics {
	d.SetId(inst.ID)
	d.Set("name", inst.Name)
	d.Set("image_name", inst.ImageName)
	d.Set("image_id", inst.ImageID)
	d.Set("flavor_name", inst.FlavorName)
	d.Set("volume_ids", inst.VolumeIDs)
	d.Set("status", inst.Status)
	d.Set("power_state", inst.PowerState)
	d.Set("network_interface_ids", inst.NetworkInterfaceIDs)
	d.Set("key_pair", inst.KeyPairID)
	d.Set("security_groups", inst.SecurityGroupIDs)
	d.Set("server_group_id", inst.ServerGroupID)
	d.Set("zone_id", inst.ZoneID)
	d.Set("created_at", inst.CreatedAt)
	return nil
}

func DataSourceInstances() *schema.Resource {
	return &schema.Resource{
		ReadContext: dataSourceInstancesRead,
		Schema: map[string]*schema.Schema{
			"instances": {
				Type:     schema.TypeList,
				Computed: true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"id":              {Type: schema.TypeString, Computed: true},
						"name":            {Type: schema.TypeString, Computed: true},
						"image_name":      {Type: schema.TypeString, Computed: true},
						"flavor_name":     {Type: schema.TypeString, Computed: true},
						"status":          {Type: schema.TypeString, Computed: true},
						"power_state":     {Type: schema.TypeString, Computed: true},
						"key_pair":        {Type: schema.TypeString, Computed: true},
						"server_group_id": {Type: schema.TypeString, Computed: true},
						"zone_id":         {Type: schema.TypeString, Computed: true},
						"created_at":      {Type: schema.TypeString, Computed: true},
					},
				},
			},
		},
	}
}

func dataSourceInstancesRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	cfg := meta.(*config.Config)

	listResp := &dto.ListInstancesResponse{}
	_, err := cfg.Client.Get(ctx, client.ApiPath.Instances(cfg.ProjectID), listResp, nil)
	if err != nil {
		return diag.Errorf("Error listing vnpaycloud_instances: %s", err)
	}

	var instances []map[string]interface{}
	for _, inst := range listResp.Instances {
		instances = append(instances, map[string]interface{}{
			"id":              inst.ID,
			"name":            inst.Name,
			"image_name":      inst.ImageName,
			"flavor_name":     inst.FlavorName,
			"status":          inst.Status,
			"power_state":     inst.PowerState,
			"key_pair":        inst.KeyPairID,
			"server_group_id": inst.ServerGroupID,
			"zone_id":         inst.ZoneID,
			"created_at":      inst.CreatedAt,
		})
	}

	d.SetId(fmt.Sprintf("instances-%s", cfg.ProjectID))
	d.Set("instances", instances)

	return nil
}
