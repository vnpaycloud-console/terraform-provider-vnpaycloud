package databaseredis

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
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
)

func ResourceDatabaseRedisInstance() *schema.Resource {
	return &schema.Resource{
		CreateContext: resourceRedisInstanceCreate,
		ReadContext:   resourceRedisInstanceRead,
		UpdateContext: resourceRedisInstanceUpdate,
		DeleteContext: resourceRedisInstanceDelete,
		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},
		Timeouts: &schema.ResourceTimeout{
			Create: schema.DefaultTimeout(30 * time.Minute),
			Update: schema.DefaultTimeout(30 * time.Minute),
			Delete: schema.DefaultTimeout(15 * time.Minute),
		},
		Schema: map[string]*schema.Schema{
			"name": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},
			"description": {
				Type:     schema.TypeString,
				Optional: true,
				ForceNew: true,
			},
			"flavor_database_id": {
				Type:     schema.TypeString,
				Required: true,
			},
			"version": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},
			"volume_type": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},
			"volume_size": {
				Type:         schema.TypeInt,
				Required:     true,
				ValidateFunc: validation.IntAtLeast(1),
			},
			"replica": {
				Type:         schema.TypeInt,
				Required:     true,
				ForceNew:     true,
				ValidateFunc: validation.IntAtLeast(1),
			},
			"purpose": {
				Type:     schema.TypeString,
				Optional: true,
				ForceNew: true,
			},
			"enable_tls": {
				Type:     schema.TypeBool,
				Optional: true,
				Default:  false,
			},
			"certificate_id": {
				Type:     schema.TypeString,
				Optional: true,
			},
			"is_auto_expand_volume": {
				Type:     schema.TypeBool,
				Optional: true,
				Default:  false,
			},
			"usage_threshold": {
				Type:         schema.TypeInt,
				Optional:     true,
				ValidateFunc: validation.IntAtLeast(1),
			},
			"scale_percent": {
				Type:         schema.TypeInt,
				Optional:     true,
				ValidateFunc: validation.IntAtLeast(1),
			},

			// Computed
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
	}
}

func resourceRedisInstanceCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	cfg := meta.(*config.Config)

	createOpts := dto.CreateRedisInstanceRequest{
		Name:             d.Get("name").(string),
		FlavorDatabaseID: d.Get("flavor_database_id").(string),
		Version:          d.Get("version").(string),
		VolumeType:       d.Get("volume_type").(string),
		VolumeSize:       int64(d.Get("volume_size").(int)),
		Replica:          d.Get("replica").(int),
		ZoneID:           cfg.ZoneID,
	}

	if v, ok := d.GetOk("description"); ok {
		createOpts.Description = v.(string)
	}
	if v, ok := d.GetOk("purpose"); ok {
		createOpts.Purpose = v.(string)
	}
	if v, ok := d.GetOk("enable_tls"); ok && v.(bool) {
		createOpts.EnableTls = true
		if cm, ok := d.GetOk("certificate_id"); ok {
			createOpts.CertificateID = cm.(string)
		}
	}

	tflog.Debug(ctx, "vnpaycloud_database_redis_instance create", map[string]interface{}{"opts": createOpts})

	resp := &dto.RedisInstanceResponse{}
	_, err := cfg.Client.Post(ctx, client.ApiPath.DatabaseRedisInstances(cfg.ProjectID), createOpts, resp, nil)
	if err != nil {
		return diag.Errorf("Error creating vnpaycloud_database_redis_instance: %s", err)
	}

	d.SetId(resp.RedisInstance.ID)

	stateConf := &retry.StateChangeConf{
		Pending:    []string{"creating", "pending_create", "unknown"},
		Target:     []string{"active"},
		Refresh:    redisInstanceStateRefreshFunc(ctx, cfg.Client, cfg.ProjectID, resp.RedisInstance.ID),
		Timeout:    d.Timeout(schema.TimeoutCreate),
		Delay:      30 * time.Second,
		MinTimeout: 15 * time.Second,
	}

	_, err = stateConf.WaitForStateContext(ctx)
	if err != nil {
		return diag.Errorf("Error waiting for vnpaycloud_database_redis_instance %s to become active: %s", d.Id(), err)
	}

	if d.Get("is_auto_expand_volume").(bool) {
		autoExpandOpts := dto.EnableAutoExpandVolumeRedisInstanceRequest{
			UsageThreshold: d.Get("usage_threshold").(int),
			ScalePercent:   d.Get("scale_percent").(int),
		}
		_, err := cfg.Client.Post(ctx, client.ApiPath.DatabaseRedisInstanceEnableAutoExpandVolume(cfg.ProjectID, d.Id()), autoExpandOpts, nil, nil)
		if err != nil {
			return diag.Errorf("Error enabling auto-expand volume for vnpaycloud_database_redis_instance %s: %s", d.Id(), err)
		}
		if diags := waitForRedisActive(ctx, d, cfg); diags != nil {
			return diags
		}
	}

	return resourceRedisInstanceRead(ctx, d, meta)
}

func resourceRedisInstanceRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	cfg := meta.(*config.Config)

	resp := &dto.RedisInstanceResponse{}
	_, err := cfg.Client.Get(ctx, client.ApiPath.DatabaseRedisInstanceWithID(cfg.ProjectID, d.Id()), resp, nil)
	if err != nil {
		return diag.FromErr(util.CheckNotFound(d, err, "vnpaycloud_database_redis_instance"))
	}

	inst := resp.RedisInstance
	d.Set("name", inst.Name)
	d.Set("description", inst.Description)
	d.Set("flavor_database_id", inst.FlavorDatabaseID)
	d.Set("version", inst.Version)
	d.Set("volume_type", inst.VolumeType)
	d.Set("volume_size", int(inst.VolumeSize))
	d.Set("replica", inst.Replica)
	d.Set("purpose", inst.Purpose)
	d.Set("enable_tls", inst.EnableTls)
	d.Set("certificate_id", inst.CertificateID)
	d.Set("is_auto_expand_volume", inst.IsAutoExpandVolume)
	if inst.IsAutoExpandVolume {
		d.Set("usage_threshold", inst.UsageThreshold)
		d.Set("scale_percent", inst.ScalePercent)
	} else {
		d.Set("usage_threshold", 0)
		d.Set("scale_percent", 0)
	}
	d.Set("primary_ip", inst.PrimaryIP)
	d.Set("primary_port", inst.PrimaryPort)
	d.Set("status", inst.Status)
	d.Set("created_at", inst.CreatedAt)

	return nil
}

