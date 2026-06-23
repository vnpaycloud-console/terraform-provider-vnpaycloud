package databasepostgres

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

func ResourceDatabasePostgresInstance() *schema.Resource {
	return &schema.Resource{
		CreateContext: resourcePostgresInstanceCreate,
		ReadContext:   resourcePostgresInstanceRead,
		UpdateContext: resourcePostgresInstanceUpdate,
		DeleteContext: resourcePostgresInstanceDelete,
		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},
		CustomizeDiff: func(ctx context.Context, d *schema.ResourceDiff, meta interface{}) error {
			// Scaling replica and the read-only endpoint are cluster-only operations.
			if d.Get("mode").(string) == "standalone" {
				if d.Get("enable_read_only_endpoint").(bool) {
					return fmt.Errorf("enable_read_only_endpoint is only supported when mode = cluster")
				}
				if d.Id() != "" && d.HasChange("replica") {
					return fmt.Errorf("replica can only be scaled when mode = cluster; standalone replica is fixed at create time")
				}
			}
			// certificate_id / tls_mode may only be set when TLS is enabled.
			if !d.Get("enable_tls").(bool) {
				if v, ok := d.GetOk("certificate_id"); ok && v.(string) != "" {
					return fmt.Errorf("certificate_id can only be set when enable_tls = true")
				}
				if v, ok := d.GetOk("tls_mode"); ok && v.(string) != "" {
					return fmt.Errorf("tls_mode can only be set when enable_tls = true")
				}
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
			"mode": {
				Type:         schema.TypeString,
				Required:     true,
				ForceNew:     true,
				ValidateFunc: validation.StringInSlice([]string{"standalone", "cluster"}, false),
			},
			"replica": {
				Type:         schema.TypeInt,
				Required:     true,
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
			"tls_mode": {
				Type:         schema.TypeString,
				Optional:     true,
				ValidateFunc: validation.StringInSlice([]string{"require", "verify-ca"}, false),
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

func resourcePostgresInstanceCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	cfg := meta.(*config.Config)

	createOpts := dto.CreatePostgresInstanceRequest{
		Name:             d.Get("name").(string),
		FlavorDatabaseID: d.Get("flavor_database_id").(string),
		Version:          d.Get("version").(string),
		VolumeType:       d.Get("volume_type").(string),
		VolumeSize:       int64(d.Get("volume_size").(int)),
		Mode:             d.Get("mode").(string),
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
		if tm, ok := d.GetOk("tls_mode"); ok {
			createOpts.TlsMode = tm.(string)
		}
	}

	tflog.Debug(ctx, "vnpaycloud_database_postgres_instance create", map[string]interface{}{"opts": createOpts})

	resp := &dto.PostgresInstanceResponse{}
	_, err := cfg.Client.Post(ctx, client.ApiPath.DatabasePostgresInstances(cfg.ProjectID), createOpts, resp, nil)
	if err != nil {
		return diag.Errorf("Error creating vnpaycloud_database_postgres_instance: %s", err)
	}

	d.SetId(resp.PostgresInstance.ID)

	stateConf := &retry.StateChangeConf{
		Pending:    []string{"creating", "pending_create", "unknown"},
		Target:     []string{"active"},
		Refresh:    postgresInstanceStateRefreshFunc(ctx, cfg.Client, cfg.ProjectID, resp.PostgresInstance.ID),
		Timeout:    d.Timeout(schema.TimeoutCreate),
		Delay:      30 * time.Second,
		MinTimeout: 15 * time.Second,
	}

	_, err = stateConf.WaitForStateContext(ctx)
	if err != nil {
		return diag.Errorf("Error waiting for vnpaycloud_database_postgres_instance %s to become active: %s", d.Id(), err)
	}

	// Enable auto-expand volume if configured
	if d.Get("is_auto_expand_volume").(bool) {
		autoExpandOpts := dto.EnableAutoExpandVolumePostgresInstanceRequest{
			UsageThreshold: d.Get("usage_threshold").(int),
			ScalePercent:   d.Get("scale_percent").(int),
		}
		_, err := cfg.Client.Post(ctx, client.ApiPath.DatabasePostgresInstanceEnableAutoExpandVolume(cfg.ProjectID, d.Id()), autoExpandOpts, nil, nil)
		if err != nil {
			return diag.Errorf("Error enabling auto-expand volume for vnpaycloud_database_postgres_instance %s: %s", d.Id(), err)
		}
		if diags := waitForPostgresActive(ctx, d, cfg); diags != nil {
			return diags
		}
	}

	// Read-only endpoint is not part of the create request; enable it after the instance is active.
	if d.Get("enable_read_only_endpoint").(bool) {
		_, err := cfg.Client.Post(ctx, client.ApiPath.DatabasePostgresInstanceEnableReadOnlyEndpoint(cfg.ProjectID, d.Id()), nil, nil, nil)
		if err != nil {
			return diag.Errorf("Error enabling read-only endpoint for vnpaycloud_database_postgres_instance %s: %s", d.Id(), err)
		}
		if diags := waitForPostgresActive(ctx, d, cfg); diags != nil {
			return diags
		}
	}

	return resourcePostgresInstanceRead(ctx, d, meta)
}

func resourcePostgresInstanceRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	cfg := meta.(*config.Config)

	resp := &dto.PostgresInstanceResponse{}
	_, err := cfg.Client.Get(ctx, client.ApiPath.DatabasePostgresInstanceWithID(cfg.ProjectID, d.Id()), resp, nil)
	if err != nil {
		return diag.FromErr(util.CheckNotFound(d, err, "vnpaycloud_database_postgres_instance"))
	}

	inst := resp.PostgresInstance
	d.Set("name", inst.Name)
	d.Set("description", inst.Description)
	d.Set("flavor_database_id", inst.FlavorDatabaseID)
	d.Set("version", inst.Version)
	d.Set("volume_type", inst.VolumeType)
	d.Set("volume_size", int(inst.VolumeSize))
	d.Set("mode", inst.Mode)
	d.Set("replica", inst.Replica)
	d.Set("purpose", inst.Purpose)
	d.Set("enable_tls", inst.EnableTls)
	d.Set("certificate_id", inst.CertificateID)
	d.Set("tls_mode", inst.TlsMode)
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

func resourcePostgresInstanceUpdate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	cfg := meta.(*config.Config)

	if d.HasChange("replica") {
		opts := dto.ScalePostgresInstanceRequest{
			Replica: d.Get("replica").(int),
		}
		_, err := cfg.Client.Post(ctx, client.ApiPath.DatabasePostgresInstanceScale(cfg.ProjectID, d.Id()), opts, nil, nil)
		if err != nil {
			return diag.Errorf("Error scaling vnpaycloud_database_postgres_instance %s: %s", d.Id(), err)
		}
		if diags := waitForPostgresActive(ctx, d, cfg); diags != nil {
			return diags
		}
	}

	if d.HasChange("flavor_database_id") {
		opts := dto.ChangeFlavorPostgresInstanceRequest{
			FlavorDatabaseID: d.Get("flavor_database_id").(string),
		}
		_, err := cfg.Client.Post(ctx, client.ApiPath.DatabasePostgresInstanceChangeFlavor(cfg.ProjectID, d.Id()), opts, nil, nil)
		if err != nil {
			return diag.Errorf("Error changing flavor for vnpaycloud_database_postgres_instance %s: %s", d.Id(), err)
		}
		if diags := waitForPostgresActive(ctx, d, cfg); diags != nil {
			return diags
		}
	}

	if d.HasChange("volume_size") {
		opts := dto.ExpandVolumePostgresInstanceRequest{
			VolumeSize: int64(d.Get("volume_size").(int)),
		}
		_, err := cfg.Client.Post(ctx, client.ApiPath.DatabasePostgresInstanceExpandVolume(cfg.ProjectID, d.Id()), opts, nil, nil)
		if err != nil {
			return diag.Errorf("Error expanding volume for vnpaycloud_database_postgres_instance %s: %s", d.Id(), err)
		}
		if diags := waitForPostgresActive(ctx, d, cfg); diags != nil {
			return diags
		}
	}

	if d.HasChange("is_auto_expand_volume") || d.HasChange("usage_threshold") || d.HasChange("scale_percent") {
		if d.Get("is_auto_expand_volume").(bool) {
			// If already enabled, disable first before re-enabling with new params
			old, _ := d.GetChange("is_auto_expand_volume")
			if old.(bool) {
				_, err := cfg.Client.Post(ctx, client.ApiPath.DatabasePostgresInstanceDisableAutoExpandVolume(cfg.ProjectID, d.Id()), nil, nil, nil)
				if err != nil {
					return diag.Errorf("Error disabling auto-expand volume for vnpaycloud_database_postgres_instance %s: %s", d.Id(), err)
				}
				if diags := waitForPostgresActive(ctx, d, cfg); diags != nil {
					return diags
				}
			}
			opts := dto.EnableAutoExpandVolumePostgresInstanceRequest{
				UsageThreshold: d.Get("usage_threshold").(int),
				ScalePercent:   d.Get("scale_percent").(int),
			}
			_, err := cfg.Client.Post(ctx, client.ApiPath.DatabasePostgresInstanceEnableAutoExpandVolume(cfg.ProjectID, d.Id()), opts, nil, nil)
			if err != nil {
				return diag.Errorf("Error enabling auto-expand volume for vnpaycloud_database_postgres_instance %s: %s", d.Id(), err)
			}
			if diags := waitForPostgresActive(ctx, d, cfg); diags != nil {
				return diags
			}
		} else if d.HasChange("is_auto_expand_volume") {
			old, _ := d.GetChange("is_auto_expand_volume")
			if old.(bool) {
				_, err := cfg.Client.Post(ctx, client.ApiPath.DatabasePostgresInstanceDisableAutoExpandVolume(cfg.ProjectID, d.Id()), nil, nil, nil)
				if err != nil {
					return diag.Errorf("Error disabling auto-expand volume for vnpaycloud_database_postgres_instance %s: %s", d.Id(), err)
				}
				if diags := waitForPostgresActive(ctx, d, cfg); diags != nil {
					return diags
				}
			}
		}
	}

	if d.HasChange("enable_tls") || d.HasChange("certificate_id") || d.HasChange("tls_mode") {
		if d.Get("enable_tls").(bool) {
			// If already enabled, disable first before re-enabling with new params
			old, _ := d.GetChange("enable_tls")
			if old.(bool) {
				_, err := cfg.Client.Post(ctx, client.ApiPath.DatabasePostgresInstanceDisableTls(cfg.ProjectID, d.Id()), nil, nil, nil)
				if err != nil {
					return diag.Errorf("Error disabling TLS for vnpaycloud_database_postgres_instance %s: %s", d.Id(), err)
				}
				if diags := waitForPostgresActive(ctx, d, cfg); diags != nil {
					return diags
				}
			}
			opts := dto.EnableTlsPostgresInstanceRequest{
				CertificateID: d.Get("certificate_id").(string),
				TlsMode:       d.Get("tls_mode").(string),
			}
			_, err := cfg.Client.Post(ctx, client.ApiPath.DatabasePostgresInstanceEnableTls(cfg.ProjectID, d.Id()), opts, nil, nil)
			if err != nil {
				return diag.Errorf("Error enabling TLS for vnpaycloud_database_postgres_instance %s: %s", d.Id(), err)
			}
		} else {
			_, err := cfg.Client.Post(ctx, client.ApiPath.DatabasePostgresInstanceDisableTls(cfg.ProjectID, d.Id()), nil, nil, nil)
			if err != nil {
				return diag.Errorf("Error disabling TLS for vnpaycloud_database_postgres_instance %s: %s", d.Id(), err)
			}
		}
		if diags := waitForPostgresActive(ctx, d, cfg); diags != nil {
			return diags
		}
	}

	if d.HasChange("enable_read_only_endpoint") {
		var err error
		if d.Get("enable_read_only_endpoint").(bool) {
			_, err = cfg.Client.Post(ctx, client.ApiPath.DatabasePostgresInstanceEnableReadOnlyEndpoint(cfg.ProjectID, d.Id()), nil, nil, nil)
		} else {
			_, err = cfg.Client.Post(ctx, client.ApiPath.DatabasePostgresInstanceDisableReadOnlyEndpoint(cfg.ProjectID, d.Id()), nil, nil, nil)
		}
		if err != nil {
			return diag.Errorf("Error updating read-only endpoint for vnpaycloud_database_postgres_instance %s: %s", d.Id(), err)
		}
		if diags := waitForPostgresActive(ctx, d, cfg); diags != nil {
			return diags
		}
	}

	return resourcePostgresInstanceRead(ctx, d, meta)
}

func resourcePostgresInstanceDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	cfg := meta.(*config.Config)

	_, err := cfg.Client.Delete(ctx, client.ApiPath.DatabasePostgresInstanceWithID(cfg.ProjectID, d.Id()), nil)
	if err != nil {
		return diag.FromErr(util.CheckDeleted(d, err, "vnpaycloud_database_postgres_instance"))
	}

	stateConf := &retry.StateChangeConf{
		Pending:    []string{"deleting", "active", "unknown"},
		Target:     []string{"deleted"},
		Refresh:    postgresInstanceStateRefreshFunc(ctx, cfg.Client, cfg.ProjectID, d.Id()),
		Timeout:    d.Timeout(schema.TimeoutDelete),
		Delay:      15 * time.Second,
		MinTimeout: 10 * time.Second,
	}

	_, err = stateConf.WaitForStateContext(ctx)
	if err != nil {
		return diag.Errorf("Error waiting for vnpaycloud_database_postgres_instance %s to delete: %s", d.Id(), err)
	}

	return nil
}

func waitForPostgresActive(ctx context.Context, d *schema.ResourceData, cfg *config.Config) diag.Diagnostics {
	stateConf := &retry.StateChangeConf{
		Pending:    []string{"creating", "scaling", "changing_flavor", "expanding_volume", "updating", "pending_update", "unknown"},
		Target:     []string{"active"},
		Refresh:    postgresInstanceStateRefreshFunc(ctx, cfg.Client, cfg.ProjectID, d.Id()),
		Timeout:    d.Timeout(schema.TimeoutUpdate),
		Delay:      15 * time.Second,
		MinTimeout: 10 * time.Second,
	}

	_, err := stateConf.WaitForStateContext(ctx)
	if err != nil {
		return diag.Errorf("Error waiting for vnpaycloud_database_postgres_instance %s to become active: %s", d.Id(), err)
	}
	return nil
}
