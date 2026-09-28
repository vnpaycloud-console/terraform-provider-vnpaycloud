package maasaccesskey

import (
	"context"
	"regexp"
	"terraform-provider-vnpaycloud/vnpaycloud/config"
	"terraform-provider-vnpaycloud/vnpaycloud/dto"
	"terraform-provider-vnpaycloud/vnpaycloud/helper/client"
	"terraform-provider-vnpaycloud/vnpaycloud/util"

	"github.com/hashicorp/terraform-plugin-log/tflog"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
)

var nameRegex = regexp.MustCompile(`^[a-zA-Z0-9-_. ]*$`)

func ResourceMaasAccessKey() *schema.Resource {
	return &schema.Resource{
		CreateContext: resourceMaasAccessKeyCreate,
		ReadContext:   resourceMaasAccessKeyRead,
		UpdateContext: resourceMaasAccessKeyUpdate,
		DeleteContext: resourceMaasAccessKeyDelete,
		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},
		Schema: map[string]*schema.Schema{
			"name": {
				Type:     schema.TypeString,
				Required: true,
				ValidateFunc: validation.All(
					validation.StringLenBetween(3, 255),
					validation.StringMatch(nameRegex, "must contain only letters, digits, spaces, hyphens, underscores and dots"),
				),
			},
			"description": {
				Type:     schema.TypeString,
				Optional: true,
				ValidateFunc: validation.All(
					validation.StringLenBetween(0, 255),
					validation.StringMatch(nameRegex, "must contain only letters, digits, spaces, hyphens, underscores and dots"),
				),
			},
			"log_permission": {
				Type:         schema.TypeString,
				Optional:     true,
				Default:      "none",
				ValidateFunc: validation.StringInSlice(permissions, false),
			},
			"metric_permission": {
				Type:         schema.TypeString,
				Optional:     true,
				Default:      "none",
				ValidateFunc: validation.StringInSlice(permissions, false),
			},
			"log_pipeline_id": {
				Type:     schema.TypeString,
				Optional: true,
			},
			"metric_pipeline_id": {
				Type:     schema.TypeString,
				Optional: true,
			},
			"log_label_matcher": {
				Type:     schema.TypeSet,
				Optional: true,
				MaxItems: 16,
				Elem:     labelMatcherWithValidation(),
			},
			"metric_label_matcher": {
				Type:     schema.TypeSet,
				Optional: true,
				MaxItems: 16,
				Elem:     labelMatcherWithValidation(),
			},
			"status": {
				Type:         schema.TypeString,
				Optional:     true,
				Default:      statusActive,
				ValidateFunc: validation.StringInSlice([]string{statusActive, statusInactive}, false),
			},
			"api_key": {
				Type:      schema.TypeString,
				Computed:  true,
				Sensitive: true,
			},
			"username": {
				Type:      schema.TypeString,
				Computed:  true,
				Sensitive: true,
			},
			"password": {
				Type:      schema.TypeString,
				Computed:  true,
				Sensitive: true,
			},
			"endpoints":  endpointsSchema(),
			"created_at": {Type: schema.TypeString, Computed: true},
		},
	}
}

func labelMatcherWithValidation() *schema.Resource {
	r := labelMatcherResource()
	r.Schema["name"].ValidateFunc = validation.StringMatch(
		regexp.MustCompile(`^[a-zA-Z_][a-zA-Z0-9_]*$`),
		"must be a valid label name",
	)
	r.Schema["op"].ValidateFunc = validation.StringInSlice(matcherOps, false)
	r.Schema["value"].ValidateFunc = validation.StringLenBetween(1, 2048)
	return r
}

func resourceMaasAccessKeyCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	cfg := meta.(*config.Config)

	createOpts := dto.CreateMaasAccessKeyRequest{
		Name:                d.Get("name").(string),
		Description:         d.Get("description").(string),
		LogPermission:       d.Get("log_permission").(string),
		MetricPermission:    d.Get("metric_permission").(string),
		LogPipelineID:       d.Get("log_pipeline_id").(string),
		MetricPipelineID:    d.Get("metric_pipeline_id").(string),
		LogLabelMatchers:    expandLabelMatchers(d.Get("log_label_matcher")),
		MetricLabelMatchers: expandLabelMatchers(d.Get("metric_label_matcher")),
	}

	tflog.Debug(ctx, "vnpaycloud_maas_access_key create options", map[string]interface{}{
		"name":              createOpts.Name,
		"log_permission":    createOpts.LogPermission,
		"metric_permission": createOpts.MetricPermission,
	})

	createResp := &dto.MaasAccessKeyResponse{}
	if _, err := cfg.Client.Post(ctx, client.ApiPath.MaasAccessKeys(cfg.ProjectID), createOpts, createResp, nil); err != nil {
		return diag.Errorf("Error creating vnpaycloud_maas_access_key: %s", err)
	}

	d.SetId(createResp.AccessKey.ID)

	// The credentials come back once, in this response, and are never readable again.
	d.Set("username", createResp.AccessKey.Username)
	d.Set("password", createResp.AccessKey.Password)

	if status := d.Get("status").(string); status == statusInactive {
		if err := setAccessKeyStatus(ctx, cfg, d.Id(), statusInactive); err != nil {
			return diag.FromErr(err)
		}
	}

	return resourceMaasAccessKeyRead(ctx, d, meta)
}

func resourceMaasAccessKeyRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	cfg := meta.(*config.Config)

	resp := &dto.MaasAccessKeyResponse{}
	if _, err := cfg.Client.Get(ctx, client.ApiPath.MaasAccessKeyWithID(cfg.ProjectID, d.Id()), resp, nil); err != nil {
		return diag.FromErr(util.CheckNotFound(d, err, "Error retrieving vnpaycloud_maas_access_key"))
	}

	setAccessKeyData(d, &resp.AccessKey)

	return nil
}

func resourceMaasAccessKeyUpdate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	cfg := meta.(*config.Config)

	fieldsChanged := d.HasChanges(
		"name", "description", "log_permission", "metric_permission",
		"log_pipeline_id", "metric_pipeline_id", "log_label_matcher", "metric_label_matcher",
	)

	// The backend refuses to update a disabled key, so re-enable before touching anything else and
	// disable only once the new values are in.
	if d.HasChange("status") && d.Get("status").(string) == statusActive {
		if err := setAccessKeyStatus(ctx, cfg, d.Id(), statusActive); err != nil {
			return diag.FromErr(err)
		}
	}

	if fieldsChanged {
		updateOpts := dto.UpdateMaasAccessKeyRequest{
			Name:                d.Get("name").(string),
			Description:         d.Get("description").(string),
			LogPermission:       d.Get("log_permission").(string),
			MetricPermission:    d.Get("metric_permission").(string),
			LogPipelineID:       d.Get("log_pipeline_id").(string),
			MetricPipelineID:    d.Get("metric_pipeline_id").(string),
			LogLabelMatchers:    expandLabelMatchers(d.Get("log_label_matcher")),
			MetricLabelMatchers: expandLabelMatchers(d.Get("metric_label_matcher")),
		}
		if _, err := cfg.Client.Put(ctx, client.ApiPath.MaasAccessKeyWithID(cfg.ProjectID, d.Id()), updateOpts, nil, nil); err != nil {
			return diag.Errorf("Error updating vnpaycloud_maas_access_key %s: %s", d.Id(), err)
		}
	}

	if d.HasChange("status") && d.Get("status").(string) == statusInactive {
		if err := setAccessKeyStatus(ctx, cfg, d.Id(), statusInactive); err != nil {
			return diag.FromErr(err)
		}
	}

	return resourceMaasAccessKeyRead(ctx, d, meta)
}

func resourceMaasAccessKeyDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	cfg := meta.(*config.Config)

	if _, err := cfg.Client.Delete(ctx, client.ApiPath.MaasAccessKeyWithID(cfg.ProjectID, d.Id()), nil); err != nil {
		return diag.FromErr(util.CheckDeleted(d, err, "Error deleting vnpaycloud_maas_access_key"))
	}

	return nil
}

func setAccessKeyStatus(ctx context.Context, cfg *config.Config, id, status string) error {
	opts := dto.UpdateMaasAccessKeyStatusRequest{Status: status}
	if _, err := cfg.Client.Put(ctx, client.ApiPath.MaasAccessKeyStatus(cfg.ProjectID, id), opts, nil, nil); err != nil {
		return err
	}
	return nil
}