func resourceRedisInstanceUpdate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	cfg := meta.(*config.Config)

	if d.HasChange("flavor_database_id") {
		opts := dto.ChangeFlavorRedisInstanceRequest{
			FlavorDatabaseID: d.Get("flavor_database_id").(string),
		}
		_, err := cfg.Client.Post(ctx, client.ApiPath.DatabaseRedisInstanceChangeFlavor(cfg.ProjectID, d.Id()), opts, nil, nil)
		if err != nil {
			return diag.Errorf("Error changing flavor for vnpaycloud_database_redis_instance %s: %s", d.Id(), err)
		}
		if diags := waitForRedisActive(ctx, d, cfg); diags != nil {
			return diags
		}
	}

	if d.HasChange("volume_size") {
		opts := dto.ExpandVolumeRedisInstanceRequest{
			VolumeSize: int64(d.Get("volume_size").(int)),
		}
		_, err := cfg.Client.Post(ctx, client.ApiPath.DatabaseRedisInstanceExpandVolume(cfg.ProjectID, d.Id()), opts, nil, nil)
		if err != nil {
			return diag.Errorf("Error expanding volume for vnpaycloud_database_redis_instance %s: %s", d.Id(), err)
		}
		if diags := waitForRedisActive(ctx, d, cfg); diags != nil {
			return diags
		}
	}

	if d.HasChange("is_auto_expand_volume") || d.HasChange("usage_threshold") || d.HasChange("scale_percent") {
		if d.Get("is_auto_expand_volume").(bool) {
			old, _ := d.GetChange("is_auto_expand_volume")
			if old.(bool) {
				_, err := cfg.Client.Post(ctx, client.ApiPath.DatabaseRedisInstanceDisableAutoExpandVolume(cfg.ProjectID, d.Id()), nil, nil, nil)
				if err != nil {
					return diag.Errorf("Error disabling auto-expand volume for vnpaycloud_database_redis_instance %s: %s", d.Id(), err)
				}
				if diags := waitForRedisActive(ctx, d, cfg); diags != nil {
					return diags
				}
			}
			opts := dto.EnableAutoExpandVolumeRedisInstanceRequest{
				UsageThreshold: d.Get("usage_threshold").(int),
				ScalePercent:   d.Get("scale_percent").(int),
			}
			_, err := cfg.Client.Post(ctx, client.ApiPath.DatabaseRedisInstanceEnableAutoExpandVolume(cfg.ProjectID, d.Id()), opts, nil, nil)
			if err != nil {
				return diag.Errorf("Error enabling auto-expand volume for vnpaycloud_database_redis_instance %s: %s", d.Id(), err)
			}
			if diags := waitForRedisActive(ctx, d, cfg); diags != nil {
				return diags
			}
		} else if d.HasChange("is_auto_expand_volume") {
			old, _ := d.GetChange("is_auto_expand_volume")
			if old.(bool) {
				_, err := cfg.Client.Post(ctx, client.ApiPath.DatabaseRedisInstanceDisableAutoExpandVolume(cfg.ProjectID, d.Id()), nil, nil, nil)
				if err != nil {
					return diag.Errorf("Error disabling auto-expand volume for vnpaycloud_database_redis_instance %s: %s", d.Id(), err)
				}
				if diags := waitForRedisActive(ctx, d, cfg); diags != nil {
					return diags
				}
			}
		}
	}

	if d.HasChange("enable_tls") || d.HasChange("certificate_id") {
		if d.Get("enable_tls").(bool) {
			old, _ := d.GetChange("enable_tls")
			if old.(bool) {
				_, err := cfg.Client.Post(ctx, client.ApiPath.DatabaseRedisInstanceDisableTls(cfg.ProjectID, d.Id()), nil, nil, nil)
				if err != nil {
					return diag.Errorf("Error disabling TLS for vnpaycloud_database_redis_instance %s: %s", d.Id(), err)
				}
				if diags := waitForRedisActive(ctx, d, cfg); diags != nil {
					return diags
				}
			}
			opts := dto.EnableTlsRedisInstanceRequest{
				CertificateID: d.Get("certificate_id").(string),
			}
			_, err := cfg.Client.Post(ctx, client.ApiPath.DatabaseRedisInstanceEnableTls(cfg.ProjectID, d.Id()), opts, nil, nil)
			if err != nil {
				return diag.Errorf("Error enabling TLS for vnpaycloud_database_redis_instance %s: %s", d.Id(), err)
			}
		} else {
			_, err := cfg.Client.Post(ctx, client.ApiPath.DatabaseRedisInstanceDisableTls(cfg.ProjectID, d.Id()), nil, nil, nil)
			if err != nil {
				return diag.Errorf("Error disabling TLS for vnpaycloud_database_redis_instance %s: %s", d.Id(), err)
			}
		}
		if diags := waitForRedisActive(ctx, d, cfg); diags != nil {
			return diags
		}
	}

	return resourceRedisInstanceRead(ctx, d, meta)
}

func resourceRedisInstanceDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	cfg := meta.(*config.Config)

	_, err := cfg.Client.Delete(ctx, client.ApiPath.DatabaseRedisInstanceWithID(cfg.ProjectID, d.Id()), nil)
	if err != nil {
		return diag.FromErr(util.CheckDeleted(d, err, "vnpaycloud_database_redis_instance"))
	}

	stateConf := &retry.StateChangeConf{
		Pending:    []string{"deleting", "active", "unknown"},
		Target:     []string{"deleted"},
		Refresh:    redisInstanceStateRefreshFunc(ctx, cfg.Client, cfg.ProjectID, d.Id()),
		Timeout:    d.Timeout(schema.TimeoutDelete),
		Delay:      15 * time.Second,
		MinTimeout: 10 * time.Second,
	}

	_, err = stateConf.WaitForStateContext(ctx)
	if err != nil {
		return diag.Errorf("Error waiting for vnpaycloud_database_redis_instance %s to delete: %s", d.Id(), err)
	}

	return nil
}

func waitForRedisActive(ctx context.Context, d *schema.ResourceData, cfg *config.Config) diag.Diagnostics {
	stateConf := &retry.StateChangeConf{
		Pending:    []string{"creating", "scaling", "changing_flavor", "expanding_volume", "updating", "pending_update", "unknown"},
		Target:     []string{"active"},
		Refresh:    redisInstanceStateRefreshFunc(ctx, cfg.Client, cfg.ProjectID, d.Id()),
		Timeout:    d.Timeout(schema.TimeoutUpdate),
		Delay:      15 * time.Second,
		MinTimeout: 10 * time.Second,
	}

	_, err := stateConf.WaitForStateContext(ctx)
	if err != nil {
		return diag.Errorf("Error waiting for vnpaycloud_database_redis_instance %s to become active: %s", d.Id(), err)
	}
	return nil
}
