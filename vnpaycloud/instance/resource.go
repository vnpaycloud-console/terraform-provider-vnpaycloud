package instance

import (
	"context"
	"fmt"
	"terraform-provider-vnpaycloud/vnpaycloud/config"
	"terraform-provider-vnpaycloud/vnpaycloud/dto"
	"terraform-provider-vnpaycloud/vnpaycloud/helper/client"
	"terraform-provider-vnpaycloud/vnpaycloud/util"
	"time"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/retry"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
)

func ResourceInstance() *schema.Resource {
	return &schema.Resource{
		CreateContext: resourceInstanceCreate,
		ReadContext:   resourceInstanceRead,
		UpdateContext: resourceInstanceUpdate,
		DeleteContext: resourceInstanceDelete,
		CustomizeDiff: validateInstanceDiff,
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
				Type:         schema.TypeString,
				Optional:     true,
				ForceNew:     true,
				ExactlyOneOf: []string{"image", "snapshot_id", "restore_point_id"},
			},
			"snapshot_id": {
				Type:         schema.TypeString,
				Optional:     true,
				ForceNew:     true,
				ExactlyOneOf: []string{"image", "snapshot_id", "restore_point_id"},
			},
			"restore_point_id": {
				Type:         schema.TypeString,
				Optional:     true,
				ForceNew:     true,
				ExactlyOneOf: []string{"image", "snapshot_id", "restore_point_id"},
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
				Type:         schema.TypeInt,
				Optional:     true,
				ForceNew:     true,
				ValidateFunc: validation.IntAtLeast(20),
			},
			"root_disk_type": {
				Type:     schema.TypeString,
				Optional: true,
				ForceNew: true,
			},
			"key_pair": {
				Type:     schema.TypeString,
				Optional: true,
				Computed: true,
				ForceNew: true,
			},
			"network_interface_ids": {
				Type:     schema.TypeList,
				Optional: true,
				Computed: true,
				ForceNew: true,
				Elem:     &schema.Schema{Type: schema.TypeString},
			},
			"security_groups": {
				Type:     schema.TypeList,
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

func validateInstanceDiff(ctx context.Context, d *schema.ResourceDiff, meta interface{}) error {
	if d.Get("is_custom_flavor").(bool) {
		return fmt.Errorf("custom flavor is not supported: set is_custom_flavor = false and specify a named 'flavor'")
	}
	if d.Get("flavor").(string) == "" {
		return fmt.Errorf("'flavor' is required: specify a named flavor (custom flavor is not supported)")
	}

	if d.Get("image").(string) != "" {
		if d.Get("root_disk_gb").(int) < 20 {
			return fmt.Errorf("'root_disk_gb' is required and must be at least 20 when creating from 'image'")
		}
		if d.Get("root_disk_type").(string) == "" {
			return fmt.Errorf("'root_disk_type' is required when creating from 'image'")
		}
	}

	if d.Id() == "" {
		if raw := d.GetRawConfig(); raw.IsKnown() && !raw.IsNull() {
			ni := raw.GetAttr("network_interface_ids")
			if ni.IsNull() || (ni.IsKnown() && ni.LengthInt() == 0) {
				return fmt.Errorf("network_interface_ids: at least one network interface is required when creating an instance")
			}
		}
	}
	return nil
}

func resourceInstanceCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	cfg := meta.(*config.Config)

	if d.Get("is_custom_flavor").(bool) {
		return diag.Errorf("custom flavor is not supported: set is_custom_flavor = false and specify a named 'flavor'")
	}
	if d.Get("flavor").(string) == "" {
		return diag.Errorf("'flavor' is required: specify a named flavor (custom flavor is not supported)")
	}

	createOpts := dto.CreateInstanceRequest{
		Name:               d.Get("name").(string),
		Image:              d.Get("image").(string),
		SnapshotID:         d.Get("snapshot_id").(string),
		RestorePointID:     d.Get("restore_point_id").(string),
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

	inst := instResp.Instance
	d.Set("name", inst.Name)
	d.Set("image_name", inst.ImageName)
	d.Set("image_id", inst.ImageID)
	d.Set("flavor_name", inst.FlavorName)
	d.Set("volume_ids", inst.VolumeIDs)
	d.Set("status", inst.Status)
	d.Set("power_state", inst.PowerState)
	if inst.KeyPairName != "" {
		d.Set("key_pair", inst.KeyPairName)
	}
	d.Set("security_groups", inst.SecurityGroupIDs)
	d.Set("server_group_id", inst.ServerGroupID)
	d.Set("zone_id", inst.ZoneID)
	d.Set("created_at", inst.CreatedAt)

	if d.Get("snapshot_id").(string) == "" && d.Get("restore_point_id").(string) == "" {
		if v, ok := d.GetOk("image"); !ok || v.(string) == "" {
			if inst.ImageName != "" {
				d.Set("image", inst.ImageName)
			}
		}
		if v, ok := d.GetOk("root_disk_gb"); !ok || v.(int) == 0 {
			if inst.RootDiskGB > 0 {
				d.Set("root_disk_gb", int(inst.RootDiskGB))
			}
		}
		if v, ok := d.GetOk("root_disk_type"); !ok || v.(string) == "" {
			if inst.RootDiskVolumeType != "" {
				d.Set("root_disk_type", inst.RootDiskVolumeType)
			}
		}
	}

	if v, ok := d.GetOk("flavor"); !ok || v.(string) == "" {
		if inst.FlavorName != "" {
			d.Set("flavor", inst.FlavorName)
		}
	}
	if v, ok := d.GetOk("network_interface_ids"); !ok || len(v.([]interface{})) == 0 {
		if len(inst.NetworkInterfaceIDs) > 0 {
			d.Set("network_interface_ids", inst.NetworkInterfaceIDs)
		}
	}

	return nil
}

func resourceInstanceUpdate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	if diags := resourceInstanceUpdateInner(ctx, d, meta); diags.HasError() {
		return append(resourceInstanceRead(ctx, d, meta), diags...)
	} else {
		return diags
	}
}

func resourceInstanceUpdateInner(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	cfg := meta.(*config.Config)

	if d.HasChange("name") {
		updateOpts := dto.UpdateInstanceRequest{
			Name: d.Get("name").(string),
		}
		_, err := cfg.Client.Put(ctx, client.ApiPath.InstanceWithID(cfg.ProjectID, d.Id()), updateOpts, nil, nil)
		if err != nil {
			return diag.Errorf("Error updating vnpaycloud_instance %s: %s", d.Id(), err)
		}
	}

	if d.HasChanges("flavor", "custom_vcpus", "custom_ram_mb") {
		if d.Get("is_custom_flavor").(bool) {
			return diag.Errorf("custom flavor is not supported: set is_custom_flavor = false and specify a named 'flavor'")
		}
		if d.Get("flavor").(string) == "" {
			return diag.Errorf("'flavor' is required: specify a named flavor (custom flavor is not supported)")
		}

		resizeOpts := dto.ResizeInstanceRequest{
			Flavor:         d.Get("flavor").(string),
			IsCustomFlavor: d.Get("is_custom_flavor").(bool),
			CustomVCPUs:    int32(d.Get("custom_vcpus").(int)),
			CustomRAMMB:    int32(d.Get("custom_ram_mb").(int)),
		}

		_, err := cfg.Client.Post(ctx, client.ApiPath.InstanceResize(cfg.ProjectID, d.Id()), resizeOpts, nil, nil)
		if err != nil {
			oldFlavor, _ := d.GetChange("flavor")
			d.Set("flavor", oldFlavor)
			return diag.Errorf("Error resizing vnpaycloud_instance %s: %s", d.Id(), err)
		}

		targetFlavor := d.Get("flavor").(string)
		var stateConf *retry.StateChangeConf
		if targetFlavor != "" {
			stateConf = &retry.StateChangeConf{
				Pending:    []string{"resizing"},
				Target:     []string{"done"},
				Refresh:    instanceFlavorRefreshFunc(ctx, cfg.Client, cfg.ProjectID, d.Id(), targetFlavor),
				Timeout:    d.Timeout(schema.TimeoutUpdate),
				Delay:      10 * time.Second,
				MinTimeout: 5 * time.Second,
			}
		} else {
			stateConf = &retry.StateChangeConf{
				Pending:    []string{"resizing", "verify_resize", "migrating"},
				Target:     []string{"active", "running"},
				Refresh:    instanceStateRefreshFunc(ctx, cfg.Client, cfg.ProjectID, d.Id()),
				Timeout:    d.Timeout(schema.TimeoutUpdate),
				Delay:      10 * time.Second,
				MinTimeout: 5 * time.Second,
			}
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
