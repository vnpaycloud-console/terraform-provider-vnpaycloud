package databaseredissentinel

import (
	"context"
	"fmt"
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

func ResourceDatabaseRedisSentinelInstance() *schema.Resource {
	return &schema.Resource{
		CreateContext: resourceRedisSentinelInstanceCreate,
		ReadContext:   resourceRedisSentinelInstanceRead,
		UpdateContext: resourceRedisSentinelInstanceUpdate,
		DeleteContext: resourceRedisSentinelInstanceDelete,
		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},
		CustomizeDiff: func(ctx context.Context, d *schema.ResourceDiff, meta interface{}) error {
			if d.Get("enable_tls").(bool) {
				return nil
			}
			if v, ok := d.GetOk("certificate_id"); ok && v.(string) != "" {
				return fmt.Errorf("certificate_id can only be set when enable_tls = true")
			}
			return nil
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
				ValidateFunc: validation.IntAtLeast(2),
			},
			"purpose": {
				Type:     schema.TypeString,
				Optional: true,
				ForceNew: true,
			},

			// Sentinel config
			"sentinel_name": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},
			"sentinel_replica": {
				Type:         schema.TypeInt,
				Required:     true,
				ValidateFunc: validation.IntAtLeast(3),
			},
			"sentinel_flavor_database_id": {
				Type:     schema.TypeString,
				Required: true,
			},
			"sentinel_volume_size": {
				Type:         schema.TypeInt,
				Required:     true,
				ForceNew:     true,
				ValidateFunc: validation.IntAtLeast(1),
			},

			// TLS
			"enable_tls": {
				Type:     schema.TypeBool,
				Optional: true,
				Default:  false,
			},
			"certificate_id": {
				Type:     schema.TypeString,
				Optional: true,
			},

			// Auto expand volume
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

			"enable_read_only_endpoint": {
				Type:     schema.TypeBool,
				Optional: true,
				Default:  false,
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
			"standby_ip": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"standby_port": {
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

func resourceRedisSentinelInstanceCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	cfg := meta.(*config.Config)

	createOpts := dto.CreateRedisSentinelInstanceRequest{
		Name:                     d.Get("name").(string),
		FlavorDatabaseID:         d.Get("flavor_database_id").(string),
		Version:                  d.Get("version").(string),
		VolumeType:               d.Get("volume_type").(string),
		VolumeSize:               int64(d.Get("volume_size").(int)),
		Replica:                  d.Get("replica").(int),
		SentinelName:             d.Get("sentinel_name").(string),
		SentinelReplica:          d.Get("sentinel_replica").(int),
		SentinelFlavorDatabaseID: d.Get("sentinel_flavor_database_id").(string),
		SentinelVolumeSize:       int64(d.Get("sentinel_volume_size").(int)),
		ZoneID:                   cfg.ZoneID,
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

	tflog.Debug(ctx, "vnpaycloud_database_redis_sentinel_instance create", map[string]interface{}{"opts": createOpts})

	resp := &dto.RedisSentinelInstanceResponse{}
	_, err := cfg.Client.Post(ctx, client.ApiPath.DatabaseRedisSentinelInstances(cfg.ProjectID), createOpts, resp, nil)
	if err != nil {
		return diag.Errorf("Error creating vnpaycloud_database_redis_sentinel_instance: %s", err)
	}

	d.SetId(resp.RedisSentinelInstance.ID)

	stateConf := &retry.StateChangeConf{
		Pending:    []string{"creating", "pending_create", "unknown"},
		Target:     []string{"active"},
		Refresh:    redisSentinelInstanceStateRefreshFunc(ctx, cfg.Client, cfg.ProjectID, resp.RedisSentinelInstance.ID),
		Timeout:    d.Timeout(schema.TimeoutCreate),
		Delay:      30 * time.Second,
		MinTimeout: 15 * time.Second,
	}

	_, err = stateConf.WaitForStateContext(ctx)
	if err != nil {
		return diag.Errorf("Error waiting for vnpaycloud_database_redis_sentinel_instance %s to become active: %s", d.Id(), err)
	}

	if d.Get("is_auto_expand_volume").(bool) {
		autoExpandOpts := dto.EnableAutoExpandVolumeRedisSentinelInstanceRequest{
			UsageThreshold: d.Get("usage_threshold").(int),
			ScalePercent:   d.Get("scale_percent").(int),
		}
		_, err := cfg.Client.Post(ctx, client.ApiPath.DatabaseRedisSentinelInstanceEnableAutoExpandVolume(cfg.ProjectID, d.Id()), autoExpandOpts, nil, nil)
		if err != nil {
			return diag.Errorf("Error enabling auto-expand volume for vnpaycloud_database_redis_sentinel_instance %s: %s", d.Id(), err)
		}
		if diags := waitForRedisSentinelActive(ctx, d, cfg); diags != nil {
			return diags
		}
	}

	// Read-only endpoint is not part of the create request; enable it after the instance is active.
	if d.Get("enable_read_only_endpoint").(bool) {
		_, err := cfg.Client.Post(ctx, client.ApiPath.DatabaseRedisSentinelInstanceEnableReadOnlyEndpoint(cfg.ProjectID, d.Id()), nil, nil, nil)
		if err != nil {
			return diag.Errorf("Error enabling read-only endpoint for vnpaycloud_database_redis_sentinel_instance %s: %s", d.Id(), err)
		}
		if diags := waitForRedisSentinelActive(ctx, d, cfg); diags != nil {
			return diags
		}
	}

	return resourceRedisSentinelInstanceRead(ctx, d, meta)
}

func resourceRedisSentinelInstanceRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	cfg := meta.(*config.Config)

	resp := &dto.RedisSentinelInstanceResponse{}
	_, err := cfg.Client.Get(ctx, client.ApiPath.DatabaseRedisSentinelInstanceWithID(cfg.ProjectID, d.Id()), resp, nil)
	if err != nil {
		return diag.FromErr(util.CheckNotFound(d, err, "vnpaycloud_database_redis_sentinel_instance"))
	}

	inst := resp.RedisSentinelInstance
	d.Set("name", inst.Name)
	d.Set("description", inst.Description)
	d.Set("flavor_database_id", inst.FlavorDatabaseID)
	d.Set("version", inst.Version)
	d.Set("volume_type", inst.VolumeType)
	d.Set("volume_size", int(inst.VolumeSize))
	d.Set("replica", inst.Replica)
	d.Set("purpose", inst.Purpose)
	d.Set("sentinel_name", inst.SentinelName)
	d.Set("sentinel_replica", inst.SentinelReplica)
	d.Set("sentinel_flavor_database_id", inst.SentinelFlavorDatabaseID)
	d.Set("sentinel_volume_size", int(inst.SentinelVolumeSize))
	d.Set("enable_tls", inst.EnableTls)
	d.Set("certificate_id", inst.CertificateID)
	d.Set("enable_read_only_endpoint", inst.EnableReadOnlyEndpoint)
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
	d.Set("standby_ip", inst.StandbyIP)
	d.Set("standby_port", inst.StandbyPort)
	d.Set("status", inst.Status)
	d.Set("created_at", inst.CreatedAt)

	return nil
}

func resourceRedisSentinelInstanceUpdate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	cfg := meta.(*config.Config)

	if d.HasChange("replica") {
		opts := dto.ScaleRedisSentinelInstanceRequest{
			Replica: d.Get("replica").(int),
		}
		_, err := cfg.Client.Post(ctx, client.ApiPath.DatabaseRedisSentinelInstanceScale(cfg.ProjectID, d.Id()), opts, nil, nil)
		if err != nil {
			return diag.Errorf("Error scaling vnpaycloud_database_redis_sentinel_instance %s: %s", d.Id(), err)
		}
		if diags := waitForRedisSentinelActive(ctx, d, cfg); diags != nil {
			return diags
		}
	}

	if d.HasChange("flavor_database_id") {
		opts := dto.ChangeFlavorRedisSentinelInstanceRequest{
			FlavorDatabaseID: d.Get("flavor_database_id").(string),
		}
		_, err := cfg.Client.Post(ctx, client.ApiPath.DatabaseRedisSentinelInstanceChangeFlavor(cfg.ProjectID, d.Id()), opts, nil, nil)
		if err != nil {
			return diag.Errorf("Error changing flavor for vnpaycloud_database_redis_sentinel_instance %s: %s", d.Id(), err)
		}
		if diags := waitForRedisSentinelActive(ctx, d, cfg); diags != nil {
			return diags
		}
	}

	if d.HasChange("volume_size") {
		opts := dto.ExpandVolumeRedisSentinelInstanceRequest{
			VolumeSize: int64(d.Get("volume_size").(int)),
		}
		_, err := cfg.Client.Post(ctx, client.ApiPath.DatabaseRedisSentinelInstanceExpandVolume(cfg.ProjectID, d.Id()), opts, nil, nil)
		if err != nil {
			return diag.Errorf("Error expanding volume for vnpaycloud_database_redis_sentinel_instance %s: %s", d.Id(), err)
		}
		if diags := waitForRedisSentinelActive(ctx, d, cfg); diags != nil {
			return diags
		}
	}

	if d.HasChange("sentinel_replica") {
		opts := dto.ScaleRedisSentinelRequest{
			SentinelReplica: d.Get("sentinel_replica").(int),
		}
		_, err := cfg.Client.Post(ctx, client.ApiPath.DatabaseRedisSentinelInstanceSentinelScale(cfg.ProjectID, d.Id()), opts, nil, nil)
		if err != nil {
			return diag.Errorf("Error scaling sentinel for vnpaycloud_database_redis_sentinel_instance %s: %s", d.Id(), err)
		}
		if diags := waitForRedisSentinelActive(ctx, d, cfg); diags != nil {
			return diags
		}
	}

	if d.HasChange("sentinel_flavor_database_id") {
		opts := dto.ChangeFlavorRedisSentinelRequest{
			SentinelFlavorDatabaseID: d.Get("sentinel_flavor_database_id").(string),
		}
		_, err := cfg.Client.Post(ctx, client.ApiPath.DatabaseRedisSentinelInstanceSentinelChangeFlavor(cfg.ProjectID, d.Id()), opts, nil, nil)
		if err != nil {
			return diag.Errorf("Error changing sentinel flavor for vnpaycloud_database_redis_sentinel_instance %s: %s", d.Id(), err)
		}
		if diags := waitForRedisSentinelActive(ctx, d, cfg); diags != nil {
			return diags
		}
	}

	if d.HasChange("is_auto_expand_volume") || d.HasChange("usage_threshold") || d.HasChange("scale_percent") {
		if d.Get("is_auto_expand_volume").(bool) {
			old, _ := d.GetChange("is_auto_expand_volume")
			if old.(bool) {
				_, err := cfg.Client.Post(ctx, client.ApiPath.DatabaseRedisSentinelInstanceDisableAutoExpandVolume(cfg.ProjectID, d.Id()), nil, nil, nil)
				if err != nil {
					return diag.Errorf("Error disabling auto-expand volume for vnpaycloud_database_redis_sentinel_instance %s: %s", d.Id(), err)
				}
				if diags := waitForRedisSentinelActive(ctx, d, cfg); diags != nil {
					return diags
				}
			}
			opts := dto.EnableAutoExpandVolumeRedisSentinelInstanceRequest{
				UsageThreshold: d.Get("usage_threshold").(int),
				ScalePercent:   d.Get("scale_percent").(int),
			}
			_, err := cfg.Client.Post(ctx, client.ApiPath.DatabaseRedisSentinelInstanceEnableAutoExpandVolume(cfg.ProjectID, d.Id()), opts, nil, nil)
			if err != nil {
				return diag.Errorf("Error enabling auto-expand volume for vnpaycloud_database_redis_sentinel_instance %s: %s", d.Id(), err)
			}
			if diags := waitForRedisSentinelActive(ctx, d, cfg); diags != nil {
				return diags
			}
		} else if d.HasChange("is_auto_expand_volume") {
			old, _ := d.GetChange("is_auto_expand_volume")
			if old.(bool) {
				_, err := cfg.Client.Post(ctx, client.ApiPath.DatabaseRedisSentinelInstanceDisableAutoExpandVolume(cfg.ProjectID, d.Id()), nil, nil, nil)
				if err != nil {
					return diag.Errorf("Error disabling auto-expand volume for vnpaycloud_database_redis_sentinel_instance %s: %s", d.Id(), err)
				}
				if diags := waitForRedisSentinelActive(ctx, d, cfg); diags != nil {
					return diags
				}
			}
		}
	}

	if d.HasChange("enable_tls") || d.HasChange("certificate_id") {
		if d.Get("enable_tls").(bool) {
			old, _ := d.GetChange("enable_tls")
			if old.(bool) {
				_, err := cfg.Client.Post(ctx, client.ApiPath.DatabaseRedisSentinelInstanceDisableTls(cfg.ProjectID, d.Id()), nil, nil, nil)
				if err != nil {
					return diag.Errorf("Error disabling TLS for vnpaycloud_database_redis_sentinel_instance %s: %s", d.Id(), err)
				}
				if diags := waitForRedisSentinelActive(ctx, d, cfg); diags != nil {
					return diags
				}
			}
			opts := dto.EnableTlsRedisSentinelInstanceRequest{
				CertificateID: d.Get("certificate_id").(string),
			}
			_, err := cfg.Client.Post(ctx, client.ApiPath.DatabaseRedisSentinelInstanceEnableTls(cfg.ProjectID, d.Id()), opts, nil, nil)
			if err != nil {
				return diag.Errorf("Error enabling TLS for vnpaycloud_database_redis_sentinel_instance %s: %s", d.Id(), err)
			}
		} else {
			_, err := cfg.Client.Post(ctx, client.ApiPath.DatabaseRedisSentinelInstanceDisableTls(cfg.ProjectID, d.Id()), nil, nil, nil)
			if err != nil {
				return diag.Errorf("Error disabling TLS for vnpaycloud_database_redis_sentinel_instance %s: %s", d.Id(), err)
			}
		}
		if diags := waitForRedisSentinelActive(ctx, d, cfg); diags != nil {
			return diags
		}
	}

	if d.HasChange("enable_read_only_endpoint") {
		var err error
		if d.Get("enable_read_only_endpoint").(bool) {
			_, err = cfg.Client.Post(ctx, client.ApiPath.DatabaseRedisSentinelInstanceEnableReadOnlyEndpoint(cfg.ProjectID, d.Id()), nil, nil, nil)
		} else {
			_, err = cfg.Client.Post(ctx, client.ApiPath.DatabaseRedisSentinelInstanceDisableReadOnlyEndpoint(cfg.ProjectID, d.Id()), nil, nil, nil)
		}
		if err != nil {
			return diag.Errorf("Error updating read-only endpoint for vnpaycloud_database_redis_sentinel_instance %s: %s", d.Id(), err)
		}
		if diags := waitForRedisSentinelActive(ctx, d, cfg); diags != nil {
			return diags
		}
	}

	return resourceRedisSentinelInstanceRead(ctx, d, meta)
}

func resourceRedisSentinelInstanceDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	cfg := meta.(*config.Config)

	_, err := cfg.Client.Delete(ctx, client.ApiPath.DatabaseRedisSentinelInstanceWithID(cfg.ProjectID, d.Id()), nil)
	if err != nil {
		return diag.FromErr(util.CheckDeleted(d, err, "vnpaycloud_database_redis_sentinel_instance"))
	}

	stateConf := &retry.StateChangeConf{
		Pending:    []string{"deleting", "active", "unknown"},
		Target:     []string{"deleted"},
		Refresh:    redisSentinelInstanceStateRefreshFunc(ctx, cfg.Client, cfg.ProjectID, d.Id()),
		Timeout:    d.Timeout(schema.TimeoutDelete),
		Delay:      15 * time.Second,
		MinTimeout: 10 * time.Second,
	}

	_, err = stateConf.WaitForStateContext(ctx)
	if err != nil {
		return diag.Errorf("Error waiting for vnpaycloud_database_redis_sentinel_instance %s to delete: %s", d.Id(), err)
	}

	return nil
}

func waitForRedisSentinelActive(ctx context.Context, d *schema.ResourceData, cfg *config.Config) diag.Diagnostics {
	stateConf := &retry.StateChangeConf{
		Pending:    []string{"creating", "scaling", "changing_flavor", "expanding_volume", "updating", "pending_update", "unknown"},
		Target:     []string{"active"},
		Refresh:    redisSentinelInstanceStateRefreshFunc(ctx, cfg.Client, cfg.ProjectID, d.Id()),
		Timeout:    d.Timeout(schema.TimeoutUpdate),
		Delay:      15 * time.Second,
		MinTimeout: 10 * time.Second,
	}

	_, err := stateConf.WaitForStateContext(ctx)
	if err != nil {
		return diag.Errorf("Error waiting for vnpaycloud_database_redis_sentinel_instance %s to become active: %s", d.Id(), err)
	}
	return nil
}
