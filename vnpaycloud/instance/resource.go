package instance

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

func ResourceInstance() *schema.Resource {
	return &schema.Resource{
		CreateContext: resourceInstanceCreate,
		ReadContext:   resourceInstanceRead,
		UpdateContext: resourceInstanceUpdate,
		DeleteContext: resourceInstanceDelete,
		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},
		Timeouts: &schema.ResourceTimeout{
			Create: schema.DefaultTimeout(30 * time.Minute),
			Update: schema.DefaultTimeout(30 * time.Minute),
			Delete: schema.DefaultTimeout(10 * time.Minute),
		},
		Schema: map[string]*schema.Schema{
			"name": {
				Type:     schema.TypeString,
				Required: true,
			},
			"image": {
				Type:     schema.TypeString,
				Optional: true,
				ForceNew: true,
			},
			"snapshot_id": {
				Type:     schema.TypeString,
				Optional: true,
				ForceNew: true,
			},
			"flavor": {
				Type:     schema.TypeString,
				Optional: true,
			},
			"is_custom_flavor": {
				Type:     schema.TypeBool,
				Optional: true,
			},
			"custom_vcpus": {
				Type:     schema.TypeInt,
				Optional: true,
			},
			"custom_ram_mb": {
				Type:     schema.TypeInt,
				Optional: true,
			},
			"root_disk_gb": {
				Type:     schema.TypeInt,
				Required: true,
				ForceNew: true,
			},
			"root_disk_type": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},
			"key_pair": {
				Type:     schema.TypeString,
				Optional: true,
				Computed: true,
				ForceNew: true,
			},
			"security_groups": {
				Type:     schema.TypeList,
				Optional: true,
				Computed: true,
				Elem:     &schema.Schema{Type: schema.TypeString},
			},
			"network_interface_ids": {
				Type:     schema.TypeList,
				Optional: true,
				Computed: true,
				Elem:     &schema.Schema{Type: schema.TypeString},
			},
			"server_group_id": {
				Type:     schema.TypeString,
				Optional: true,
				ForceNew: true,
			},
			"user_data": {
				Type:      schema.TypeString,
				Optional:  true,
				ForceNew:  true,
				Sensitive: true,
			},
			"is_user_data_base64": {
				Type:     schema.TypeBool,
				Optional: true,
				ForceNew: true,
			},
			// Computed attributes
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

func resourceInstanceCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	cfg := meta.(*config.Config)

	createOpts := dto.CreateInstanceRequest{
		Name:               d.Get("name").(string),
		Image:              d.Get("image").(string),
		SnapshotID:         d.Get("snapshot_id").(string),
		Flavor:             d.Get("flavor").(string),
		RootDiskGB:         int32(d.Get("root_disk_gb").(int)),
		RootDiskVolumeType: d.Get("root_disk_type").(string),
		KeyPair:            d.Get("key_pair").(string),
		ServerGroupID:      d.Get("server_group_id").(string),
		UserData:           d.Get("user_data").(string),
		IsUserDataBase64:   d.Get("is_user_data_base64").(bool),
	}

	if v, ok := d.GetOk("network_interface_ids"); ok {
		niList := v.([]interface{})
		nids := make([]string, len(niList))
		for i, ni := range niList {
			nids[i] = ni.(string)
		}
		createOpts.NetworkInterfaceIDs = nids
	}

	tflog.Debug(ctx, "vnpaycloud_instance create options", map[string]interface{}{"create_opts": createOpts})

	createResp := &dto.InstanceResponse{}
	_, err := cfg.Client.Post(ctx, client.ApiPath.Instances(cfg.ProjectID), createOpts, createResp, nil)
	if err != nil {
		return diag.Errorf("Error creating vnpaycloud_instance: %s", err)
	}

	d.SetId(createResp.Instance.ID)

	stateConf := &retry.StateChangeConf{
		Pending:    []string{"initiating", "creating", "build", "building", "unknown"},
		Target:     []string{"active", "running"},
		Refresh:    instanceStateRefreshFunc(ctx, cfg.Client, cfg.ProjectID, createResp.Instance.ID),
		Timeout:    d.Timeout(schema.TimeoutCreate),
		Delay:      10 * time.Second,
		MinTimeout: 5 * time.Second,
	}

	_, err = stateConf.WaitForStateContext(ctx)
	if err != nil {
		return diag.Errorf("Error waiting for vnpaycloud_instance %s to become ready: %s", createResp.Instance.ID, err)
	}

	return resourceInstanceRead(ctx, d, meta)
}

func resourceInstanceRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	cfg := meta.(*config.Config)

	instResp := &dto.InstanceResponse{}
	_, err := cfg.Client.Get(ctx, client.ApiPath.InstanceWithID(cfg.ProjectID, d.Id()), instResp, nil)
	if err != nil {
		return diag.FromErr(util.CheckNotFound(d, err, "Error retrieving vnpaycloud_instance"))
	}

	tflog.Debug(ctx, "Retrieved vnpaycloud_instance "+d.Id(), map[string]interface{}{"instance": instResp.Instance})

	inst := instResp.Instance
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

func resourceInstanceUpdate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	cfg := meta.(*config.Config)

	// Update name and/or security_groups
	if d.HasChanges("name", "security_groups") {
		updateOpts := dto.UpdateInstanceRequest{}

		if d.HasChange("name") {
			updateOpts.Name = d.Get("name").(string)
		}

		if d.HasChange("security_groups") {
			sgList := d.Get("security_groups").([]interface{})
			sgs := make([]string, len(sgList))
			for i, sg := range sgList {
				sgs[i] = sg.(string)
			}
			updateOpts.SecurityGroups = sgs
		}

		tflog.Debug(ctx, "vnpaycloud_instance update options", map[string]interface{}{"update_opts": updateOpts})

		_, err := cfg.Client.Put(ctx, client.ApiPath.InstanceWithID(cfg.ProjectID, d.Id()), updateOpts, nil, nil)
		if err != nil {
			return diag.Errorf("Error updating vnpaycloud_instance %s: %s", d.Id(), err)
		}
	}

	// Resize (flavor change)
	if d.HasChanges("flavor", "custom_vcpus", "custom_ram_mb") {
		resizeOpts := dto.ResizeInstanceRequest{
			Flavor:         d.Get("flavor").(string),
			IsCustomFlavor: d.Get("is_custom_flavor").(bool),
			CustomVCPUs:    int32(d.Get("custom_vcpus").(int)),
			CustomRAMMB:    int32(d.Get("custom_ram_mb").(int)),
		}

		tflog.Debug(ctx, "vnpaycloud_instance resize options", map[string]interface{}{"resize_opts": resizeOpts})

		_, err := cfg.Client.Post(ctx, client.ApiPath.InstanceResize(cfg.ProjectID, d.Id()), resizeOpts, nil, nil)
		if err != nil {
			return diag.Errorf("Error resizing vnpaycloud_instance %s: %s", d.Id(), err)
		}

		stateConf := &retry.StateChangeConf{
			Pending:    []string{"resizing", "verify_resize", "migrating"},
			Target:     []string{"active", "running"},
			Refresh:    instanceStateRefreshFunc(ctx, cfg.Client, cfg.ProjectID, d.Id()),
			Timeout:    d.Timeout(schema.TimeoutUpdate),
			Delay:      10 * time.Second,
			MinTimeout: 5 * time.Second,
		}

		_, err = stateConf.WaitForStateContext(ctx)
		if err != nil {
			return diag.Errorf("Error waiting for vnpaycloud_instance %s to finish resizing: %s", d.Id(), err)
		}
	}

	return resourceInstanceRead(ctx, d, meta)
}

func resourceInstanceDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	cfg := meta.(*config.Config)

	instResp := &dto.InstanceResponse{}
	_, err := cfg.Client.Get(ctx, client.ApiPath.InstanceWithID(cfg.ProjectID, d.Id()), instResp, nil)
	if err != nil {
		return diag.FromErr(util.CheckDeleted(d, err, "Error retrieving vnpaycloud_instance"))
	}

	if instResp.Instance.Status != "deleting" {
		if _, err := cfg.Client.Delete(ctx, client.ApiPath.InstanceWithID(cfg.ProjectID, d.Id()), nil); err != nil {
			return diag.FromErr(util.CheckDeleted(d, err, "Error deleting vnpaycloud_instance"))
		}
	}

	stateConf := &retry.StateChangeConf{
		Pending:    []string{"deleting", "active", "running", "shutoff", "stopped"},
		Target:     []string{"deleted"},
		Refresh:    instanceStateRefreshFunc(ctx, cfg.Client, cfg.ProjectID, d.Id()),
		Timeout:    d.Timeout(schema.TimeoutDelete),
		Delay:      10 * time.Second,
		MinTimeout: 5 * time.Second,
	}

	_, err = stateConf.WaitForStateContext(ctx)
	if err != nil {
		return diag.Errorf("Error waiting for vnpaycloud_instance %s to delete: %s", d.Id(), err)
	}

	return nil
}
